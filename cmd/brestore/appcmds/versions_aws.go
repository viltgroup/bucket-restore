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
	"github.com/viltgroup/bucket-restore/internal/brestore/awsrestore/history/versions"
)

func doVersionsAWS(profile string, bucketName string, path string) error {
	client, err := awsrestore.GetS3Client(profile)
	if err != nil {
		return fmt.Errorf("unable to get s3 client for profile '%v': %w", profile, err)
	}

	vMap, err := versions.OfPathByName(client, bucketName, path)
	if err != nil {
		return fmt.Errorf("running 'versions' command: %w", err)
	}

	for name, vs := range vMap {
		vs.SortByLastModifiedAsc()
		fmt.Printf("%s\n", name)
		for _, v := range vs {
			fmt.Printf("    %v\n", v.StringWithoutName())
		}
	}

	return nil
}
