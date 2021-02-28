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
	"github.com/viltgroup/bucket-restore/internal/brestore/gcprestore"

	"github.com/viltgroup/bucket-restore/internal/brestore/gcprestore/history/generations"
)

func doVersionsGCP(keyFile string, bucketName string, path string) error {
	client, _, err := gcprestore.GetStorageClientFromFile(keyFile)
	if err != nil {
		return fmt.Errorf("running 'generations' command: %w", err)
	}
	bucket := client.Bucket(bucketName)

	allGens, err := generations.OfPathByName(bucket, path)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	for name, fileGens := range allGens {
		fmt.Printf("%v\n", name)
		fileGens.SortByCreatedDateAsc()

		for _, fileGen := range fileGens {
			fmt.Printf("\tGen #%v md5: %x size: %s\n",
				fileGen.Generation,
				fileGen.MD5,
				brestore.ByteCountIECString(fileGen.Size))

			fmt.Printf("\t\tCreated: %v\n", fileGen.Created)
			if !fileGen.Deleted.IsZero() {
				fmt.Printf("\t\tDeleted: %v\n", fileGen.Deleted)
			}
			if !fileGen.Created.Equal(fileGen.Updated) {
				fmt.Printf("\t\tUpdated: %v\n", fileGen.Updated)
			}
		}

	}

	return nil
}
