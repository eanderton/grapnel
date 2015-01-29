package grapnel
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
  log "grapnel/log"
)

type LibSource interface {
  Resolve(*Dependency) (*Library, error)
  ToDSD(*Library) string
}

type LibSourceMap map[string]LibSource

type Resolver struct {
  LibSources LibSourceMap
  RewriteRules RewriteRuleArray
}

func NewResolver() *Resolver {
  return &Resolver {
    LibSources: LibSourceMap{},
    RewriteRules: RewriteRuleArray{},
  }
}

func (self *Resolver) AddRewriteRules(rules RewriteRuleArray) {
  self.RewriteRules = append(self.RewriteRules, rules...)
}

// resolve a single dependency
func (self *Resolver) Resolve(dep *Dependency) (*Library, error) {
  // apply rewrite rules
  // TODO: consider preserving original dependency
  if err := self.RewriteRules.Apply(dep); err != nil {
    return nil, err
  }

  // match by registered type - rewrite rules should have set 'type' by now
  if source, ok := self.LibSources[dep.Type]; ok {
    var lib *Library
    var err error

    // resolve through the LibSource
    lib, err = source.Resolve(dep)
    if err != nil {
      return nil, err
    }

    // follow up with lib specific touches
    err = lib.AddDependencies()
    if err != nil {
      return nil, err
    }

    return lib, nil
  }

  return nil, fmt.Errorf("Cannot identify resolver for dependency: '%v'", dep.Import)
}

// remove duplicates while preserving dependency order
func (self *Resolver) DeduplicateDeps(deps []*Dependency) ([]*Dependency, error) {
  tempQueue := make([]*Dependency, 0)
  for ii, src := range deps {
    jj := ii+1;
    for ; jj < len(deps); jj++ {
      dest := deps[jj]
      if src.Import == dest.Import {
        // reconcile the collision and place it at jj
        if next, err := src.Reconcile(dest); err != nil {
          return nil, err
        } else {
          deps[jj] = next
          break
        }
      }
    }
    // add src to the queue if there were no collisions
    if jj >= len(deps) {
      tempQueue = append(tempQueue, src)
    }
  }
  return tempQueue, nil
}

// resolve against libraries that are already provided
func (self *Resolver) LibResolveDeps(libs map[string]*Library, deps []*Dependency) ([]*Dependency, error) {
  tempQueue := make([]*Dependency, 0)
  for _, dep := range deps {
    if lib, ok := libs[dep.Import]; ok {
      if !dep.VersionSpec.IsSatisfiedBy(lib.Version) {
        return nil, fmt.Errorf("Cannot reconcile '%v'", dep.Import)
      }
    } else {
      tempQueue = append(tempQueue, dep)
    }
  }
  return tempQueue, nil
}

// resolve all dependencies against configuration
func (self *Resolver) ResolveDependencies(deps []*Dependency) ([]*Library, error) {
  masterLibs := []*Library{}
  resolved := map[string]*Library{}
  results := make(chan *Library)
  errors := make(chan error)
  workQueue := deps

  for len(workQueue) > 0 {
    // de-duplicate the queue
    var err error
    if workQueue, err = self.DeduplicateDeps(workQueue); err != nil {
      return nil, err
    }

    // look for already resolved deps that may match
    if workQueue, err = self.LibResolveDeps(resolved, workQueue); err != nil {
      return nil, err
    }

    // spawn goroutines for each dependency to be resolved
    for _, dep := range workQueue {
      go func(dep *Dependency) {
        lib, err := self.Resolve(dep)
        if err != nil {
          errors <- err
        } else {
          results <- lib
        }
      }(dep)
    }

    // wait on all goroutines to finish or fail
    tempQueue := make([]*Dependency, 0)
    failed := false
    for ii := 0; ii < len(workQueue); ii++ {
      log.Debug("working on %s of  %s", ii, len(workQueue))
      select {
      case lib := <- results:
        log.Debug("Reconciled library: %s", lib.Import)
        resolved[lib.Import] = lib
        masterLibs = append(masterLibs, lib)
        for _, importPath := range lib.Provides {
          log.Debug("Submodule:  %s", importPath)
          resolved[importPath] = lib
        }
        tempQueue = append(tempQueue, lib.Dependencies...)
      case err := <- errors:
        log.Error(err)
        failed = true
      }
    }
    if failed {
      return nil, fmt.Errorf("One or more errors while resolving dependencies.")
    }
    workQueue = tempQueue
  }
  return masterLibs, nil
}

func (self *Resolver) ToDsd(filename string, libs []*Library) error {
  fmt.Printf("#!/bin/bash\n")
  fmt.Printf("# Grapnel Dead-simple Downloader\n\n")
  // TODO: call toDSD on all libs via assigned resolver
  return nil
}

func (self *Resolver) InstallLibraries(installRoot string, libs []*Library) error {
  for _, lib := range libs {
    if err := lib.Install(installRoot); err != nil {
      return fmt.Errorf("While installing %v: %v", lib.Import, err)
    }
  }
  return nil
}
