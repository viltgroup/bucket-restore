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
	"github.com/viltgroup/bucket-restore/internal/brestore/gcprestore/history/generations"
	"time"
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
	// Current generation of the path/object. This value must be ignored if status is NOT_EXISTENT
	Generation int64
	// Name of the file/directory. Typically this is actually the full path of the object
	Name string
	// MD5 checksum of the object
	MD5 []byte
}

// StateAtTime gives the state of a file/object at a certain point in time, given its generations.
// The collection of generations must refer to the same object/path.
func StateAtTime(gens generations.Generations, t time.Time) PathState {
	res, _ := StateDiffAtTime(gens, t)
	return res
}

// StateDiffAtTime gives the state of a file/object at a certain point in time, given its generations.
// Also return a second value with the last known state of the object.
// The collection of generations must refer to the same object/path.
func StateDiffAtTime(gens generations.Generations, t time.Time) (PathState, PathState) {
	nGens := len(gens)

	if nGens == 0 {
		return PathState{PathStatus: NOT_EXISTENT}, PathState{PathStatus: NOT_EXISTENT}
	}

	gens.SortByCreatedDateAsc()
	firstGen, lastGen := gens[0], gens[nGens-1]
	lastGenState := StateOfGeneration(lastGen)

	if t.After(lastGen.Created) {
		return StateOfGenerationAtTime(lastGen, t), lastGenState
	}

	if t.Before(firstGen.Created) {
		return PathState{PathStatus: NOT_EXISTENT, Name: firstGen.Name}, lastGenState
	}

	var res PathState

	for i := nGens - 1; i >= 0; i-- {
		if t.After(gens[i].Created) {
			res = StateOfGenerationAtTime(gens[i], t)
			break
		}
	}

	return res, lastGenState
}

// StateOfGeneration returns the last known path state of a generation.
func StateOfGeneration(g generations.Generation) PathState {
	res := PathState{Generation: g.Generation, Name: g.Name, MD5: g.MD5}

	if !g.Deleted.IsZero() {
		res.PathStatus = DELETED
	} else {
		res.PathStatus = EXISTS
	}

	return res
}

// StateOfGenerationAtTime returns the path state of a generation at a given point in time.
func StateOfGenerationAtTime(g generations.Generation, t time.Time) PathState {
	res := PathState{Generation: g.Generation, Name: g.Name, MD5: g.MD5}

	if !g.Deleted.IsZero() && g.Deleted.Before(t) {
		res.PathStatus = DELETED
	} else {
		res.PathStatus = EXISTS
	}

	return res
}
