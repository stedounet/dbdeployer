// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2019 Giuseppe Maxia
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

package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/cookbook"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/globals"
)

func displayDefaults(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		common.Exit(1,
			"'defaults' requires a label",
			"Example: dbdeployer info defaults master-slave-base-port")
	}
	label := args[0]
	defaultsMap := defaults.DefaultsToMap()
	value, ok := defaultsMap[label]
	if ok {
		fmt.Println(value)
	} else {
		fmt.Printf("# ERROR: field %s not found in defaults\n", label)
	}
}

func displayAllVersions(basedir, wantedVersion, flavor string) {
	result := ""
	var versionInfoList []common.VersionInfo = common.GetVersionInfoFromDir(basedir)
	for _, verInfo := range versionInfoList {
		versionList, err := common.VersionToList(verInfo.Version)
		if err != nil {
			common.Exitf(1, "error retrieving version list from %s", verInfo.Version)
		}
		shortVersion := fmt.Sprintf("%d.%d", versionList[0], versionList[1])
		if wantedVersion == shortVersion || strings.ToLower(wantedVersion) == "all" {
			if verInfo.Flavor == flavor {
				if result != "" {
					result += " "
				}
				result += verInfo.Version
			}
		}
	}
	if result != "" {
		fmt.Println(result)
	}
}

func displayVersion(cmd *cobra.Command, args []string) {
	wantedVersion := ""
	allVersions := ""
	if len(args) > 0 {
		wantedVersion = args[0]
	}

	reNotFound := regexp.MustCompile(cookbook.VersionNotFound)
	if len(args) > 1 {
		allVersions = args[1]
	}
	flavor, _ := cmd.Flags().GetString(globals.FlavorLabel)
	showEarliest, _ := cmd.Flags().GetBool(globals.EarliestLabel)
	if flavor == "" {
		flavor = common.MySQLFlavor
	}
	if strings.ToLower(allVersions) == "all" {
		basedir, err := getAbsolutePathFromFlag(cmd, "sandbox-binary")
		common.ErrCheckExitf(err, 1, "error getting absolute path for 'sandbox-binary'")
		displayAllVersions(basedir, wantedVersion, flavor)
	} else {
		if strings.ToLower(wantedVersion) == "all" {
			result := ""
			for _, v := range globals.SupportedAllVersions {
				latest := cookbook.GetLatestVersion(v, flavor)
				if !reNotFound.MatchString(latest) {
					if result != "" {
						result += " "
					}
					result += latest
				}
			}
			if result != "" {
				fmt.Println(result)
			}
		} else {
			var result string
			if showEarliest {
				result = cookbook.GetEarliestVersion(wantedVersion, flavor)
			} else {
				result = cookbook.GetLatestVersion(wantedVersion, flavor)
			}
			if !reNotFound.MatchString(result) {
				fmt.Println(result)

			}
		}
	}
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Shows information about dbdeployer environment samples",
	Long:  `Shows current information about defaults and environment.`,
}

var infoDefaultsCmd = &cobra.Command{
	Use: "defaults field-name",

	Short: "displays a defaults value",
	Example: `
	$ dbdeployer info defaults master-slave-base-port 
`,
	Long:        `Displays one field of the defaults.`,
	Run:         displayDefaults,
	Annotations: map[string]string{"export": ExportAnnotationToJson(StringExport)},
}

var infoVersionCmd = &cobra.Command{
	Use: "version [short-version|all] [all]",

	Short: "displays the latest version available",
	Example: `
    # Shows the latest version available
    $ dbdeployer info version
    8.0.16

    # shows the latest version belonging to 5.7
    $ dbdeployer info version 5.7
    5.7.26

    # shows the latest version for every short version
    $ dbdeployer info version all
    5.0.96 5.1.73 5.5.53 5.6.41 5.7.26 8.0.16

    # shows all the versions for a given short version
    $ dbdeployer info version 8.0 all
    8.0.11 8.0.12 8.0.13 8.0.14 8.0.15 8.0.16
`,
	Long: `Displays the latest version available for deployment.
If a short version is indicated (such as 5.7, or 8.0), only the versions belonging to that short
version are searched.
If "all" is indicated after the short version, displays all versions belonging to that short version.
`,
	Run:         displayVersion,
	Annotations: map[string]string{"export": ExportAnnotationToJson(StringExport)},
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.AddCommand(infoDefaultsCmd)
	infoCmd.AddCommand(infoVersionCmd)
	setPflag(infoCmd, globals.FlavorLabel, "", "", "", "For which flavor this info is", false)
	infoCmd.PersistentFlags().Bool(globals.EarliestLabel, false, "return the earliest version")
}
