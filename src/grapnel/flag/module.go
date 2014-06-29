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
  "strings"
)

type FlagFn func(name string, values[] string) (int, error)

type Flag struct {
  Alias string
  Desc string
  Fn FlagFn
}

type CommandFn func(cmd *Command, args []string) error
type FlagMap map[string]*Flag
type CommandMap map[string]*Command

type Command struct {
  Alias string
  Desc string
  Help string
  Fn CommandFn
  Commands CommandMap
  Flags FlagMap
}

// internal dispatch for flags - eases composition of the Execute loop
func dispatchFlag(cmdName string, flags FlagMap, name string, values []string) (int, error) {
  if flag, ok := flags[name]; ok {
    if consumed, err := flag.Fn(name, values); err != nil {
      return consumed, err
    } else {
      return consumed, nil
    }
  } 
  return 0, fmt.Errorf("Unknown flag '%v' on command '%v'", name, cmdName)
}

// Process args as subcommands and flags.
func (self *Command) Execute(args... string) error {

  // default to help if no args
  if len(args) == 0 {
    return self.ShowHelp()
  }
  
  // figure out commands and remaining args to use based on command map
  var cmdName string
  flags := make(FlagMap)
  cmd := self
  ii := 0
  for ii = ii; ii < len(args); ii++ {
    // add the flags and break if no more subcommands to process
    for k,v := range cmd.Flags {
      flags[k] = v
      if v.Alias != "" {
        flags[v.Alias] = v
      }
    }
    if len(cmd.Commands) == 0 { break }
  
    // verify the command name
    nextCmdName := args[ii]
    if nextCmdName == "help" {
      // special case for 'help'
      return cmd.showHelp(args[ii+1:])
    }
    nextCmd, ok := cmd.Commands[nextCmdName]
    if !ok {
      break
    }

    // advance to next scope
    cmdName = nextCmdName
    cmd = nextCmd
  }

  // Parse flags and args
  // NOTE: continue where we left off in previous loop
  posArgs := []string{}
  for ii = ii; ii < len(args); ii++ {
    //fmt.Printf("%v [%v]\n", ii, args[ii])
    name := args[ii]

    // special case for help
    if name == "--help" || name == "-h" {
      return cmd.showHelp(args[ii+1:])
    }

    // flag parsing
    if strings.HasPrefix(name, "-") {
      if strings.HasPrefix(name, "--") {
        value := ""
        name = name[2:] // strip leading '--'
        // handle multiple special cases for var=name handling
        if strings.HasSuffix(name, "=") {
          //--opt= value
          name = name[:len(name)-1]
          if ii+1 < len(args) {
            ii++
            value = args[ii]
          }
        } else if idx := strings.Index(name, "="); idx > 0 {
          //--opt=value
          value = name[idx+1:]
          name = name[:idx]
        } else if ii+1 < len(args) && strings.HasPrefix(args[ii+1], "=") {
          ii++
          if args[ii] == "=" {
            //--opt = value
            if ii+1 < len(args) {
              ii++
              value = args[ii]
            }
          } else {
            //--opt =value
            value = args[ii][1:]
          }
        } 
        // dispatch the flag
        if _, err := dispatchFlag(cmdName, flags, name, []string{value}); err != nil {
          return err
        }

      } else {
        name = name[1:] // strip leading '-'
        // handle multiple single-char flags smushed together
        // NOTE: stop short of last char
        var jj int = 0
        for jj = jj; jj < len(name)-1; jj++ {
          internalFlag := string(name[jj])
          if _, err := dispatchFlag(cmdName, flags, internalFlag, []string{""}); err != nil {
            return err
          }
        }
        // reset name and get optional value
        // NOTE: this is done so the last single-char in the set gets a value argument
        name = string(name[jj])
        
        // execute the flag
        flagArgs := args[ii+1:]
        if consumed, err := dispatchFlag(cmdName, flags, name, flagArgs); err != nil {
          return err
        } else {
          ii += consumed
        }
      }
    } else {
      // simple arg
      posArgs = append(posArgs, name)
    }
  }
  // execute the function
  if cmd.Fn != nil {
    return cmd.Fn(cmd, posArgs) 
  } else if len(posArgs) > 0 {
    return fmt.Errorf("Extra arguments passed to command")
  }
  return nil
}


// internal help command that takes an array to ease calling
// NOTE; only the first element of args is considered
func (self *Command) showHelp(args []string) error {
  if len(args) == 0 {
    return self.ShowHelp()
  } 
  cmdName := args[0]
  if searchCmd, ok := self.Commands[cmdName]; ok  {
    return searchCmd.ShowHelp()  
  }
  return fmt.Errorf("'%v' is not a valid command", cmdName)
}


func optStr(name string) string {
  if len(name) == 1 {
    return "-" + name
  } else {
    return "--" + name
  }
}

// Shows help for the command
func (self *Command) ShowHelp() error {
  // Usage section
  fmt.Printf("Usage:\n   %v", self.Alias)
  if len(self.Flags) > 0 {
    fmt.Printf(" [flags]")
  }
  if len(self.Commands) > 0 {
    fmt.Printf(" [command]")
  }
  fmt.Printf("\n")

  // Subcommand section
  if len(self.Commands) > 0 {
    fmt.Printf("\nAvailable Commands:\n")
    for k,v := range self.Commands {
      fmt.Printf("  %v %v\n", k, v.Desc)
    }
  }

  // Flag section
  if len(self.Flags) > 0 {
    fmt.Printf("\nAvailable Flags:\n")
    for k,v := range self.Flags {
      fmt.Printf("  ")
      if v.Alias != "" {
        fmt.Printf("%v,", optStr(v.Alias))
      }
      fmt.Printf("%v %v\n", optStr(k), v.Desc)
    }
  }

  // Misc help
  if self.Help != "" {
    fmt.Printf("\n%v\n", self.Help)
  }
  return nil
}
/*
Usage: 
  grapnel [command]

Available Commands: 
  install                   :: Ensure that dependencies are installed and ready for use.
  help [command]            :: Help about any command

 Available Flags:
  -c, --config="./toml": configuration file
  -q, --quiet: quiet output
  -t, --target="./src": where to manage packages
  -v, --verbose: verbose output

Use "grapnel help [command]" for more information about that command.
*/

func BoolFlagFn(ptr *bool) FlagFn {
  return func (name string, values []string) (int, error) {
    *ptr = true
    return 0, nil
  }
}

func StringFlagFn(ptr *string) FlagFn {
  return func (name string, values []string) (int, error) {
    if len(values) == 0 {
      return 0, fmt.Errorf("Flag %v requires a value", name)
    }
    *ptr = values[0]
    return 1, nil
  }
}

//NOTE: add additional types here
