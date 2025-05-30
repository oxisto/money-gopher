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
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/oxisto/money-gopher/cli/commands"
)

func main() {
	if err := commands.CLICmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("Error while running command", tint.Err(err))
	}
}
