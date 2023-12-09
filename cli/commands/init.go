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

var CLI struct {
	Debug bool `help:"Enable debug mode."`

	Security  SecurityCmd  `cmd:"" help:"Security commands."`
	Portfolio PortfolioCmd `cmd:"" help:"Portfolio commands."`

	Completion kongcompletion.Completion `cmd:"" help:"Outputs shell code for initializing tab completions" hidden:""`
}
