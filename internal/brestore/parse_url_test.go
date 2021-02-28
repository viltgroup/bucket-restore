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

package brestore

import "testing"

type UrlTestCase struct {
	Url      string
	Expected BucketURLInfo
}

func TestParseUrl(t *testing.T) {

	tests := []UrlTestCase{
		{
			Url: "s3:/mybucket",
			Expected: BucketURLInfo{
				Type:       "s3",
				BucketName: "mybucket",
				Prefix:     "",
			},
		},
		{
			Url: "s3://mybucket",
			Expected: BucketURLInfo{
				Type:       "s3",
				BucketName: "mybucket",
				Prefix:     "",
			},
		},
		{
			Url: "s3://mybucket/path",
			Expected: BucketURLInfo{
				Type:       "s3",
				BucketName: "mybucket",
				Prefix:     "path",
			},
		},
		{
			Url: "s3://mybucket/nested/path",
			Expected: BucketURLInfo{
				Type:       "s3",
				BucketName: "mybucket",
				Prefix:     "nested/path",
			},
		},
	}

	for _, test := range tests {
		info, err := ParseBucketURL(test.Url)
		if err != nil {
			t.Fatalf("error testing parsing url '%v': %v", test.Url, err)
		}

		if info.Type != test.Expected.Type ||
			info.BucketName != test.Expected.BucketName ||
			info.Prefix != test.Expected.Prefix {
			t.Fatalf("unexpected result for parse url: expected %v | got: %v", test.Expected, info)
		}

	}

}
