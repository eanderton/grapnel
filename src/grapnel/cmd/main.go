package main

/*
  Bundler-like support for Go.  That works.
*/

// archive support like *everything else* (tar.gz, zip, and bz2)

// walk dependencies for imports

// rewrite imports and scrub scm tracking dirs

import (
  . "grapnel"
  "github.com/spf13/cobra"
  "path/filepath"
  "go/build"
  "os"
  "fmt"
  log "github.com/ngmoco/timber"
)

// application configurables
var verbose bool
var quiet bool
var configFileName string
var targetPath string

// set all supported SCM processors
var resolvers map[string]Resolver = map[string]Resolver{
  "git": NewGitSCM(),
  "std": NewStdResolver(),
}

// toml configuration file and .lock file
type dependencyConfig struct {
  Deps map[string]*Dependency
}

// processing stages for resolution
var stagedDeps = NewDepArray()
var unresolvedDeps = NewDepSet()
var resolvedDeps = NewDepArray()
var finalizedDeps = NewDepSet()

func cleanupDeps() {
  cleanFn := func(dep *Dependency) bool {
    dep.Destroy()
    return true
  }
  stagedDeps.Each(cleanFn)
  unresolvedDeps.Each(cleanFn)
  resolvedDeps.Each(cleanFn)
  finalizedDeps.Each(cleanFn)
}

func initLogging() {
  // configure logging level based on flags
  var logLevel log.Level
  if quiet {
    logLevel = log.ERROR
  } else if verbose {
    logLevel = log.INFO
  } else {
    logLevel = log.WARNING
  }
  // set a vanilla console writer
  log.AddLogger(log.ConfigLogger{
    LogWriter: new(log.ConsoleWriter),
    Level:     logLevel,
    Formatter: log.NewPatFormatter("[%L] %M"),
  })
}

/*
  Dependencies 'breathe' in and out of the above queues until
  everything is finalized, or there is an error.

  0. The grapnel file is parsed and the root dependencies are
     staged for processing. 
  1. Filter all staged deps into unresolved, while performing
     de-duplication of both entries within the staged queue, 
     and the finalized dependencies map. Conflicts are resolved
     here, or the program ends with an error.
  2. Work all unresolved deps concurrently.  The program talks 
     to configured SCMs here. Additional imports discovered in
     newly downloaded dependencies are staged for later. The 
     dependency is then added to the resolved queue
  3. After all goroutines are done draining the unresolved 
     queue, the resolved deps are recorded in the finalized
     dependency map.
  4. If there are more items in staged, continue with step 1.
  5. The finalized dependency map is written to a lockfile.
*/

//TODO: move into unified match+configure method on the SCM
func setResolver(dep *Dependency) {
  var resolver Resolver = nil
  for _, resolver = range resolvers {
    if resolver.MatchDependency(dep) {
      if err := resolver.ValidateDependency(dep); err != nil {
        log.Fatalf("Failed to validate dependency '%s'", dep.Import)
      }
      break
    }
  }
  if resolver == nil {
    log.Fatalf("Could not find compatible resolver for dependency '%s'", dep.Import)
  }
  dep.Resolver = resolver
}

// Returns true if a should replace b; panic if no resolution can be found 
func resolveCollision(a *Dependency, b *Dependency) bool {
  log.Info("Resolving collision: '%s'", a.Import)
  //TODO: implement me - resolution should be SCM specific
  return true
}

//TODO: merge resolver methods into the dependency itself
func resolveDependency(dep *Dependency) {
  if err := dep.Resolver.FetchDependency(dep); err != nil {
    log.Fatalf("Failed to fetch dependency: '%s'", dep.Import)
  }
 
  // look for additional dependencies
  if dep.Type != "std" { 
    pkg, err := build.ImportDir(dep.TempRoot, 0)
    if err != nil {
      log.Error("", err)
      log.Error("Could not gather import information for dependency '%s'", dep.Import)
    }
    for _, importName := range pkg.Imports {
      newDep := &Dependency{Import: importName}
      setResolver(newDep)
      stagedDeps.Push(newDep)
    }
  }

  // mark this as resolved
  if dep.Type != "std" {
    log.Info("Resolved: '%s'", dep.Import)
  }
  resolvedDeps.Push(dep) 
}

func installFn(cmd *cobra.Command, args []string) {
  defer cleanupDeps()
  initLogging()
  config := &dependencyConfig{}
  LoadTomlFile(configFileName, config)

  // figure out the install target
  pwd, err := os.Getwd()
  if err != nil {
    log.Fatal("%s", err.Error())
  }
  installTarget := filepath.Join(pwd, targetPath)

  // fill unresolved channel with dependencies to work 
  log.Info("Installing Dependencies")
  for name, dep := range config.Deps {
    if err := dep.Init(); err != nil {
      log.Error(err)
      log.Fatalf("Failed to init dependency '%s'", name)
    }
    setResolver(dep) 
    stagedDeps.Push(dep)
  }
  stagedDeps.Each(func(dep *Dependency) bool {
    log.Info(dep.Import)
    return true
  })

  for stagedDeps.Len() > 0 { //|| unresolvedDeps.Len() > 0 {
    // bounce staged to unresolved
    stagedDeps.Each(func(dep *Dependency) bool {
      // resolve collision with other unresolved imports
      if otherDep, ok := unresolvedDeps.Find(dep.Import); ok {
        if replace := resolveCollision(dep, otherDep); replace {
          unresolvedDeps.Remove(dep.Import)
        } else {
          return true  // keep the existing import
        }
      }
      // resolve collision with other finalized imoprts
      if otherDep, ok := finalizedDeps.Find(dep.Import); ok {
        if replace := resolveCollision(dep, otherDep); replace {
          finalizedDeps.Remove(dep.Import)
        } else {
          return true  // keep the existing import
        }
      }
      unresolvedDeps.Insert(dep)
      return true
    })
    stagedDeps.Clear()
    
    // resolve all unresolved dependencies concurrently
    log.Info("resolving depenencies")
    unresolvedDeps.GoEach(resolveDependency)
    unresolvedDeps.Clear()

    // bounce resolved deps to finalized
    resolvedDeps.Each(func(dep *Dependency) bool {
      if _, ok := finalizedDeps.Find(dep.Import); ok {
        log.Fatal("Import already finalized: '%s'", dep.Import)
      }
      // record finalized dependency
      finalizedDeps.Insert(dep)
      return true
    })
    resolvedDeps.Clear()
  }

  // install everything
  finalizedDeps.Each(func(dep *Dependency) bool {
    dep.Resolver.InstallDependency(dep, installTarget)
    return true
  })

  // log everything in the lockfile
  // TODO: make lockfile desintaion configurable
  lockfileName := filepath.Join(installTarget, "lock")
  lockfile, err := os.Create(lockfileName) // For read access.
  defer lockfile.Close()

  fmt.Fprintln(lockfile, "# Auto-generated Grapnel Lock-file")
  fmt.Fprintln(lockfile, "# DO NOT MODIFY")
  if err != nil {
    log.Fatal(err)
  }
  finalizedDeps.Each(func(dep *Dependency) bool {
    dep.ToToml(lockfile)
    return true
  })
  
  // crawl imports for more dependencies
  log.Info("Install complete")
}

func updateFn(cmd *cobra.Command, args []string) {
  initLogging()
  // Do Stuff Here
}

func infoFn(cmd *cobra.Command, args[]string) {
  initLogging()
  config := &dependencyConfig{}
  LoadTomlFile(configFileName, config)
  for _, dep := range config.Deps {
    dep.ToToml(os.Stdout) 
  }
}

var installCmd = &cobra.Command{
  Use: "install",
  Short: "Ensure that dependencies are installed and ready for use.",
  Run: installFn,
}

var updateCmd = &cobra.Command{
  Use: "update",
  Short: "Update the current environment",
  Run: updateFn,
}

var infoCmd = &cobra.Command{
  Use: "info",
  Short: "Query packer for package information",
  Run: infoFn,
}

func main() {
  defer log.Close()
  var rootCmd = &cobra.Command{Use: "grapnel"}
  
  rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
  rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output")

  rootCmd.PersistentFlags().StringVarP(&configFileName, "config", "c", "./toml",
    "configuration file")

  rootCmd.PersistentFlags().StringVarP(&targetPath, "target", "t", "./src",
    "where to manage packages")

  rootCmd.AddCommand(installCmd, updateCmd, infoCmd)
  rootCmd.Execute()
}
