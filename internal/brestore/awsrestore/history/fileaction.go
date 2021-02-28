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
	"fmt"
	"net/url"
	"sort"
	"time"
)

// Enumeration of Actions.
const (
	CREATE Action = iota
	DELETE
	NO_ACTION
)

// Action represents an action to be taken.
type Action int

// String converts an Action to a string representation.
// Implements the Stringer interface.
func (a Action) String() string {
	switch a {
	case CREATE:
		return "Create"
	case DELETE:
		return "Delete"
	case NO_ACTION:
		return "No Action"
	default:
		return "Unknown Action"
	}
}

// FileOperand represents a file argument to an operation/action
type FileOperand struct {
	// Name of the file/path
	Key string
	// Version of the file(path
	Version string
	// Size of the file in bytes
	Size int64
}

// String converts an FileOperand to a string.
// Implements the Stringer interface.
func (fo *FileOperand) String() string {
	return fo.Key
}

// ToSourceURL converts a file operand to an URL string that can be used as argument to AWS operations
func (fo *FileOperand) ToSourceURL(bucketName string) string {
	return fmt.Sprintf("%s/%s?versionId=%s", bucketName, url.QueryEscape(fo.Key), url.QueryEscape(fo.Version))
}

// FileActions represents a collection of file actions.
type FileActions []FileAction

// SortBySourceNameLenAsc sorts the collection of file Actions by ascending order
// of the length of the path of the objects in the actions.
func (fa FileActions) SortBySourceNameLenAsc() {
	ascOrder := func(i, j int) bool {
		return len(fa[i].Source.Key) < len(fa[j].Source.Key)
	}
	fa.SortIfNeeded(ascOrder)
}

// SortIfNeeded sorts the collection of file actions by the given sortFunc,
// but checks first if the list isn't already sorted by the same sortFunc.
// If the collection is already sorted, nothing is done.
func (fa FileActions) SortIfNeeded(sortFunc func(i, j int) bool) {
	if !sort.SliceIsSorted(fa, sortFunc) {
		sort.Slice(fa, sortFunc)
	}
}

// FileAction represents an action to be taken on a file.
type FileAction struct {
	// The kind of action to apply to the file
	Action
	// The file to which the action should be applied
	Source FileOperand
	// Unmodified since pre-condition. The action should only be applied if the current
	// version of the object was not modified since this time.
	UnmodifiedPreCondition time.Time
}

// ActionForStateChange determines the action that should be taken to transition
// a file from a state to another.
func ActionForStateChange(from PathState, to PathState) FileAction {
	source := FileOperand{Key: to.Key, Version: to.Version.ID, Size: from.Size}

	switch from.PathStatus {
	case DELETED:
		if to.PathStatus == EXISTS {
			source.Size = to.Size
			return FileAction{Action: CREATE, Source: source}
		}
		return FileAction{Action: NO_ACTION, Source: source}
	case EXISTS:
		if to.PathStatus == DELETED || to.PathStatus == NOT_EXISTENT {
			source.Version = from.Version.ID
			return FileAction{Action: DELETE, Source: source, UnmodifiedPreCondition: from.LastModified}
		}
		if to.Version != from.Version && to.ETag != from.ETag {
			return FileAction{Action: CREATE, Source: source, UnmodifiedPreCondition: from.LastModified}
		}
		return FileAction{Action: NO_ACTION, Source: source}
	default:
		return FileAction{Action: NO_ACTION, Source: source}
	}
}
