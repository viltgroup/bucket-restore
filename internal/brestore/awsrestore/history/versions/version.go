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
	"github.com/aws/aws-sdk-go/service/s3"
	"strings"
	"time"
)

// Version represents a version or delete marker of an object in AWS
type Version struct {
	Key            string
	ID             string
	LastModified   time.Time
	IsLatest       bool
	IsDeleteMarker bool
	ETag           string
	Size           int64
}

// String converts a Version into a string.
func (v *Version) String() string {
	var latestPrefix, markerString string

	if v.IsLatest {
		latestPrefix = " LATEST"
	}

	if v.IsDeleteMarker {
		markerString = " (Delete Marker)"
	}

	return fmt.Sprintf("%s (%s, %v)%s%s",
		v.Key,
		v.ID,
		v.LastModified,
		markerString,
		latestPrefix,
	)
}

// StringWithoutName converts a Version to a string, omitting the name of the file.
// This is useful for situations where the name is implicit,
func (v *Version) StringWithoutName() string {
	var latestPrefix, markerString string

	if v.IsLatest {
		latestPrefix = " LATEST"
	}

	if v.IsDeleteMarker {
		markerString = " (Delete Marker)"
	}

	return fmt.Sprintf("%s (%v)%s%s",
		v.ID,
		v.LastModified,
		markerString,
		latestPrefix,
	)
}

// FromAWSVersion builds a Version object from a s3.ObjectVersion
func FromAWSVersion(obj *s3.ObjectVersion) Version {
	return Version{
		Key:            *obj.Key,
		ID:             *obj.VersionId,
		LastModified:   *obj.LastModified,
		IsLatest:       *obj.IsLatest,
		IsDeleteMarker: false,
		ETag:           strings.Trim(*obj.ETag, "\""),
		Size:           *obj.Size,
	}
}

// FromAWSDeleteMarker builds a Version object from a s3.DeleteMarkerEntry
func FromAWSDeleteMarker(marker *s3.DeleteMarkerEntry) Version {
	return Version{
		Key:            *marker.Key,
		ID:             *marker.VersionId,
		LastModified:   *marker.LastModified,
		IsLatest:       *marker.IsLatest,
		IsDeleteMarker: true,
	}
}
