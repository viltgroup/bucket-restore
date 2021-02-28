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

import (
	"fmt"
	"time"
)

var supportedLayouts = []string{

	"January 02, 2006, 15:04:05 (UTC-07:00)", // AWS object listing
	"Jan 02, 2006, 15:04:05 (UTC-07:00)",     // AWS object listing, month shortened
	"January 02, 2006, 15:04:05 (-07:00)",    // AWS object listing, timezone shortened
	"Jan 02, 2006, 15:04:05 (-07:00)",        // AWS object listing, month and timezone shortened

	"Jan 02, 2006, 03:04:05 PM -07:00",     // GCP object listing + timezone
	"Jan 02, 2006, 03:04:05 PM MST",        // GCP object listing + timezone name
	"January 02, 2006, 03:04:05 PM -07:00", // GCP object listing + timezone, complete month
	"January 02, 2006, 03:04:05 PM MST",    // GCP object listing + timezone name, complete month

	"02-01-2006 15:04:05 -07:00",
	"02-01-2006 15:04:05 MST",
	"02-01-2006 03:04:05 PM -07:00",
	"02-01-2006 03:04:05 PM MST",

	"2006-01-02 15:04:05 -07:00",
	"2006-01-02 03:04:05 PM MST",
	"2006-01-02 03:04:05 PM -07:00",
	"2006-01-02 03:04:05 PM MST",

	time.RFC1123, // "Mon, 02 Jan 2006 15:04:05 MST"
	time.RFC3339, // "2006-01-02T15:04:05Z07:00"
}

// ParseTimestamp parsed a string with date, time and timezone into a time.Time
// by trying to parse the string against different time layout strings.
func ParseTimestamp(timestamp string) (time.Time, error) {
	var ts time.Time

	for _, layout := range supportedLayouts {
		ts, error := time.Parse(layout, timestamp)
		if error == nil {
			return ts, nil
		}
	}

	return ts, fmt.Errorf("could not parse timestamp '%v' into any of the supported formats", timestamp)
}
