package main

/*
  Bundler-like support for Go.  That works.
*/

// specify url, branch, tag, and commit like gopack does

// archive support like *everything else* (tar.gz, zip, and bz2)

// scm integration
//use exec.Command to run hg, svn, and git

// walk dependencies for imports

// rewrite imports and scrub scm tracking dirs

import (
  "github.com/spf13/cobra"
  "github.com/spf13/viper"
  "github.com/spf13/cast"
  "grapnel"
  "path/filepath"
  log "github.com/ngmoco/timber"
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

func loadConfig() bool {
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

  // load the config file
  viper.SetConfigType("toml")
  viper.SetConfigFile(configFileName)
  if err := viper.ReadInConfig(); err != nil {
    log.Error("Error reading in config file '%s'", configFileName)
    log.Error(err.Error())
    return false
  }
  return true
}

func installFn(cmd *cobra.Command, args []string) {
  if !loadConfig() {
    return
  }

  // figure out the install target
  pwd, err := os.Getwd()
  if err != nil {
    log.Fatal("%s", err.Error())
  }
  installTarget := filepath.Join(pwd, targetPath)


  log.Info("Installing Dependencies")
  allDeps := make(map[string]*grapnel.Spec)
  for name, dep := range viper.GetStringMap("deps") {
    spec, err := grapnel.NewSpec(name, cast.ToStringMap(dep))
    if err != nil {
      log.Fatal("%s", err.Error())
    }
    for scmName, scm := range scms {
      log.Info("matching: %s", scmName)
      if scm.MatchDependencySpec(spec) {
        if err := scm.ValidateDependencySpec(spec); err != nil {
          continue
        }
        if err := scm.InstallDependency(spec, installTarget); err != nil {
          log.Fatalf("Failed to install dependency: '%s'", name)
        }
        // save for later
        allDeps[name] = spec
        break
      } else {
        log.Info("no match for: %s", scmName)
      }
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
  for _, dep := range allDeps {
    dep.ToToml(lockfile)
  }

  // crawl imports for more dependencies
  log.Info("Install complete")
}

func updateFn(cmd *cobra.Command, args []string) {
  if !loadConfig() {
    return
  }
  // Do Stuff Here
}

func infoFn(cmd *cobra.Command, args[]string) {
  if !loadConfig() {
    return
  }
  if verbose {
    viper.Debug()
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
