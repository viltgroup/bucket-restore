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

package generations

import (
	"context"
	"fmt"
	"sort"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// Generations represents a collection of generations
type Generations []Generation

// SortByCreatedDateAsc sorts the generations by ascending order of their creation date.
func (gs Generations) SortByCreatedDateAsc() {
	ascOrder := func(i, j int) bool {
		return gs[i].Created.Before(gs[j].Created)
	}
	gs.SortIfNeeded(ascOrder)
}

// SortByCreatedDateDesc sorts the generations by descending order of their creation date.
func (gs Generations) SortByCreatedDateDesc() {
	descOrder := func(i, j int) bool {
		return gs[i].Created.After(gs[j].Created)
	}
	gs.SortIfNeeded(descOrder)
}

// SortIfNeeded sorts the collection of generations by the given sortFunc,
// but checks first if the list isn't already sorted by the same sortFunc.
// If the collection is already sorted, nothing is done.
func (gs Generations) SortIfNeeded(sortFunc func(i, j int) bool) {
	if !sort.SliceIsSorted(gs, sortFunc) {
		sort.Slice(gs, sortFunc)
	}
}

// OfBucket returns a collection of all generations of all objects in a bucket.
func OfBucket(bucket *storage.BucketHandle) (Generations, error) {
	return OfPath(bucket, "")
}

// OfPath returns a collection of generations of objects in a bucket that have the given path prefix.
// If an empty string is given as a path prefix, all generations of all objects in the bucket will be returned.
func OfPath(bucket *storage.BucketHandle, path string) (Generations, error) {
	var res Generations
	ctx := context.Background()

	query := &storage.Query{Prefix: path, Versions: true}

	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return res, fmt.Errorf("getting objects in path '%s': %w", path, err)
		}
		res = append(res, Generation{attrs})
	}

	return res, nil
}

// OfPathByName returns a collection of generations of objects in a bucket that have the given path prefix, indexed
// by path name as a map of object paths to its generation list.
// If an empty string is given as a path prefix, all generations of all objects in the bucket
// will be returned.
func OfPathByName(bucket *storage.BucketHandle, path string) (map[string]Generations, error) {
	res := make(map[string]Generations)
	ctx := context.Background()

	query := &storage.Query{Prefix: path, Versions: true}

	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return res, fmt.Errorf("getting objects in path '%s': %w", path, err)
		}

		if _, ok := res[attrs.Name]; !ok {
			res[attrs.Name] = Generations{}
		}
		res[attrs.Name] = append(res[attrs.Name], Generation{attrs})
	}

	return res, nil
}
