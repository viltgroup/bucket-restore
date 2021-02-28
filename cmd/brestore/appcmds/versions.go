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
	"github.com/spf13/cobra"
	"github.com/viltgroup/bucket-restore/internal/brestore"
)

func init() {
	rootCmd.AddCommand(versionsCmd)
}

var versionsExamples = "" +
	"  Show all versions for all objects in a bucket:\n" +
	"    brestore versions --bucket s3://mybucket\n\n" +
	"  Show all versions for a specific object or objects in a path:\n" +
	"    brestore versions --bucket s3://mybucket/path/to/dir_or_file"

var versionsCmd = &cobra.Command{
	Use:     "versions",
	Aliases: []string{"gens", "history", "generations"},
	Short:   "Shows the versions/generations of objects",
	Long: "" +
		"Description:\n" +
		"  Shows the versions/generations of objects in the bucket.",
	RunE:    versionsEntryPoint,
	Example: versionsExamples,
}

func versionsEntryPoint(cmd *cobra.Command, args []string) error {
	if *sourceBucketFlag == "" {
		return fmt.Errorf("No bucket specified. Specify the bucket to which this action should be applied with -b <bucket_url>.")
	}

	binfo, err := brestore.ParseBucketURL(*sourceBucketFlag)
	if err != nil {
		return fmt.Errorf("could not parse bucket information from url: %v", err)
	}

	return doVersions(binfo)
}

func doVersions(binfo brestore.BucketURLInfo) error {
	switch binfo.Type {
	case "s3":
		return doVersionsAWS(*profileFlag, binfo.BucketName, binfo.Prefix)
	case "gs":
		return doVersionsGCP(*keyFileFlag, binfo.BucketName, binfo.Prefix)
	}
	return nil
}
