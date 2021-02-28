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
	"context"
	"fmt"
	"github.com/viltgroup/bucket-restore/internal/brestore/gcprestore"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	gcp_history "github.com/viltgroup/bucket-restore/internal/brestore/gcprestore/history"
	gcp_generations "github.com/viltgroup/bucket-restore/internal/brestore/gcprestore/history/generations"
)

func doRestoreGCP(keyfile string, bucketName string, path string, timestamp time.Time, quiet bool) error {

	client, _, err := gcprestore.GetStorageClientFromFile(keyfile)
	if err != nil {
		return fmt.Errorf("getting storage client for key file '%v': %w", keyfile, err)
	}

	bucket := client.Bucket(bucketName)

	var created, deleted, noAction int64
	var errors []error

	actions := gcp_history.FileActions{}

	listingStarted := time.Now()

	allGens, err := gcp_generations.OfPathByName(bucket, path)
	if err != nil {
		return fmt.Errorf("listing contents of bucket: %v", err)
	}

	listingElapsed := time.Since(listingStarted)

	decisionsStarted := time.Now()

	for _, fileGens := range allGens {
		fileGens.SortByCreatedDateAsc()
		desiredState, lastState := gcp_history.StateDiffAtTime(fileGens, timestamp)
		action := gcp_history.ActionForStateChange(lastState, desiredState)
		if action.Action != gcp_history.NO_ACTION {
			actions = append(actions, action)
		} else {
			noAction++
		}
	}

	decisionsElapsed := time.Since(decisionsStarted)

	actionsStarted := time.Now()

	ctx := context.Background()
	resChan := make(chan RunActionResultGCP, 1024)

	nActions := len(actions)
	parts := 1
	if nActions > 4 {
		parts = *maxConcurrencyFlag
	}

	go concurrentActionsGCP(ctx, bucket, actions, parts, resChan)

	i := 1
	for result := range resChan {
		if result.Err != nil {
			errors = append(errors, result.Err)
			fmt.Printf("[%d/%d] Error for %s '%s': %v\n",
				i, nActions, result.Action.String(), result.Action.Source.Name, result.Err)
		} else {
			switch result.Action.Action {
			case gcp_history.CREATE:
				if !quiet {
					fmt.Printf("[%d/%d] Created %s(#%d) from #%d\n",
						i,
						nActions,
						result.NewObj.Name,
						result.NewObj.Generation,
						result.Action.Source.Generation)
				}
				created++
			case gcp_history.DELETE:
				if !quiet {
					fmt.Printf("[%d/%d] Deleted %s(#%d)\n",
						i,
						nActions,
						result.Action.Source.Name,
						result.Action.Source.Generation)
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
	fmt.Printf("    %d errors\n", errors)
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

// RunActionResultGCP contains info about the execution of an action
type RunActionResultGCP struct {
	Action gcp_history.FileAction
	NewObj *storage.ObjectAttrs
	Err    error
}

func concurrentActionsGCP(ctx context.Context, bucket *storage.BucketHandle, actions gcp_history.FileActions, parts int, resultChan chan<- RunActionResultGCP) {

	var wg sync.WaitGroup

	chunks := divideActionsGCP(actions, parts)
	wg.Add(len(chunks))

	for _, chunk := range chunks {
		chunk := chunk
		go func() {
			runActionsGCP(ctx, bucket, chunk, resultChan)
			wg.Done()
		}()
	}

	wg.Wait()
	close(resultChan)
}

func divideActionsGCP(actions gcp_history.FileActions, parts int) []gcp_history.FileActions {
	var divided []gcp_history.FileActions

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

func runActionsGCP(ctx context.Context, bucket *storage.BucketHandle, actions gcp_history.FileActions, resultChan chan<- RunActionResultGCP) {

	for _, action := range actions {
		res := RunActionResultGCP{Action: action}

		switch action.Action {
		case gcp_history.CREATE:
			newObject, err := doCreateGCP(ctx, bucket, action)
			if err != nil {
				res.Err = fmt.Errorf("creating object '%s': %v", action.Source.Name, err)
			} else {
				res.NewObj = newObject
			}

		case gcp_history.DELETE:
			err := doDeleteGCP(ctx, bucket, action)
			if err != nil {
				res.Err = fmt.Errorf("deleting object '%s': %v", action.Source.Name, err)
			}
		default:
		}

		resultChan <- res
	}

}

func doDryRunExplainGCP(keyfile string, bucketName string, path string, time time.Time) error {
	client, _, err := gcprestore.GetStorageClientFromFile(keyfile)
	if err != nil {
		return fmt.Errorf("getting storage client for key file '%v': %w", keyfile, err)
	}

	bucket := client.Bucket(bucketName)

	actions := gcp_history.FileActions{}

	allGens, err := gcp_generations.OfPathByName(bucket, path)
	if err != nil {
		return fmt.Errorf("listing contents of bucket: %w", err)
	}

	for _, fileGens := range allGens {
		fileGens.SortByCreatedDateAsc()
		desiredState, lastState := gcp_history.StateDiffAtTime(fileGens, time)
		action := gcp_history.ActionForStateChange(lastState, desiredState)
		actions = append(actions, action)
		fmt.Printf(""+
			"%s: %s\n"+
			"  Current state: %s\n"+
			"  Restore state: %s\n",
			action.Source.Name,
			formatActionGCP(action),
			formatStateGCP(lastState),
			formatStateGCP(desiredState))
	}

	return nil
}

func doDryRunGCP(keyfile string, bucketName string, path string, time time.Time) error {

	client, _, err := gcprestore.GetStorageClientFromFile(keyfile)
	if err != nil {
		return fmt.Errorf("getting storage client for key file '%v': %w", keyfile, err)
	}

	bucket := client.Bucket(bucketName)

	var toCreate, toDelete, noAction int64

	allGens, err := gcp_generations.OfPathByName(bucket, path)
	if err != nil {
		return fmt.Errorf("listing contents of bucket: %w", err)
	}

	for _, fileGens := range allGens {
		fileGens.SortByCreatedDateAsc()
		desiredState, lastState := gcp_history.StateDiffAtTime(fileGens, time)
		action := gcp_history.ActionForStateChange(lastState, desiredState)
		switch action.Action {
		case gcp_history.CREATE:
			toCreate++
		case gcp_history.DELETE:
			toDelete++
		case gcp_history.NO_ACTION:
			noAction++
		}
	}

	fmt.Printf("To create: %d objects\n", toCreate)
	fmt.Printf("To delete %d objects\n", toDelete)
	fmt.Printf("No action: %d objects\n", noAction)

	return nil
}

func doCreateGCP(ctx context.Context, bucket *storage.BucketHandle, action gcp_history.FileAction) (*storage.ObjectAttrs, error) {
	toObject := bucket.Object(action.Source.Name)
	fromObject := toObject.Generation(action.Source.Generation)

	if action.GenerationPreCondition != 0 {
		toObject = toObject.If(storage.Conditions{GenerationMatch: action.GenerationPreCondition})
	}

	copier := toObject.CopierFrom(fromObject)
	return copier.Run(ctx)
}

func doDeleteGCP(ctx context.Context, bucket *storage.BucketHandle, action gcp_history.FileAction) error {
	obj := bucket.Object(action.Source.Name).If(storage.Conditions{GenerationMatch: action.GenerationPreCondition})
	return obj.Delete(ctx)
}

func formatStateGCP(state gcp_history.PathState) string {
	switch state.PathStatus {
	case gcp_history.NOT_EXISTENT:
		return "Not Existent"
	case gcp_history.EXISTS:
		return fmt.Sprintf("Exists at generation #%d, md5: %x", state.Generation, state.MD5)
	case gcp_history.DELETED:
		return fmt.Sprintf("Deleted on generation #%d, md5: %x", state.Generation, state.MD5)
	default:
		return "Unknown Status"
	}
}

func formatActionGCP(action gcp_history.FileAction) string {
	switch action.Action {
	case gcp_history.DELETE:
		return "Delete"
	case gcp_history.CREATE:
		return fmt.Sprintf("Create from #%d", action.Source.Generation)
	case gcp_history.NO_ACTION:
		return "No Action"
	default:
		return "Unknown Status"
	}
}
