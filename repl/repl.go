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

var cmdMap map[string]Command

func init() {
	cmdMap = map[string]Command{
		"quit": &quitCmd{},
	}
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
		ok = s.Scan()
		if !ok {
			continue
		}

		// Retrieve a line of text and split strings by white-space to compute
		// our command and arguments
		line = s.Text()
		rr = strings.Split(line, " ")

		// Try to look up command in our command map
		cmd, ok = cmdMap[rr[0]]
		if ok {
			cmd.Exec(r, rr[1:]...)
		} else {
			fmt.Print("Command not found.\n")
		}
	}
}
