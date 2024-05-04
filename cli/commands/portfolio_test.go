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
	"testing"

	"github.com/oxisto/money-gopher/cli"
	"github.com/oxisto/money-gopher/service/portfolio/portfoliotest"
)

func TestCreatePortfolioCmd_Run(t *testing.T) {
	srv := portfoliotest.NewServer(t)
	defer srv.Close()

	type fields struct {
		Name        string
		DisplayName string
	}
	type args struct {
		s *cli.Session
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				Name:        "myportfolio",
				DisplayName: "My Portfolio",
			},
			args: args{
				s: func() *cli.Session {
					return cli.NewSession(&cli.SessionOptions{
						BaseURL:    srv.URL,
						HttpClient: srv.Client(),
					})
				}(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CreatePortfolioCmd{
				Name:        tt.fields.Name,
				DisplayName: tt.fields.DisplayName,
			}
			if err := cmd.Run(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("CreatePortfolioCmd.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
