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
	"github.com/spf13/cobra"
	"os"
)

var (
	keyFileFlag      *string
	profileFlag      *string
	cpuProfileFlag   *string
	sourceBucketFlag *string
	timestampFlag    *string
)

func init() {
	timestampFlag = rootCmd.PersistentFlags().StringP("time", "t", "",
		"the point in time where to restore to. "+
			"To see all allowed formats run 'brestore -h'. "+
			"e.g: --time \"January 02, 2006, 15:04:05 (UTC-07:00)\".")
	sourceBucketFlag = rootCmd.PersistentFlags().StringP("bucket", "b", "",
		"the URI to the bucket to which rollback/listing actions should be applied. "+
			"Optionally can include a path to a file or directory, in which case, the actions performed "+
			"by brestore will apply only to objects inside the provided path. "+
			"AWS buckets URI's should start with 's3://' and GCP URI's should start with 'gs://'. "+
			"e.g: -b \"s3://mybucket\", -b \"s3://mybucket/path/to/file_or_directory\"")
	keyFileFlag = rootCmd.PersistentFlags().StringP("gcp-key-file", "k", "",
		"path to a JSON key file of a Service Account with permissions to list/create/delete objects in the "+
			"bucket/path given to the --bucket flag. This path can also be given by setting the path to the json file "+
			"in the environment variable GOOGLE_APPLICATION_CREDENTIALS. If the environment variable already contains the "+
			"path to the key file desired to perform the requests, this flag can be omitted. If the environment variable is set"+
			"and a file is also passed with this flag, the file given via the flag will take precedence. "+
			"For more info about authentication, run 'brestore -h'. "+
			"e.g: --gcp-key-file \"~/mysecretsfolder/dev-sa.json\"")
	profileFlag = rootCmd.PersistentFlags().StringP("aws-profile", "p", "",
		"name of the AWS profile to use to perform requests. This profile should be present in the AWS shared credentials "+
			"file, usually placed in ~/.aws/credentials. If no value is specified for this flag, brestore will attempt to use"+
			"the default profile inside the shared credentials folder. "+
			"For more info about authentication, run 'brestore -h'. "+
			"e.g: --aws-profile \"production\"")
}

var rootCmd = &cobra.Command{
	Use:   "brestore",
	Short: "Point in time restores for buckets in Amazon Web Services (AWS) S3 and Google Cloud Platform (GCP).",
	Long: "" +
		"BUCKET RESTORE - brestore\n\n" +
		"SYNOPSIS\n\n" +
		"A program that does point in time restores for buckets in GCP Storage and AWS S3.\n\n" +
		"brestore allows you to check the versions of the objects in your buckets and restore " +
		"them to a previous state. " +
		"brestore can rollback all objects in a bucket or only specific objects in a path.\n\n" +
		"The rollback actions happen in-place in the bucket and no object history is deleted, only appended. " +
		"This means it's always possible to undo a rollback if the bucket has object versioning activated. " +
		"brestore will also perform actions on buckets with no object versioning activated, but no deleted " +
		"objects can be recovered by a rollback if object versioning is not activated. \n\n" +
		"AUTHENTICATION\n\n" +
		"To perform actions, brestore needs requests to AWS or GCP to be authenticated. This section describes " +
		"how you can provide authentication credentials to brestore.\n\n" +
		"   Amazon Web Services (AWS)\n\n" +
		"   brestore will try to fetch the credential information from the AWS shared credentials file, " +
		"typically located at ~/.aws/credentials. You can put credentials in this file either manually or with AWS Cli. " +
		"This file can include different credential profiles including a default one. brestore will try to use " +
		"the default profile if no profile is given. You can tell brestore to use another profile inside the shared" +
		"credentials file by adding --aws-profile <profile_name> to your command. The profile used must have permissions to read " +
		"bucket information and to create/delete/list objects.\n\n" +
		"   Google Cloud Platform (GCP)\n\n" +
		"   brestore will try to fetch credentials from a JSON File containing a key of a Service Account. The key file " +
		"can be specified with the --gcp-key-file <path_to_key_file> or by setting the environment variable GOOGLE_APPLICATION_CREDENTIALS " +
		"with the path to the key file. If the environment variable is set and the flag is set, the path given by the flag " +
		"will take precedence. The Service Account to which the key belongs must have permissions to read bucket " +
		"information and to create/delete/list objects.\n\n" +
		"TIME FORMATS\n\n" +
		"  brestore accepts several time formats when using the '--time <point_in_time>' flag. Below are examples " +
		"of the date 'January 02, 2006, 15:04:05 (UTC-07:00)' in all formats accepted by brestore:\n\n" +
		"  Formats similar to the ones in the AWS UI:\n\n" +
		"  January 02, 2006, 15:04:05 (UTC-07:00)\n" +
		"  Jan 02, 2006, 15:04:05 (UTC-07:00)\n" +
		"  January 02, 2006, 15:04:05 (-07:00)\n" +
		"  Jan 02, 2006, 15:04:05 (-07:00)\n\n" +
		"  Formats similar to the ones in the GCP UI:\n\n" +
		"  Jan 02, 2006, 03:04:05 PM -07:00\n" +
		"  Jan 02, 2006, 03:04:05 PM MST\n" +
		"  January 02, 2006, 03:04:05 PM -07:00\n" +
		"  January 02, 2006, 03:04:05 PM MST\n\n" +
		"  Other allowed formats:\n\n" +
		"  02-01-2006 15:04:05 -07:00\n" +
		"  02-01-2006 15:04:05 MST\n" +
		"  02-01-2006 03:04:05 PM -07:00\n" +
		"  02-01-2006 03:04:05 PM MST\n" +
		"  2006-01-02 15:04:05 -07:00\n" +
		"  2006-01-02 03:04:05 PM MST\n" +
		"  2006-01-02 03:04:05 PM -07:00\n" +
		"  2006-01-02 03:04:05 PM MST\n" +
		"  Mon, 02 Jan 2006 15:04:05 MST\n" +
		"  2006-01-02T15:04:05Z07:00\n",
	Example: "\n" +
		"Rollback actions:\n\n" +
		rollbackExamples + "\n\n" +
		"Check object versions/generations:\n\n" +
		versionsExamples,
	RunE: rootEntryPoint,
}

// Execute executes the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootEntryPoint(cmd *cobra.Command, args []string) error {
	cmd.Help()
	return nil
}
