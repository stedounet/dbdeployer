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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
	"github.com/datacharmer/dbdeployer/downloads"
	"github.com/datacharmer/dbdeployer/globals"
	"github.com/datacharmer/dbdeployer/sandbox"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dbdeployer",
	Short: "Installs multiple MySQL servers on the same host",
	Long: `dbdeployer makes MySQL server installation an easy task.
Runs single, multiple, and replicated sandboxes.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
	Version: common.VersionDef,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// If the command line was not set in the abbreviations module,
	// we save it here, before it is processed by Cobra
	if len(common.CommandLineArgs) == 0 {
		common.CommandLineArgs = append(common.CommandLineArgs, os.Args...)
	}

	// Sets flags normalization (allows --double-word and --double_word)
	// and aliases
	for _, c := range rootCmd.Commands() {
		for _, c1 := range c.Commands() {
			for _, c2 := range c1.Commands() {
				customizeFlags(c2, strings.Join([]string{c.Name(), c1.Name(), c2.Name()}, "."))
			}
			customizeFlags(c1, strings.Join([]string{c.Name(), c1.Name()}, "."))
		}
		customizeFlags(c, c.Name())
	}
	if err := rootCmd.Execute(); err != nil {
		common.Exitf(1, "%s", err)
	}
}

func setPflag(cmd *cobra.Command, key string, abbr string, envVar string, defaultVar string, helpStr string, isSlice bool) {
	var defaultValue string
	if envVar != "" {
		defaultValue = os.Getenv(envVar)
	}
	if defaultValue == "" {
		defaultValue = defaultVar
	}
	if isSlice {
		cmd.PersistentFlags().StringSliceP(key, abbr, []string{defaultValue}, helpStr)
	} else {
		cmd.PersistentFlags().StringP(key, abbr, defaultValue, helpStr)
	}
}

func checkDefaultsFile() {
	flags := rootCmd.Flags()
	defaults.CustomConfigurationFile, _ = flags.GetString(globals.ConfigLabel)
	if defaults.CustomConfigurationFile != defaults.ConfigurationFile {
		if common.FileExists(defaults.CustomConfigurationFile) {
			defaults.ConfigurationFile = defaults.CustomConfigurationFile
		} else {
			common.Exitf(1, globals.ErrFileNotFound, defaults.CustomConfigurationFile)
		}
	}
	defaults.LoadConfiguration()

	shellPath, _ := flags.GetString(globals.ShellPathLabel)

	shellPath, err := common.GetBashPath(shellPath)
	if err != nil {
		common.Exitf(1, "error validating shell '%s'", err)
	}
	defaults.UpdateDefaults(globals.ShellPathLabel, shellPath, false)
	err = sandbox.FillMockTemplates()
	if err != nil {
		common.Exitf(1, "error filling mock templates: %s", err)
	}
	globals.MockTemplatesFilled = true
	loadTemplates()
	if downloads.TarballRegistryFileExist() {
		err = downloads.LoadTarballFileInfo()
		if err != nil {
			fmt.Printf("tarball load from %s failed: %s", downloads.TarballFileRegistry, err)
			fmt.Println("Tarball list not loaded. Using defaults. Correct the issues listed above before using again.")
		}
	}
}

func customizeFlags(cmd *cobra.Command, cmdName string) {
	normalizeFlags := func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		from := []string{".", "_"}
		to := "-"
		for _, sep := range from {
			name = strings.Replace(name, sep, to, -1)
		}
		for _, alias := range globals.FlagAliases {
			if name == alias.Alias && (cmdName == alias.Command || alias.Command == "ANY") {
				name = alias.FlagName
			}
		}
		return pflag.NormalizedName(name)
	}
	cmd.Flags().SetNormalizeFunc(normalizeFlags)
}

func init() {
	cobra.OnInitialize(checkDefaultsFile)
	rootCmd.PersistentFlags().StringVar(&defaults.CustomConfigurationFile, globals.ConfigLabel, defaults.ConfigurationFile, "configuration file")
	setPflag(rootCmd, globals.SandboxHomeLabel, "", "SANDBOX_HOME", defaults.Defaults().SandboxHome, "Sandbox deployment directory", false)
	setPflag(rootCmd, globals.SandboxBinaryLabel, "", "SANDBOX_BINARY", defaults.Defaults().SandboxBinary, "Binary repository", false)
	setPflag(rootCmd, globals.ShellPathLabel, "", "SHELL_PATH", globals.ShellPathValue, "Which shell to use for generated scripts", false)

	rootCmd.InitDefaultVersionFlag()

	// Indicates that we're using dbdeployer command line interface
	// rather than calling its sandbox creation functions from other apps.
	globals.UsingDbDeployer = true

}
