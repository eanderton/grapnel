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
	"testing"
)

var (
	flagQuiet      bool
	targetPath     string
	cmdTestFnFired bool
	cmdTestFnArgs  []string
)

func resetCmdTest() {
	flagQuiet = false
	targetPath = ""
	cmdTestFnFired = false
	cmdTestFnArgs = []string{}
}

var rootCmd = &Command{
	Desc: "grapnel",
	Flags: FlagMap{
		"quiet": &Flag{
			Alias: "q",
			Desc:  "quiet output",
			Fn:    BoolFlagFn(&flagQuiet),
		},
		"target": &Flag{
			Alias: "t",
			Desc:  "target",
			Fn:    StringFlagFn(&targetPath),
		},
	},
	Commands: CommandMap{
		"test": &Command{
			Fn: cmdTestFn,
		},
	},
}

func cmdTestFn(cmd *Command, args []string) error {
	cmdTestFnFired = true
	cmdTestFnArgs = args
	return nil
}

func argsEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestExecute(t *testing.T) {
	for _, data := range []struct {
		Args     []string
		Quiet    bool
		Target   string
		CmdFired bool
		CmdArgs  []string
	}{
		{
			[]string{"cmd", "-q"},
			true, "", false,
			[]string{},
		},
		{
			[]string{"cmd", "-q", "--quiet"},
			true, "", false,
			[]string{},
		},
		{
			[]string{"cmd", "--target=foobar"},
			false, "foobar", false,
			[]string{},
		},
		{
			[]string{"cmd", "--target=", "foobar"},
			false, "foobar", false,
			[]string{},
		},
		{
			[]string{"cmd", "--target", "=", "foobar"},
			false, "foobar", false,
			[]string{},
		},
		{
			[]string{"cmd", "--target", "=foobar"},
			false, "foobar", false,
			[]string{},
		},
		{
			[]string{"cmd", "-qt", "foobar"},
			true, "foobar", false,
			[]string{},
		},
		{
			[]string{"cmd", "test", "foo", "bar", "baz"},
			false, "", true,
			[]string{"foo", "bar", "baz"},
		},
		{
			[]string{"cmd", "test", "-qt", "foo", "bar", "baz"},
			true, "foo", true,
			[]string{"bar", "baz"},
		},
	} {
		resetCmdTest()
		if err := rootCmd.Execute(data.Args...); err != nil {
			t.Errorf("%v", err)
		}
		if flagQuiet != data.Quiet {
			t.Errorf("flagQuiet is %v, expected %v", flagQuiet, data.Quiet)
		}
		if cmdTestFnFired != data.CmdFired {
			t.Errorf("cmdTestFnFired is %v, expected %v", cmdTestFnFired, data.CmdFired)
		}
		if targetPath != data.Target {
			t.Errorf("target is %v, expected %v", targetPath, data.Target)
		}
		if !argsEq(cmdTestFnArgs, data.CmdArgs) {
			t.Errorf("cmdTestFnArgs is %v, expected %v", cmdTestFnArgs, data.CmdArgs)
		}
	}
}
