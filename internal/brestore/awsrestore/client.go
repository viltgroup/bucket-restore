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

package awsrestore

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// GetS3Client tries to get a S3 client for the given profile
func GetS3Client(profile string) (*s3.S3, error) {
	options := session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}

	if profile != "" {
		options.Profile = profile
	}

	sess, err := session.NewSessionWithOptions(options)
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)

	return svc, nil
}
