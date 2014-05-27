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
  log "github.com/ngmoco/timber"
  "go/build"
  "sync"
  "os"
  "fmt"
)

var verbose bool
var quiet bool
var configFileName string
var targetPath string

// set all supported SCM processors
var scms map[string]SCM = map[string]SCM{
  "git": NewGitSCM(),
}

type DependencyMap map[string]*Dependency

var stagedDeps = NewDepArray()
var unresolvedDeps = NewDepArray()
var resolvedDeps = NewDepArray()
var finalizedDeps = make(DependencyMap)

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


func resolveDependency(dep *Dependency, installTarget string) {
  // find an SCM to resolve this dependency  
  for _, scm := range scms {
    if scm.MatchDependency(dep) {
      if err := scm.ValidateDependency(dep); err != nil {
        log.Fatalf("Failed to validate dependency '%s'", dep.Import)
      }
      dep.Scm = scm
    }
  }
  if dep.Scm == nil {
    log.Fatalf("Could not find compatible SCM for dependency '%s'", dep.Import)
  }
  if err := dep.Scm.InstallDependency(dep, installTarget); err != nil {
    log.Fatalf("Failed to install dependency: '%s'", dep.Import)
  }
  
  // look for additional dependencies
  // TODO: actually log these as staged
  installPath := filepath.Join(installTarget, dep.Import)
  pkg, err := build.ImportDir(installPath, 0)
  if err != nil {
    log.Error(err)
    log.Error("Could not gather import information for dependency '%s'", dep.Name)
  }
  for _, importName := range pkg.Imports {
    fmt.Printf("%s depends on %s\n",dep.Name, importName)
    //stagedDeps.Add(&Dependency{Import: importName})
  }

  // mark this as resolved
  log.Info("Resolved: '%s'", dep.Import)
  resolvedDeps.Push(dep) 
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

type dependencyConfig struct {
  Deps map[string]*Dependency
}

func installFn(cmd *cobra.Command, args []string) {
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
    stagedDeps.Push(dep)
  }
  stagedDeps.Each(func(dep *Dependency) bool {
    log.Info(dep.Import)
    return true
  })

  for stagedDeps.Len() > 0 { //|| unresolvedDeps.Len() > 0 {
    var wg sync.WaitGroup
    // bounce staged to unresolved
    // TODO: be smarter about handling collisions
    var unresolvedMap DependencyMap
    stagedDeps.Each(func(dep *Dependency) bool {
      if _, ok := unresolvedMap[dep.Import]; ok {
        log.Fatalf("Import collision: '%s'", dep.Import)
      }
      unresolvedDeps.Push(dep)
      return true
    })
    stagedDeps.Clear()

    // resolve all unresolved dependencies concurrently
    log.Info("resolving depenencies")
    wg.Add(unresolvedDeps.Len())
    unresolvedDeps.GoEach(func(dep *Dependency) {
      resolveDependency(dep, installTarget)
      wg.Done()
    })

    // wait until all the work is done
    wg.Wait()
    log.Info("all goroutines completed")
    unresolvedDeps.Clear()

    // bounce resolved deps to finalized
    resolvedDeps.Each(func(dep *Dependency) bool {
      if _, ok := finalizedDeps[dep.Import]; ok {
        log.Fatalf("Import already finalized: '%s'", dep.Import)
      }
      // record finalized dependency
      finalizedDeps[dep.Import] = dep
      return true
    })
  }

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
  for _, dep := range finalizedDeps {
    dep.ToToml(lockfile)
  }
  
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
