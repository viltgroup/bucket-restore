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

package history

import (
	"time"

	"github.com/viltgroup/bucket-restore/internal/brestore/awsrestore/history/versions"
)

// Enumeration of PathStatus.
const (
	// Represents a file that does not exist and most likely was never created
	NOT_EXISTENT PathStatus = iota
	// Represents a file that exists at a given point
	EXISTS
	// Represents a file that existed but has been deleted
	DELETED
)

// PathStatus represents the status of a file in its history.
type PathStatus int

// String converts an PathStatus to a string.
// Implements the Stringer interface.
func (fs PathStatus) String() string {
	switch fs {
	case NOT_EXISTENT:
		return "Not Existent"
	case EXISTS:
		return "Exists"
	case DELETED:
		return "Deleted"
	default:
		return "Unknown Status"
	}
}

// PathState represents the state of a path/object at a specific generation.
type PathState struct {
	// Status of the file in its lifetime
	PathStatus
	versions.Version
}

// StateAtTime gives the state of a file/object at a certain point in time, given its versions.
// The collection of versions must refer to the same object/path.
func StateAtTime(versions versions.Versions, t time.Time) PathState {
	res, _ := StateDiffAtTime(versions, t)
	return res
}

// StateDiffAtTime gives the state of a file/object at a certain point in time, given its versions.
// Also return a second value with the last known state of the object.
// The collection of versions must refer to the same object/path.
func StateDiffAtTime(versions versions.Versions, t time.Time) (PathState, PathState) {
	nGens := len(versions)

	if nGens == 0 {
		return PathState{PathStatus: NOT_EXISTENT}, PathState{PathStatus: NOT_EXISTENT}
	}

	versions.SortByLastModifiedAsc()
	firstVersion, lastVersion := versions[0], versions[nGens-1]
	lastVersionState := StateOfVersion(lastVersion)

	if t.After(lastVersion.LastModified) {
		return StateOfVersionAtTime(lastVersion, t), lastVersionState
	}

	if t.Before(firstVersion.LastModified) {
		return PathState{PathStatus: NOT_EXISTENT, Version: firstVersion}, lastVersionState
	}

	var res PathState

	for i := nGens - 1; i >= 0; i-- {
		if t.After(versions[i].LastModified) {
			res = StateOfVersionAtTime(versions[i], t)
			break
		}
	}

	return res, lastVersionState
}

// StateOfVersion returns the last known path state of a generation.
func StateOfVersion(v versions.Version) PathState {
	res := PathState{Version: v}

	if v.IsDeleteMarker {
		res.PathStatus = DELETED
	} else {
		res.PathStatus = EXISTS
	}

	return res
}

// StateOfVersionAtTime returns the path state of a generation at a given point in time.
func StateOfVersionAtTime(v versions.Version, t time.Time) PathState {
	res := PathState{Version: v}

	if v.IsDeleteMarker && v.LastModified.Before(t) {
		res.PathStatus = DELETED
	} else {
		res.PathStatus = EXISTS
	}

	return res
}
