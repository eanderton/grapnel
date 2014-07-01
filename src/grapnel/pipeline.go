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
  "strings"
  log "grapnel/log"
)

// TODO: some kind of resolver middleware capability to support
// host subsitution, URL rewriting, and dependency type manipulation

// resolver types
type ResolverFn func(*Dependency) (*Library,error)
type ResolverFnMap map[string]ResolverFn
type IgnorePatternMap map[string]string

// resolver registration
var (
  TypeResolvers ResolverFnMap = make(ResolverFnMap)
  UrlSchemeResolvers ResolverFnMap = make(ResolverFnMap)
  UrlHostResolvers ResolverFnMap = make(ResolverFnMap)
  InstallIgnorePatterns IgnorePatternMap = make(IgnorePatternMap)
)

func GetResolver(dep *Dependency) ResolverFn {
  // TODO: apply middleware filtering here (?)

  // attempt to resolve by type
  if fn, ok := TypeResolvers[dep.Type]; ok {
    return fn
  }

  if dep.Url != nil {
    // attempt to resolve by url host
    if fn, ok := UrlHostResolvers[dep.Url.Host]; ok {
      return fn
    }

    // attempt to resolve by url host
    if fn, ok := UrlSchemeResolvers[dep.Url.Scheme]; ok {
      return fn
    }
  }

  // attempt to resolve by hostpart of import
  if dep.Import != "" {
    parts := strings.Split(dep.Import, "/")
    if len(parts) > 0 {
      hostPart := parts[0]
      log.Info("resolving by hostpart: %v", hostPart)
      if fn, ok := UrlHostResolvers[hostPart]; ok {
        return fn
      }
    }
  }
  return nil
}

func Resolve(dep *Dependency) (*Library, error) {
  // resolve the library
  fn := GetResolver(dep)
  if fn == nil {
    return nil, fmt.Errorf("Cannot identify resolver for dependency: '%v'", dep.Name)
  }

  // resolve the dependency
  lib, err := fn(dep)
  if err != nil {
    return nil, err
  }

  // add additional deps from this library
  if err := AddDependencies(lib); err != nil {
    return nil, err
  }
  return lib, nil
}

// remove duplicates while preserving dependency order
func DeduplicateDeps(deps []*Dependency) ([]*Dependency, error) {
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

func LibResolveDeps(libs map[string]*Library, deps []*Dependency) ([]*Dependency, error) {
  tempQueue := make([]*Dependency, 0)
  for _, dep := range deps {
    if lib, ok := libs[dep.Import]; ok {
      if !dep.IsSatisfiedBy(lib.Version) {
        return nil, fmt.Errorf("Cannot reconcile '%v'", dep.Import)
      }
    } else {
      tempQueue = append(tempQueue, dep)
    }
  }
  return tempQueue, nil
}

func ResolveDependencies(deps []*Dependency) (map[string]*Library, error) {
  resolved := make(map[string]*Library)
  results := make(chan *Library)
  errors := make(chan error)
  workQueue := deps

  for len(workQueue) > 0 {
    // de-duplicate the queue
    var err error
    if workQueue, err = DeduplicateDeps(workQueue); err != nil {
      return nil, err
    }

    // look for already resolved deps that may match
    if workQueue, err = LibResolveDeps(resolved, workQueue); err != nil {
      return nil, err
    }

    // spawn goroutines for each dependency to be resolved
    for _, dep := range workQueue {
      go func(dep *Dependency) {
        lib, err := Resolve(dep)
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
      select {
      case lib := <- results:
        resolved[lib.Import] = lib
        log.Debug("Reconciled library: %s", lib.Import)
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
  return resolved, nil
}

func InstallLibraries(installRoot string, libs map[string]*Library) error {
  for name, lib := range libs {
    pattern, ok := InstallIgnorePatterns[lib.Type]
    if !ok {
      pattern = ""
    }
    if err := lib.Install(installRoot, pattern); err != nil {
      return fmt.Errorf("While installing %v: %v", name, err)
    }
  }
  return nil
}
