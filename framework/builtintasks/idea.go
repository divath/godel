// Copyright 2016 Palantir Technologies, Inc.
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

package builtintasks

import (
	"github.com/nmiyake/pkg/dirs"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/palantir/godel/framework/builtintasks/idea"
	"github.com/palantir/godel/framework/godellauncher"
)

func IDEATask() godellauncher.Task {
	const intellijCmdUsage = "Create IntelliJ project files for this project"

	ideaCmd := &cobra.Command{
		Use:   "idea",
		Short: intellijCmdUsage,
		RunE: func(cmd *cobra.Command, args []string) error {
			// This command has subcommands, but does not accept any arguments itself. If the execution has reached this
			// point and "args" is non-empty, treat it as an unknown command (rather than just executing this command
			// and ignoring the extra arguments). Avoids executing the wrong command on a subcommand typo.
			if len(args) > 0 {
				return godellauncher.UnknownCommandError(cmd, args)
			}

			wd, err := dirs.GetwdEvalSymLinks()
			if err != nil {
				return errors.Wrapf(err, "failed to determine working directory")
			}
			return idea.CreateIntelliJFiles(wd)
		},
	}
	goglandSubcommand := &cobra.Command{
		Use:   "gogland",
		Short: "Create Gogland project files for this project",
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, err := dirs.GetwdEvalSymLinks()
			if err != nil {
				return errors.Wrapf(err, "failed to determine working directory")
			}
			return idea.CreateGoglandFiles(wd)
		},
	}
	intelliJSubcommand := &cobra.Command{
		Use:   "intellij",
		Short: intellijCmdUsage,
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, err := dirs.GetwdEvalSymLinks()
			if err != nil {
				return errors.Wrapf(err, "failed to determine working directory")
			}
			return idea.CreateGoglandFiles(wd)
		},
	}
	cleanSubcommand := &cobra.Command{
		Use:   "clean",
		Short: "Remove the IDEA project files for this project",
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, err := dirs.GetwdEvalSymLinks()
			if err != nil {
				return errors.Wrapf(err, "failed to determine working directory")
			}
			return idea.CleanIDEAFiles(wd)
		},
	}

	ideaCmd.AddCommand(
		goglandSubcommand,
		intelliJSubcommand,
		cleanSubcommand,
	)
	return godellauncher.CobraCLITask(ideaCmd)
}
