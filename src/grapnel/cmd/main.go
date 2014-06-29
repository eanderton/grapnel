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

// application configurables
var (
  configFileName string
  targetPath string
  flagQuiet bool
  flagVerbose bool
  flagDebug bool
)


func configureLogging() {
  if flagDebug {
    log.SetGlobalLogLevel(log.DEBUG)
  } else if flagQuiet {
    log.SetGlobalLogLevel(log.ERROR)
  } else if flagVerbose {
    log.SetGlobalLogLevel(log.INFO)
  }
}


func configurePipeline() {
  // configure Git
  TypeResolvers["git"] = GitResolver
  UrlSchemeResolvers["git"] = GitResolver
  UrlHostResolvers["github.com"] = GitResolver 
  InstallIgnorePatterns["git"] = GitIgnorePattern

  // TODO: other SCMs
}


func loadDependencies(filename string) ([]*Dependency, error) {
  tree, err := toml.LoadFile(filename)
  if err != nil {
    return nil, err
  }

  tree = tree.Get("deps").(*toml.TomlTree)
  if tree == nil {
    log.Fatal("No dependencies to process")
  }

  deplist := make([]*Dependency, 0)
  for _,key := range tree.Keys() {
    depTree := tree.Get(key).(*toml.TomlTree)
    if dep, err := NewDependencyFromToml(key, depTree); err != nil {
      return nil, fmt.Errorf("In section '%v': %v", key, err)
    } else {
      deplist = append(deplist, dep)
    }
  }

  return deplist, nil
}

func installFn(cmd *Command, args []string) error {
  configureLogging()
  configurePipeline()

  // TODO: resolve path for targetPath
  log.Info("installing to: %v", targetPath) 
  if err := os.MkdirAll(targetPath, 0755); err != nil {
    return err
  }

  // get the dependencies from the config file
  deplist, err := loadDependencies(configFileName)
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
  // TODO: make a part of a proper package file instead
  log.Info("Writing lock file.")
  pkgFile, err := os.Create(path.Join(targetPath, "grapnel-lock.toml"))
  defer pkgFile.Close()
  if err != nil {
    log.Error("Cannot open lock file")
    return err 
  }
  for _, lib := range libs {
    lib.ToToml(pkgFile)
  }
  
  log.Info("Install complete")
  return nil
}

func updateFn(cmd *Command, args []string) error {
  configureLogging()
  // Do Stuff Here
  log.Info("Update complete")
  return nil
}

func infoFn(cmd *Command, args[]string) error {
  configureLogging()
  
  // get the dependencies from the config file
  deplist, err := loadDependencies(configFileName)
  if err != nil {
    return err
  }
  for _, dep := range deplist {
    dep.ToToml(os.Stdout) 
  }
  return nil
}

var rootCmd = &Command{
  Desc: "grapnel",
  Flags: FlagMap {
    "quiet":   BoolFlag(&flagQuiet),
    "q":       BoolFlag(&flagQuiet),
    "verbose": BoolFlag(&flagVerbose),
    "v":       BoolFlag(&flagVerbose),
    "debug":   BoolFlag(&flagDebug),
    "config":  StringFlag(&configFileName),
    "c":       StringFlag(&configFileName),
    "target":  StringFlag(&targetPath),
    "t":       StringFlag(&targetPath),
  },
  Commands: CommandMap {
    "install": &Command{
      Desc: "Ensure that dependencies are installed and ready for use.",
      Fn: installFn,
    },
    "update": &Command{
      Desc: "Update the current environment.",
      Fn: updateFn,
    },
    "info": &Command{
      Desc: "Query packer for package information",
      Fn: infoFn,
    },
  },
}

func main() {
  log.SetFlags(0)
  if err := rootCmd.Execute(os.Args[1:]...); err != nil {
    log.Error(err)
    rootCmd.ShowHelp()
  }
}
