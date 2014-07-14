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

func updateFn(cmd *Command, args []string) error {
  configureLogging()
  configurePipeline()

  if len(args) > 0 {
    return fmt.Errorf("Too many arguments for 'update'")
  }

  // compose a new grapnel file path out of the old package path
  if packageFileName == "" {
    packageFileName = path.Join(path.Dir(packageFileName), "grapnel.toml")
  }

  // get dependencies from the grapnel file
  log.Info("loading package file: '%s'", packageFileName)
  deplist, err := LoadGrapnelDepsfile(packageFileName)
  if err != nil {
    return err
  } else if deplist == nil {
    // TODO: fail over to update instead?
    return fmt.Errorf("Cannot open grapnel file: '%s'", packageFileName)
  }

  log.Info("loaded %d deps", len(deplist))

  // compose a new lock file path out of the old package path
  if lockFileName == "" {
    lockFileName = path.Join(path.Dir(packageFileName), "grapnel-lock.toml")
  }

  // open it now before we expend any real effort
  lockFile, err := os.Create(lockFileName)
  defer lockFile.Close()
  if err != nil {
    log.Error("Cannot open lock file: '%s'", lockFileName)
    return err
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

  // write the library data out
  log.Info("Writing lock file")
  for _, lib := range libs {
    lib.ToToml(lockFile)
  }

  log.Info("Update complete")
  return nil
}

var updateCmd = Command{
  Desc: "Ensure that the latest dependencies are installed and ready for use.",
  Help: " Installs packages at 'targetPath', from configured package file.\n" +
    "\nDefaults:\n" +
    "  Package file = " + packageFileName + "\n" +
    "  Lock file = " + lockFileName + "\n" +
    "  Target path = " + targetPath + "\n",
  Flags: FlagMap {
    "pconfig": &Flag {
      Alias: "p",
      Desc: "Grapnel package file",
      ArgDesc: "[filename]",
      Fn: StringFlagFn(&packageFileName),
    },
    "lockfile": &Flag {
      Alias: "l",
      Desc: "Where to write the grapnel lock file",
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
  Fn: updateFn,
}