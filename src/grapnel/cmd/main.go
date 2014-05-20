package main

/*
  Bundler-like support for Go.  That works.
*/

// archive support like *everything else* (tar.gz, zip, and bz2)

// walk dependencies for imports

// rewrite imports and scrub scm tracking dirs

import (
  "github.com/spf13/cobra"
  _ "github.com/spf13/cast"
  "grapnel"
  "path/filepath"
  log "github.com/ngmoco/timber"
  _ "io/ioutil"
  "os"
  "fmt"
)

var verbose bool
var quiet bool
var configFileName string
var targetPath string

// set all supported SCM processors
var scms map[string]grapnel.SCM = map[string]grapnel.SCM{
  "git": grapnel.NewGitSCM(),
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
  Deps map[string]*grapnel.Spec
}

func installFn(cmd *cobra.Command, args []string) {
  initLogging()
  config := &dependencyConfig{}
  grapnel.LoadTomlFile(configFileName, config)

  // figure out the install target
  pwd, err := os.Getwd()
  if err != nil {
    log.Fatal("%s", err.Error())
  }
  installTarget := filepath.Join(pwd, targetPath)

  log.Info("Installing Dependencies")

  // validation of dependencies
  for name, dep := range config.Deps {
    if err := dep.InitSpec(); err != nil {
      log.Error(err)
      log.Fatalf("Failed to init dependency '%s'", name)
    }
    for _, scm := range scms {
      if scm.MatchDependencySpec(dep) {
        if err := scm.ValidateDependencySpec(dep); err != nil {
          log.Fatalf("Failed to validate dependency '%s'", name)
        }
        dep.Scm = scm
      }
    }
    if dep.Scm == nil {
      log.Fatalf("Could not find compatible SCM for dependency '%s'", name)
    }
  }

  // actual instatllation loop
  for name, dep := range config.Deps {  
    if err := dep.Scm.InstallDependency(dep, installTarget); err != nil {
      log.Fatalf("Failed to install dependency: '%s'", name)
    }
  }

  // TODO: iterate over deps to find additional dependencies

  // log everything in the lockfile
  // TODO: make lockfile desintaion configurable
  lockfileName := filepath.Join(pwd, "grapnel.lock")
  lockfile, err := os.Create(lockfileName) // For read access.
  defer lockfile.Close()

  fmt.Fprintln(lockfile, "# Auto-generated Grapnel Lock-file")
  fmt.Fprintln(lockfile, "# DO NOT MODIFY")
  if err != nil {
    log.Fatal(err)
  }
  for _, dep := range config.Deps {
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
  grapnel.LoadTomlFile(configFileName, config)
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
  var rootCmd = &cobra.Command{Use: "grapnel"}
  
  rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
  rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet output")

  rootCmd.PersistentFlags().StringVarP(&configFileName, "config", "c", "./grapnel.toml",
    "configuration file")

  rootCmd.PersistentFlags().StringVarP(&targetPath, "target", "t", "./src",
    "where to manage packages")

  rootCmd.AddCommand(installCmd, updateCmd, infoCmd)
  rootCmd.Execute()
  log.Close()
}
