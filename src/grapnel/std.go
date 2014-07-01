package grapnel

import (
  "runtime"
  "go/build"
  "strings"
  log "grapnel/log"
)

var (
  stdContext *build.Context
)

// governs stdContext - lazy creates modified copy of Default context
func getResolutionContext() *build.Context {
  if stdContext == nil {
    // reduce resolution paths to point only at root
    stdContext = &build.Context{}
    *stdContext = build.Default  // copy
    stdContext.GOPATH = runtime.GOROOT()
  }
  return stdContext
}

// checks if an import path is already globally provided
func IsStandardDependency(importName string) bool {
  context := getResolutionContext()
  _, err := context.Import(importName, "", build.FindOnly)
  return err == nil
}

func AddDependencies(lib* Library) error {
  pkg, err := build.ImportDir(lib.TempDir, 0)
  if err != nil {
    return err
  }

  for _, importName := range pkg.Imports {
    if IsStandardDependency(importName) {
      // TODO: change to debug
      log.Debug("Ignoring import: %v", importName)
    } else {
      // TODO: check for a grapnel.toml in this project
      // bring in deps unversioned
      log.Warn("Adding secondary import: %v", importName)
      dep, err := NewDependency(importName, "", "")
      if err != nil {
        return err
      }
      dep.Name = strings.Replace(importName, ".", "_", -1)
      dep.Name = strings.Replace(dep.Name, "/", "_", -1)
      lib.Dependencies = append(lib.Dependencies, dep)
    }
  }
  return nil
}
