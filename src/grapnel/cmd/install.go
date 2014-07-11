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
  toml "github.com/pelletier/go-toml"
)


func loadDependencies(filename string) ([]*Dependency, error) {
  tree, err := toml.LoadFile(filename)
  if err != nil {
    return nil, err
  }

  items := tree.Get("dependencies").([]*toml.TomlTree)
  if items == nil {
    log.Fatal("No dependencies to process")
  }

  deplist := make([]*Dependency, 0)
  for idx, item := range items {
    if dep, err := NewDependencyFromToml(item); err != nil {
      return nil, fmt.Errorf("In dependency #%d: %v", idx, err)
    } else {
      deplist = append(deplist, dep)
    }
  }

  return deplist, nil
}

// TODO: if no lock file can be found, then fail over to update instead
// TODO: move lock file writing to updateFn()
// TODO: prefer grapnel-lock.toml over grapnel.toml here

func installFn(cmd *Command, args []string) error {
  configureLogging()
  configurePipeline()

  if len(args) > 1 {
    return fmt.Errorf("Too many arguments for 'install'")
  }
  if len(args) == 1 {
    targetPath = args[0]
  }

  // compose a new lock file path out of the old package path
  if lockFileName == "" {
    lockFileName = path.Join(path.Dir(packageFileName), "grapnel-lock.toml")
  }

  // open it now before we expend any real effort
  pkgFile, err := os.Create(lockFileName)
  defer pkgFile.Close()
  if err != nil {
    log.Error("Cannot open lock file: '%s'", lockFileName)
    return err
  }

  log.Info("installing to: %v", targetPath)
  if err := os.MkdirAll(targetPath, 0755); err != nil {
    return err
  }

  // get the dependencies from the config file
  deplist, err := loadDependencies(packageFileName)
  if err != nil {
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
    lib.ToToml(pkgFile)
  }

  log.Info("Install complete")
  return nil
}

var installCmd = Command{
  Desc: "Ensure that dependencies are installed and ready for use.",
  ArgDesc: "[targetPath]",
  Help: " Installs packages at 'targetPath', from configured package file.\n" +
    "\nDefaults:\n" +
    "  Package file = " + packageFileName + "\n" +
    "  Target path = " + targetPath + "\n",
  Flags: FlagMap {
    "pconfig": &Flag {
      Alias: "p",
      Desc: "Grapnel package file",
      ArgDesc: "[filename]",
      Fn: StringFlagFn(&packageFileName),
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
