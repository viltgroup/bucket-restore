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

package gcprestore

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
)

// GetStorageClientFromFile creates a new storage client with the given credentials file.
// If the path to the key file is the empty string, a client with the default credentials
// will be created.
func GetStorageClientFromFile(keyFile string) (*storage.Client, context.Context, error) {
	var client *storage.Client
	var err error

	ctx := context.Background()

	if keyFile == "" {
		client, err = storage.NewClient(ctx)
	} else {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(keyFile))
	}

	if err != nil {
		return client, ctx, fmt.Errorf("creating storage client: %w", err)
	}

	return client, ctx, nil
}
