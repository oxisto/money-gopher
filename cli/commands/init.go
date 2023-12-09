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
	"fmt"

	kongcompletion "github.com/jotaen/kong-completion"
)

/*func init() {
	cli.AddCommand("list-securities", &listSecuritiesCmd{})
	cli.AddCommand("update-quote", &triggerQuoteUpdate{})
	cli.AddCommand("update-all-quotes", &triggerQuoteUpdateAll{})

	cli.AddCommand("create-portfolio", &createPortfolio{})
	cli.AddCommand("list-portfolios", &listPortfolio{})
	cli.AddCommand("portfolio-snapshot", &portfolioSnapshot{})
	cli.AddCommand("import-transactions", &importTransactions{})
}*/

type RmCmd struct {
	Force     bool `help:"Force removal."`
	Recursive bool `help:"Recursively remove files."`

	Paths []string `arg:"" name:"path" help:"Paths to remove." type:"path"`
}

func (r *RmCmd) Run() error {
	fmt.Println("rm", r.Paths)
	return nil
}

var CLI struct {
	Debug bool `help:"Enable debug mode."`

	Rm        RmCmd        `cmd:"" help:"Remove files."`
	Portfolio PortfolioCmd `cmd:"" help:"Portfolio commands."`

	Completion kongcompletion.Completion `cmd:"" help:"Outputs shell code for initialising tab completions"`
}
