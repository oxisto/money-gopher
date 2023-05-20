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
package repl

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var cmdMap map[string]Command = make(map[string]Command)

// AddCommand adds a command using the specific symbol.
func AddCommand(symbol string, cmd Command) {
	cmdMap[symbol] = cmd
}

// REPL is a Read-Eval-Print-Loop (REPL).
type REPL struct {
	done bool
}

// Run executes the Read-Eval-Print-Loop (REPL). This will block until the loop
// is done.
func (r *REPL) Run() {
	for !r.done {
		var (
			s    *bufio.Scanner
			line string
			rr   []string
			ok   bool
			cmd  Command
		)

		fmt.Print("(🤑) ")
		s = bufio.NewScanner(os.Stdin)
		// TODO(oxisto): We want to also split on tabs in the future to support
		//  auto-completion
		s.Split(bufio.ScanLines)
		ok = s.Scan()
		if !ok {
			continue
		}

		// Retrieve a line of text and split strings by white-space (expect within a
		// quote) to compute our command and arguments
		line = s.Text()
		var q = false
		rr = strings.FieldsFunc(line, func(r rune) bool {
			if r == '"' {
				// Flip in-quote status
				q = !q
			}

			return !q && r == ' '
		})

		// Cleanup any remaining quotes
		rr2 := make([]string, 0, len(rr))
		for _, s := range rr {
			rr2 = append(rr2, strings.Trim(s, "\""))
		}

		if len(rr2) == 0 {
			continue
		}

		// Try to look up command in our command map
		cmd, ok = cmdMap[rr2[0]]
		if ok {
			cmd.Exec(r, rr2[1:]...)
		} else {
			fmt.Print("Command not found.\n")
		}
	}
}
