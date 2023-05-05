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

// package commands contains commands that can be executed by the REPL.
package commands

import (
	"context"
	"log"
	"net/http"

	"github.com/bufbuild/connect-go"
	portfoliov1 "github.com/oxisto/money-gopher/gen"
	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/repl"
)

type listSecuritiesCmd struct{}

// Exec implements [repl.Command]
func (cmd *listSecuritiesCmd) Exec(r *repl.REPL, args ...string) {
	client := portfoliov1connect.NewSecuritiesServiceClient(http.DefaultClient, "http://localhost:8080")
	res, err := client.ListSecurities(context.Background(), connect.NewRequest(&portfoliov1.ListSecuritiesRequest{}))
	if err != nil {
		log.Println(err)
	}

	log.Println(res.Msg.Securities)
}

type triggerQuoteUpdate struct{}

// Exec implements [repl.Command]
func (cmd *triggerQuoteUpdate) Exec(r *repl.REPL, args ...string) {
	client := portfoliov1connect.NewSecuritiesServiceClient(http.DefaultClient, "http://localhost:8080")
	_, err := client.TriggerSecurityQuoteUpdate(
		context.Background(),
		connect.NewRequest(&portfoliov1.TriggerQuoteUpdateRequest{
			SecurityName: args[0],
		}),
	)
	if err != nil {
		log.Println(err)
	}
}

func init() {
	repl.AddCommand("list-securities", &listSecuritiesCmd{})
	repl.AddCommand("update-quote", &triggerQuoteUpdate{})
}
