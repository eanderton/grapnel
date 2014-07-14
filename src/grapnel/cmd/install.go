package main
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
  . "grapnel"
  . "grapnel/flag"
  "path"
  "os"
  "fmt"
  log "grapnel/log"
)

// TODO: if no lock file can be found, then fail over to update instead
// TODO: move lock file writing to updateFn()
// TODO: prefer grapnel-lock.toml over grapnel.toml here

func installFn(cmd *Command, args []string) error {
  configureLogging()
  configurePipeline()

  if len(args) > 0 {
    return fmt.Errorf("Too many arguments for 'install'")
  }

  // compose a new lock file path out of the old package path
  if lockFileName == "" {
    lockFileName = path.Join(path.Dir(packageFileName), "grapnel-lock.toml")
  }

  // get dependencies from the lockfile
  deplist, err := LoadGrapnelDepsfile(lockFileName)
  if err != nil {
    return err
  } else if deplist == nil {
    // TODO: fail over to update instead?
    return fmt.Errorf("Cannot open lock file: '%s'", lockFileName)
  }

  log.Info("installing to: %v", targetPath)
  if err := os.MkdirAll(targetPath, 0755); err != nil {
    return err
  }

  var libs map [string]*Library
  // cleanup
  defer func() {
    for _, lib := range libs {
      lib.Destroy()
    }
  }()

  // resolve all the dependencies
  libs, err = ResolveDependencies(deplist)
  if err != nil {
    return err
  }

  // install all the dependencies
  log.Info("Resolved %v dependencies. Installing.", len(libs))
  InstallLibraries(targetPath, libs)

  log.Info("Install complete")
  return nil
}

var installCmd = Command{
  Desc: "Ensure that locked dependencies are installed and ready for use.",
  Help: " Installs packages at 'targetPath', from configured lock file.\n" +
    "\nDefaults:\n" +
    "  Lock file = " + lockFileName + "\n" +
    "  Target path = " + targetPath + "\n",
  Flags: FlagMap {
    "lockfile": &Flag {
      Alias: "l",
      Desc: "Grapnel lock file",
      ArgDesc: "[filename]",
      Fn: StringFlagFn(&lockFileName),
    },
    "target": &Flag {
      Alias: "t",
      Desc: "Target installation path",
      ArgDesc: "[target]",
      Fn: StringFlagFn(&targetPath),
    },
  },
  Fn: installFn,
}
