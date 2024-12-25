// Copyright 2023 Christian Banse
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
//
// This file is part of The Money Gopher.

package commands

import (
	"github.com/urfave/cli/v3"
)

var CLI = &cli.Command{
	Name:  "mgo",
	Usage: "The money-gopher CLI",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "debug", Usage: "Enable debug mode."},
	},
	Commands: []*cli.Command{
		{
			Name:  "login",
			Usage: "Login to the money-gopher service",
		},
		PortfolioCmd,
	},
}

/*var CLI struct {
	Debug bool `help:"Enable debug mode."`

	Login LoginCmd `cmd:"" help:"Login command."`

	Security SecurityCmd `cmd:"" help:"Security commands."`
	//Portfolio   PortfolioCmd   `cmd:"" help:"Portfolio commands."`
	BankAccount BankAccountCmd `cmd:"" help:"Bank account commands."`

	Completion kongcompletion.Completion `cmd:"" help:"Outputs shell code for initializing tab completions" hidden:"" completion-shell-default:"false"`
}
*/
