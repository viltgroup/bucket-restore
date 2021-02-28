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
	// embed blank import to allow go:embed directive to inject version file
	_ "embed"
	"fmt"
	"github.com/spf13/cobra"
)

//go:embed VERSION
var version string

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the current version of brestore",
	Long: "" +
		"Description:\n" +
		"  Shows the current version of brestore.\n",
	Run:                   versionEntryPoint,
	DisableFlagsInUseLine: true,
}

func versionEntryPoint(cmd *cobra.Command, args []string) {
	fmt.Printf("brestore version %s", version)
}
