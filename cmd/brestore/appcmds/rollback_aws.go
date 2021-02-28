// Copyright 2021 VILT Group
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package appcmds

import (
	"fmt"
	"github.com/viltgroup/bucket-restore/internal/brestore/awsrestore"
	"sync"
	"time"

	"github.com/viltgroup/bucket-restore/internal/brestore/awsrestore/history"
	"github.com/viltgroup/bucket-restore/internal/brestore/awsrestore/history/versions"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// FiveGibibytes is the number of bytes in 5 Gibibytes.
//This value is important since files larger than this value need a special copy.
const FiveGibibytes = 1024 * 1024 * 1024 * 5

func doRestoreAWS(profile string, bucketName string, path string, timestamp time.Time, quiet bool) error {
	client, err := awsrestore.GetS3Client(profile)
	if err != nil {
		return fmt.Errorf("unable to get s3 client for profile '%v': %w", profile, err)
	}

	var created, deleted, noAction uint64

	actions := history.FileActions{}

	listingStarted := time.Now()

	allGens, err := versions.OfPathByName(client, bucketName, path)
	if err != nil {
		return fmt.Errorf("listing contents of client: %v", err)
	}

	listingElapsed := time.Since(listingStarted)

	decisionsStarted := time.Now()

	for _, fileGens := range allGens {
		fileGens.SortByLastModifiedAsc()
		desiredState, lastState := history.StateDiffAtTime(fileGens, timestamp)
		action := history.ActionForStateChange(lastState, desiredState)
		if action.Action != history.NO_ACTION {
			actions = append(actions, action)
		} else {
			noAction++
		}
	}

	decisionsElapsed := time.Since(decisionsStarted)

	actionsStarted := time.Now()

	resChan := make(chan RunActionResultAWS, 1024)

	nActions := len(actions)
	parts := 1
	if nActions > 4 {
		parts = *maxConcurrencyFlag
	}

	go concurrentActionsAWS(client, bucketName, actions, parts, resChan)

	i := 1
	var errors []error
	for result := range resChan {
		if result.Err != nil {
			errors = append(errors, result.Err)
			fmt.Printf("[%d/%d] Error for %s '%s': %v\n",
				i, nActions, result.Action.String(), result.Action.Source.Key, result.Err)
		} else {
			switch result.Action.Action {
			case history.CREATE:
				if !quiet {
					fmt.Printf("[%d/%d] Created %s from version #%s\n",
						i,
						nActions,
						result.Action.Source.Key,
						result.Action.Source.Version)
				}
				created++
			case history.DELETE:
				if !quiet {
					fmt.Printf("[%d/%d] Deleted %s(#%s)\n",
						i,
						nActions,
						result.Action.Source.Key,
						result.Action.Source.Version)
				}
				deleted++
			default:
			}
		}

		i++
	}

	actionsElapsed := time.Since(actionsStarted)

	fmt.Printf("\n")
	fmt.Printf("Bucket restored to %v:\n", timestamp)
	fmt.Printf("    %d objects created\n", created)
	fmt.Printf("    %d objects deleted\n", deleted)
	fmt.Printf("    %d objects did not need any action\n", noAction)
	fmt.Printf("    %d errors\n", len(errors))
	fmt.Printf(""+
		"Elapsed time: %v\n"+
		"    Retrieving object info: %v\n"+
		"    Action decision: %v\n"+
		"    Action execution: %v\n",
		listingElapsed+decisionsElapsed+actionsElapsed,
		listingElapsed, decisionsElapsed, actionsElapsed)

	if len(errors) > 0 {
		err = saveErrorsToFile("errors.log", errors)
		if err != nil {
			return fmt.Errorf("writing errors to 'error.log': %w", err)
		}
		fmt.Printf("" +
			"There were errors running the restore command.\n" +
			"A file 'errors.log' was created with the error details\n")
	}

	return nil
}

// RunActionResultAWS contains info about the execution of an action
type RunActionResultAWS struct {
	Action history.FileAction
	NewObj CopyResult
	Err    error
}

func concurrentActionsAWS(
	client *s3.S3,
	bucketName string,
	actions history.FileActions,
	parts int,
	resultChan chan<- RunActionResultAWS) {

	var wg sync.WaitGroup

	chunks := divideActions(actions, parts)
	wg.Add(len(chunks))

	for _, chunk := range chunks {
		chunk := chunk
		go func() {
			runActions(client, bucketName, chunk, resultChan)
			wg.Done()
		}()
	}

	wg.Wait()
	close(resultChan)
}
func divideActions(actions history.FileActions, parts int) []history.FileActions {
	var divided []history.FileActions

	nActions := len(actions)
	chunkSize := (nActions + parts - 1) / parts

	for i := 0; i < nActions; i += chunkSize {
		end := i + chunkSize

		if end > nActions {
			end = nActions
		}

		divided = append(divided, actions[i:end])
	}
	return divided
}

func runActions(
	bucket *s3.S3,
	bucketName string,
	actions history.FileActions,
	resultChan chan<- RunActionResultAWS) {

	for _, action := range actions {
		res := RunActionResultAWS{Action: action}

		switch action.Action {
		case history.CREATE:
			newObject, err := doCreate(bucket, bucketName, action)
			if err != nil {
				res.Err = fmt.Errorf("creating object '%s': %v", action.Source.Key, err)
			} else {
				res.NewObj = newObject
			}

		case history.DELETE:
			err := doDelete(bucket, bucketName, action)
			if err != nil {
				res.Err = fmt.Errorf("deleting object '%s': %v", action.Source.Key, err)
			}
		default:
		}

		resultChan <- res
	}

}

func doDryRunExplainAWS(profile string, bucketName string, path string, time time.Time) error {

	client, err := awsrestore.GetS3Client(profile)
	if err != nil {
		return fmt.Errorf("unable to get s3 client for profile '%v': %w", profile, err)
	}

	actions := history.FileActions{}

	allGens, err := versions.OfPathByName(client, bucketName, path)
	if err != nil {
		return fmt.Errorf("listing contents of bucket: %w", err)
	}

	for _, fileGens := range allGens {
		fileGens.SortByLastModifiedAsc()
		desiredState, lastState := history.StateDiffAtTime(fileGens, time)
		action := history.ActionForStateChange(lastState, desiredState)
		actions = append(actions, action)
		fmt.Printf(""+
			"%s: %s\n"+
			"  Current state: %s\n"+
			"  Restore state: %s\n",
			action.Source.Key,
			formatAction(action),
			formatState(lastState),
			formatState(desiredState))
	}

	return nil
}

func doDryRunAWS(profile string, bucketName string, path string, time time.Time) error {
	client, err := awsrestore.GetS3Client(profile)
	if err != nil {
		return fmt.Errorf("unable to get s3 client for profile '%v': %w", profile, err)
	}

	var toCreate, toDelete, noAction int64

	allGens, err := versions.OfPathByName(client, bucketName, path)
	if err != nil {
		return fmt.Errorf("listing contents of bucket: %w", err)
	}

	for _, fileGens := range allGens {
		fileGens.SortByLastModifiedAsc()
		desiredState, lastState := history.StateDiffAtTime(fileGens, time)
		action := history.ActionForStateChange(lastState, desiredState)
		switch action.Action {
		case history.CREATE:
			toCreate++
		case history.DELETE:
			toDelete++
		case history.NO_ACTION:
			noAction++
		}
	}

	fmt.Printf("To create: %d objects\n", toCreate)
	fmt.Printf("To delete %d objects\n", toDelete)
	fmt.Printf("No action: %d objects\n", noAction)

	return nil
}

// CopyResult is the result of a copy action
type CopyResult struct {
	Key       string
	VersionId string
	ETag      string
}

func doCreate(client *s3.S3, bucketName string, action history.FileAction) (CopyResult, error) {

	var copy *s3.CopyObjectOutput
	var res CopyResult
	var err error

	var matchUnmodified *time.Time

	if !action.UnmodifiedPreCondition.IsZero() {
		matchUnmodified = &action.UnmodifiedPreCondition
	}
	if action.Source.Size < FiveGibibytes {
		copy, err = client.CopyObject(&s3.CopyObjectInput{
			Bucket:                      aws.String(bucketName),
			Key:                         aws.String(action.Source.Key),
			CopySource:                  aws.String(action.Source.ToSourceURL(bucketName)),
			CopySourceIfUnmodifiedSince: matchUnmodified,
		})
		if err != nil {
			return res, fmt.Errorf("copying object: %w", err)
		}
		res = CopyResult{
			Key:       action.Source.Key,
			VersionId: *copy.VersionId,
			ETag:      *copy.CopyObjectResult.ETag,
		}
	} else {
		copier := s3manager.NewCopierWithClient(client)
		copy, err := copier.Copy(&s3manager.CopyInput{
			Bucket:     aws.String(bucketName),
			Key:        aws.String(action.Source.Key),
			CopySource: aws.String(action.Source.ToSourceURL(bucketName)),
		})
		if err != nil {
			return res, fmt.Errorf("copying object: %w", err)
		}
		res = CopyResult{
			Key:       action.Source.Key,
			VersionId: *copy.VersionId,
			ETag:      *copy.ETag,
		}
	}

	if err != nil {
		return res, fmt.Errorf("copying object: %w", err)
	}

	return res, nil
}

func doDelete(client *s3.S3, bucketName string, action history.FileAction) error {
	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(action.Source.Key),
	})

	return err
}

func formatState(state history.PathState) string {
	switch state.PathStatus {
	case history.NOT_EXISTENT:
		return "Not Existent"
	case history.EXISTS:
		return fmt.Sprintf("Exists at version %s ETag: %s", state.Version.ID, state.ETag)
	case history.DELETED:
		return fmt.Sprintf("Deleted on version %s", state.Version.ID)
	default:
		return "Unknown Status"
	}
}

func formatAction(action history.FileAction) string {
	switch action.Action {
	case history.DELETE:
		return "Delete"
	case history.CREATE:
		return fmt.Sprintf("Create from version %s", action.Source.Version)
	case history.NO_ACTION:
		return "No Action"
	default:
		return "Unknown Status"
	}
}
