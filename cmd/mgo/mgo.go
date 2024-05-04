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

package main

import (
	"fmt"
	"os"

	"github.com/oxisto/money-gopher/cli"
	"github.com/oxisto/money-gopher/cli/commands"

	"github.com/alecthomas/kong"
	kongcompletion "github.com/jotaen/kong-completion"
)

func main() {
	var (
		s      *cli.Session
		ctx    *kong.Context
		parser *kong.Kong
		err    error
	)

	parser = kong.Must(&commands.CLI,
		kong.Name("mgo"),
		kong.Description("A shell-like example app."),
		kong.UsageOnError(),
	)

	// Proceed as normal after kongplete.Complete.
	ctx, err = parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	// The only command we allow without a session is "Login"
	if ctx.Args[0] != "login" {
		// TODO(oxisto): Can we move this to a pre-hook?
		s, err = cli.ContinueSession()
		if err != nil {
			fmt.Println("Could not continue with existing session or session is missing. Please use `mgo login`.")
			return
		}
	}

	kongcompletion.Register(parser,
		commands.WithPredictPortfolios(s),
		commands.WithPredictSecurities(s),
	)

	err = ctx.Run(s)
	parser.FatalIfErrorf(err)
}
