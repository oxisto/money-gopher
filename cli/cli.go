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

// repl provides a simple Read-Eval-Print-Loop (REPL) to issue commands to an
// integrated client.
package cli

import (
	"fmt"
	"os"
)

var cmdMap map[string]Command = make(map[string]Command)

// AddCommand adds a command using the specific symbol.
func AddCommand(symbol string, cmd Command) {
	cmdMap[symbol] = cmd
}

// Session holds all necessary information about the current CLI session.
type Session struct {
}

// Run runs our CLI command, based on the args. We keep it very simple for now
// without any extra package, so we just take the first arg and see if if
// matches any of our commands
func Run(args []string) {
	var (
		cmd Command
		ok  bool
		s   *Session
	)

	// Create a new session. TODO(oxisto): We do not yet have auth, but in the
	// future we need to fetch a token here
	s = new(Session)

	// Try to look up command in our command map
	cmd, ok = cmdMap[args[1]]
	if ok {
		cmd.Exec(s, os.Args[1:]...)
	} else {
		fmt.Print("Command not found.\n")
	}
}
