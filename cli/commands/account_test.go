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
	"context"
	"testing"

	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/internal/testing/clitest"
	"github.com/oxisto/money-gopher/internal/testing/servertest"
	"github.com/urfave/cli/v3"
)

func TestCreateAccount(t *testing.T) {
	srv := servertest.NewServer(internal.NewTestDB(t))
	defer srv.Close()

	type args struct {
		ctx context.Context
		cmd *cli.Command
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				ctx: clitest.NewSessionContext(t, srv),
				cmd: clitest.MockCommand(t,
					AccountCmd.Command("create").Flags,
					"--id", "myaccount",
					"--display-name", "My Account",
					"--type", "BANK",
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateAccount(tt.args.ctx, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
