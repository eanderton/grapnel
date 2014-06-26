package grapnel

import (
  "fmt"
  "strings"
  log "github.com/ngmoco/timber"
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

func (self *Dependency) Resolve() (*Library, error) {
  // attempt to resolve by type
  if fn, ok := TypeResolvers[self.Type]; ok {
    return fn(self)
  }

  if self.Url != nil {
    // attempt to resolve by url host
    if fn, ok := UrlHostResolvers[self.Url.Host]; ok {
      return fn(self)
    }
     
    // attempt to resolve by url host
    if fn, ok := UrlSchemeResolvers[self.Url.Scheme]; ok {
      return fn(self)
    }
  }

  // attempt to resolve by hostpart of import
  if self.Import != "" {
    parts := strings.Split(self.Import, "/")
    if len(parts) > 0 {
      hostPart := parts[0]
      log.Info("resolving by hostpart: %v", hostPart)
      if fn, ok := UrlHostResolvers[hostPart]; ok {
        return fn(self)
      }
    }
  }
  
  return nil, log.Error("Cannot identify resolver for dependency: '%v'", self.Import)
}

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
        return nil, log.Error("Cannot reconcile '%v'", dep.Import) 
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
        lib, err := dep.Resolve()
        if err != nil {
          errors <- err
        } else {
          log.Info("Resolved: '%s'", lib.Import)
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
        log.Error("%v",err)
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
