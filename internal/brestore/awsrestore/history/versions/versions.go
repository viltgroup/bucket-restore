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

package versions

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"sort"
)

// Versions represents a collection of versions
type Versions []Version

// SortByLastModifiedAsc sorts the versions by ascending order of their creation date.
func (gs Versions) SortByLastModifiedAsc() {
	ascOrder := func(i, j int) bool {
		return gs[i].LastModified.Before(gs[j].LastModified)
	}
	gs.SortIfNeeded(ascOrder)
}

// SortByLastModifiedDesc sorts the versions by descending order of their creation date.
func (gs Versions) SortByLastModifiedDesc() {
	descOrder := func(i, j int) bool {
		return gs[i].LastModified.After(gs[j].LastModified)
	}
	gs.SortIfNeeded(descOrder)
}

// SortIfNeeded sorts the collection of versions by the given sortFunc,
// but checks first if the list isn't already sorted by the same sortFunc.
// If the collection is already sorted, nothing is done.
func (gs Versions) SortIfNeeded(sortFunc func(i, j int) bool) {
	if !sort.SliceIsSorted(gs, sortFunc) {
		sort.Slice(gs, sortFunc)
	}
}

// OfBucket returns a collection of all versions of all objects in a bucket.
func OfBucket(client *s3.S3, bucketName string) (Versions, error) {
	return OfPath(client, bucketName, "")
}

// OfPath returns a collection of versions of objects in a bucket that have the given path prefix.
// If an empty string is given as a path prefix, all versions of all objects in the bucket will be returned.
func OfPath(client *s3.S3, bucketName string, path string) (Versions, error) {
	var res Versions

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(path),
	}

	err := client.ListObjectVersionsPages(input,
		func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
			for _, v := range page.Versions {
				res = append(res, FromAWSVersion(v))
			}

			for _, dm := range page.DeleteMarkers {
				res = append(res, FromAWSDeleteMarker(dm))
			}
			return !lastPage
		})

	if err != nil {
		return res, fmt.Errorf("%w", err)
	}

	return res, nil
}

// OfPathByName returns a collection of versions of objects in a bucket that have the given path prefix, indexed
// by path name as a map of object paths to its generation list.
// If an empty string is given as a path prefix, all versions of all objects in the bucket
// will be returned.
func OfPathByName(client *s3.S3, bucketName string, path string) (map[string]Versions, error) {
	res := make(map[string]Versions)

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(path),
	}

	err := client.ListObjectVersionsPages(input,
		func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
			for _, v := range page.Versions {
				v := FromAWSVersion(v)
				if _, ok := res[v.Key]; !ok {
					res[v.Key] = Versions{}
				}
				res[v.Key] = append(res[v.Key], v)
			}

			for _, dm := range page.DeleteMarkers {
				v := FromAWSDeleteMarker(dm)
				if _, ok := res[v.Key]; !ok {
					res[v.Key] = Versions{}
				}
				res[v.Key] = append(res[v.Key], v)
			}
			return !lastPage
		})

	if err != nil {
		return res, fmt.Errorf("%w", err)
	}

	return res, nil
}
