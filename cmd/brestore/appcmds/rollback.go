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
	"github.com/viltgroup/bucket-restore/internal/brestore"
	"time"

	"github.com/spf13/cobra"
)

var (
	dryRunExplainFlag  *bool
	dryRunFlag         *bool
	quietFlag          *bool
	maxConcurrencyFlag *int
)

var rollbackExamples = "" +
	"  Rollback all objects in the AWS S3 bucket 'mybucket' to the specific point in time\n" +
	"    brestore rollback --bucket s3://mybucket --time \"February 21, 2021, 23:00:00 (UTC+01:00)\"\n\n" +
	"  The same as the previous command, but for gcp storage:\n" +
	"    brestore rollback --bucket gs://mybucket --time \"February 21, 2021, 23:00:00 (UTC+01:00)\"\n\n" +
	"  Rollback a specific object or all objects under a path inside the bucket:\n" +
	"    brestore rollback --bucket gs://mybucket/path/to/dir_or_file --time \"February 21, 2021, 23:00:00 (UTC+01:00)\"\n\n" +
	"  To perform a dry run, add the flag --dry-run-explain or --dry-run to the rollback command:\n" +
	"    brestore versions --bucket gs://mybucket/ --time \"February 21, 2021, 23:00:00 (UTC+01:00)\" --dry-run-explain"

func init() {

	dryRunExplainFlag = rollbackCmd.PersistentFlags().BoolP("dry-run-explain", "e", false,
		"same as '--dry-run' but shows additional information for each object about the current state and the "+
			"state of the object in the point in time given to the --time flag. ")
	dryRunFlag = rollbackCmd.PersistentFlags().BoolP("dry-run", "d", false,
		"if present, lists a summary of the changes that will be made if the rollback is run. "+
			"No changes to the bucket are actually performed. If instead of a summary you want detailed information "+
			"about the operation to apply to each object, use the --dry-run-explain flag.")
	quietFlag = rollbackCmd.PersistentFlags().BoolP("quiet", "q", false,
		"show less output.")
	maxConcurrencyFlag = rollbackCmd.PersistentFlags().IntP("max-concurrency", "c", 32,
		"controls the maximum number of rollback actions that can run concurrently.")

	rootCmd.AddCommand(rollbackCmd)
}

var rollbackCmd = &cobra.Command{
	Use:     "rollback",
	Aliases: []string{"restore"},
	Short:   "Rollback objects in a bucket to a specific point in time",
	Long: "" +
		"Description:\n" +
		"  Rollback objects in a bucket to a specific point in time.\n\n",
	Example:      rollbackExamples,
	RunE:         rollbackEntryPoint,
	SilenceUsage: true,
}

func rollbackEntryPoint(cmd *cobra.Command, args []string) error {

	if *sourceBucketFlag == "" {
		return fmt.Errorf("No bucket specified. Specify the bucket to which this action should be applied with -b <bucket_url>.")
	}

	if *timestampFlag == "" {
		return fmt.Errorf("No timestamp specified. Specify a timestamp for this action with -t <timestamp>.\n")
	}

	binfo, err := brestore.ParseBucketURL(*sourceBucketFlag)
	if err != nil {
		return fmt.Errorf("could not parse bucket information from url: %v", err)
	}

	ts, err := brestore.ParseTimestamp(*timestampFlag)
	if err != nil {
		return fmt.Errorf("could not parse timestamp: %v", err)
	}

	fmt.Printf("Restoring objects inside path '%v' at bucket '%s':\n"+
		"             Restore time: %v \n"+
		"    Restore time (in UTC): %v\n\n", binfo.Prefix, binfo.BucketName, ts, ts.UTC())

	if *dryRunExplainFlag {
		err = doDryRunExplain(binfo, ts)
	} else if *dryRunFlag {
		err = doDryRun(binfo, ts)
	} else {
		err = doRestore(binfo, ts)
	}

	if err != nil {
		return fmt.Errorf("error performing rollback command: %v", err)
	}

	return nil
}

func doDryRunExplain(binfo brestore.BucketURLInfo, ts time.Time) error {
	fmt.Printf("" +
		"Performing a dry-run explain. In this dry-run, the action for each file will be shown " +
		"along with details about the current state of the file and the desired state.\n" +
		"To perform a dry-run with less information, use the flag '--dry-run'.\n\n")
	switch binfo.Type {
	case "s3":
		return doDryRunExplainAWS(*profileFlag, binfo.BucketName, binfo.Prefix, ts)
	case "gs":
		return doDryRunExplainGCP(*keyFileFlag, binfo.BucketName, binfo.Prefix, ts)
	}
	return nil
}

func doDryRun(binfo brestore.BucketURLInfo, ts time.Time) error {
	fmt.Printf("" +
		"Performing a dry-run.\n" +
		"To see a dry-run with more details, use the flag '--dry-run-explain'\n\n")
	switch binfo.Type {
	case "s3":
		return doDryRunAWS(*profileFlag, binfo.BucketName, binfo.Prefix, ts)
	case "gs":
		return doDryRunGCP(*keyFileFlag, binfo.BucketName, binfo.Prefix, ts)
	}
	return nil
}

func doRestore(binfo brestore.BucketURLInfo, ts time.Time) error {
	switch binfo.Type {
	case "s3":
		return doRestoreAWS(*profileFlag, binfo.BucketName, binfo.Prefix, ts, *quietFlag)
	case "gs":
		return doRestoreGCP(*keyFileFlag, binfo.BucketName, binfo.Prefix, ts, *quietFlag)
	}
	return nil
}
