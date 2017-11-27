// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

const flagOutputLocation = "output-location"
const flagBuildTag = "build-tag"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goalias {package}",
	Short: "Creates a copy of a Go package, where the copy defers execution to the original package.",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("accepts the name of a single package")
		}

		expanded, err := ExpandPackageName(args[0])
		if err != nil {
			return err
		}

		_, err = parser.ParseDir(&token.FileSet{}, expanded, nil, 0)
		if err != nil {
			return fmt.Errorf("no go package found at %q", expanded)
		}

		return nil
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		panic("unwritten code")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// ExpandPackageName finds the path to a package on the filesystem, given a package name.
func ExpandPackageName(name string) (expanded string, err error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath, err = homedir.Dir()
		if err != nil {
			return
		}
		gopath = filepath.Join(gopath, "go")
	}

	expanded = filepath.Join(gopath, "src", name)
	return
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.goalias.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP(flagBuildTag, flagBuildTag[:1], false, "Place a comment at the top of the output files so that they will be ignored unless go1.9 or later used.")
	viper.BindPFlag(flagBuildTag, rootCmd.Flags().Lookup(flagBuildTag))

	rootCmd.Flags().StringP(flagOutputLocation, flagOutputLocation[:1], "", "The output location of the generated package. By default, this is stdout.")
	viper.BindPFlag(flagOutputLocation, rootCmd.Flags().Lookup(flagOutputLocation))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".goalias" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".goalias")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
