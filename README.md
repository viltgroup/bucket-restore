# `brestore` - Bucket Restore for AWS and GCP

[![Go Report Card](https://goreportcard.com/badge/github.com/viltgroup/bucket-restore)](https://goreportcard.com/report/github.com/viltgroup/bucket-restore)
[![GoDoc](https://godoc.org/github.com/viltgroup/bucket-restore?status.svg)](https://godoc.org/github.com/viltgroup/bucket-restore) 

A program that does point in time restores for buckets in Amazon Web Services (AWS) S3 and Google Cloud Platform (GCP) Storage.

## Features

* Rollback objects in a bucket to a previous state.
* Rollback all objects in a bucket or only specific ones.
* Dry runs - preview changes before committing to a rollback
* Check versions/generations of objects in a bucket
* Works with multiple objects at the same time for fast restores
* Doesn't delete object history - it's always possible to undo a rollback

## Download

`brestore` is distributed as a single binary that can be downloaded in the [releases section](https://github.com/viltgroup/bucket-restore/releases) of this repository.

* [Brestore releases download](https://github.com/viltgroup/bucket-restore/releases)

## Install using Go

If you have Go 1.16+ installed, you can install `brestore` on your machine with `go get`:

    go get github.com/viltgroup/bucket-restore/cmd/brestore

This will install `brestore` on your machine. The location where Go installs packages depends on how your environment is set up, but this folder is typically in `$HOME/go/bin`.

## Examples

### Rollback actions

* Rollback all objects in the AWS S3 bucket 'mybucket' to a specific point in time:

  `brestore rollback --bucket s3://mybucket --time "February 21, 2021, 23:00:00 (UTC+01:00)"`
    
* The same as the previous command, but for gcp storage:

  `brestore rollback --bucket gs://mybucket --time "February 21, 2021, 23:00:00 (UTC+01:00)"`

* Rollback a specific object or all objects under a path inside the bucket:

  `brestore rollback --bucket gs://mybucket/path/to/dir_or_file --time "February 21, 2021, 23:00:00 (UTC+01:00)"`

* To perform a dry run, add the flag `--dry-run-explain` or `--dry-run` to the rollback command:

  `brestore versions --bucket gs://mybucket/ --time "February 21, 2021, 23:00:00 (UTC+01:00)" --dry-run-explain`

### Check object versions/generations:

* Show all versions for all objects in a bucket:

  `brestore versions --bucket s3://mybucket`

* Show all versions for a specific object or objects in a path:

  `brestore versions --bucket s3://mybucket/path/to/dir_or_file`

## Usage

`brestore <command> [flags]`

**Available commands**

* `help` - Help about any command
* `rollback` - Rollback objects in a bucket to a specific point in time. Aliases: `restore`.
* `version` - Shows the current version of brestore
* `versions` - Shows the versions/generations of objects. Aliases: `gens`, `history`, `generations`.

**Global Flags**

* `-p, --aws-profile string` - name of the AWS profile to use to perform requests.

* `-b, --bucket string` - the URI to the bucket to which rollback/listing actions shouldbe applied.
* `-k, --gcp-key-file string` - path to a JSON key file of a GCP Service Account
* `-h, --help` - help for brestore
* `-t, --time string` - the point in time where to restore to.

**Rollback flags**

* `-d, --dry-run` - summary of the changes that will be made if the rollback is run.
* `-e, --dry-run-explain` -        same as '--dry-run' but shows additional information for each object about the current state and the state of the object in the point in time given to the --time flag.
* `-h, --help` - help for rollback
* `-c, --max-concurrency int` - maximum number of rollback actions that can run concurrently.
* `-q, --quiet` - show less output.

## Authentication

To perform actions, `brestore` needs make authenticated requests to AWS/GCP. This explains how authentication credentials can be provided to `brestore`.

### Amazon Web Services (AWS)

`brestore` will try to fetch the credential information from the [AWS shared credentials file](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/create-shared-credentials-file.html).
Typically, this file is located at `~/.aws/credentials`. You can put credentials in this file either manually or with [AWS Cli](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html).

This file can include different credential profiles, including a default one. `brestore` will try to use the default profile if no profile is given. You can tell `brestore` to use another profile inside the shared credentials file by adding `--aws-profile <profile_name>` to your command. The profile used must have permissions to read bucket information and to create/delete/list objects.

### Google Cloud Platform (GCP)

`brestore` will try to fetch credentials from a JSON File containing a key of a Service Account. The key file can be specified by passing `--gcp-key-file <path_to_key_file>` to `brestore` or by setting the environment variable `GOOGLE_APPLICATION_CREDENTIALS` with the path to the key file. If both the flag and the envrionemnt variables are set, the path set in the flag will take precedence and will be used. The Service Account to which the key belongs must have permissions to read bucket information and to create/delete/list objects.

## Time formats

The `--time` flag allows a point in time to be specified in several formats. Below are examples of the date 'January 02, 2006, 15:04:05 (UTC-07:00)' in all formats accepted by `brestore`:

<details>
  <summary>Supported time formats</summary>

**Formats similar to the ones in the AWS UI:**

    January 02, 2006, 15:04:05 (UTC-07:00)
    Jan 02, 2006, 15:04:05 (UTC-07:00)
    January 02, 2006, 15:04:05 (-07:00)
    Jan 02, 2006, 15:04:05 (-07:00)

**Formats similar to the ones in the GCP UI:**

    Jan 02, 2006, 03:04:05 PM -07:00
    Jan 02, 2006, 03:04:05 PM MST
    January 02, 2006, 03:04:05 PM -07:00
    January 02, 2006, 03:04:05 PM MST
	
**Other allowed formats:**

    02-01-2006 15:04:05 -07:00
    02-01-2006 15:04:05 MST
    02-01-2006 03:04:05 PM -07:00
    02-01-2006 03:04:05 PM MST
    2006-01-02 15:04:05 -07:00
    2006-01-02 03:04:05 PM MST
    2006-01-02 03:04:05 PM -07:00
    2006-01-02 03:04:05 PM MST
    Mon, 02 Jan 2006 15:04:05 MST
    2006-01-02T15:04:05Z07:00
  
</details>

## License

Apache License 2.0, see [LICENSE](https://github.com/viltgroup/bucket-restore/blob/master/LICENSE).