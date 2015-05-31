package flag

/*
Copyright (c) 2014 Eric Anderton <eric.t.anderton@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

import (
	"fmt"
	"sort"
)

// internal help command that takes an array to ease calling
// NOTE; only the first element of args is considered
func (self *Command) showHelp(cmdName string, args []string) error {
	if len(args) == 0 {
		return self.ShowHelp(cmdName)
	}
	subCmd := args[0]
	if searchCmd, ok := self.Commands[subCmd]; ok {
		return searchCmd.ShowHelp(subCmd)
	}
	return fmt.Errorf("'%v' is not a valid command", cmdName)
}

func optStr(name, alias string) string {
	result := ""
	if len(alias) == 1 {
		result += "-" + alias + ","
	} else if len(alias) > 1 {
		result += "--" + alias + ","
	}
	if len(name) == 1 {
		result += "-" + name
	} else {
		result += "--" + name
	}
	return result
}

type helpRow []string
type helpSet []helpRow

func (self helpSet) Len() int           { return len(self) }
func (self helpSet) Swap(i, j int)      { self[i], self[j] = self[j], self[i] }
func (self helpSet) Less(i, j int) bool { return self[i][0] < self[j][0] }

func tablePrint(pre string, sep string, elements helpSet) {
	// calculate widths
	widths := map[int]int{}
	for _, row := range elements {
		for jj, item := range row {
			if ll, ok := widths[jj]; !ok || len(item) > ll {
				widths[jj] = len(item)
			}
		}
	}
	// generate output
	for _, row := range elements {
		fmt.Printf("%s", pre)
		for jj, item := range row {
			if widths[jj] == 0 {
				continue
			}
			if jj > 0 {
				fmt.Printf("%s", sep)
			}
			fmt.Printf("%-*s", widths[jj], item)
		}
		fmt.Printf("\n")
	}
}

// Shows help for the command
func (self *Command) ShowHelp(cmdName string) error {
	// command name
	name := cmdName
	if name == "" {
		name = self.Alias
	}

	// description
	if self.Desc != "" {
		fmt.Printf("%v\n\n", self.Desc)
	}

	// Usage section
	fmt.Printf("Usage:\n   %s", name)
	if len(self.Flags) > 0 {
		fmt.Printf(" [flags]")
	}
	if len(self.Commands) > 0 {
		fmt.Printf(" [command]")
	}
	fmt.Printf("\n")

	// Subcommand section
	if len(self.Commands) > 0 {
		commands := helpSet{}
		commands = append(commands, helpRow{
			"help", "[command]", "Displays help for a command",
		})
		for name, cmd := range self.Commands {
			commands = append(commands, helpRow{
				name, cmd.ArgDesc, cmd.Desc,
			})
		}
		sort.Sort(commands)
		fmt.Printf("\nAvailable Commands:\n")
		tablePrint("  ", " ", commands)
	}

	// Flag section
	if len(self.Flags) > 0 {
		flags := helpSet{}
		flags = append(flags, helpRow{
			"-h,--help", "[command]", "Displays help for a command",
		})
		for name, flag := range self.Flags {
			flags = append(flags, helpRow{
				optStr(name, flag.Alias), flag.ArgDesc, flag.Desc,
			})
		}
		sort.Sort(flags)
		fmt.Printf("\nAvailable Flags:\n")
		tablePrint("  ", " ", flags)
	}

	// Misc help
	if self.Help != "" {
		fmt.Printf("\n%v\n", self.Help)
	}
	return nil
}
