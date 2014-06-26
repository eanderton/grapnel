package main
/*
  Bundler-like support for Go.  That works.
*/

// archive support like *everything else* (tar.gz, zip, and bz2)

import (
  . "grapnel"
  "github.com/spf13/cobra"
  "path"
  "os"
  "fmt"
  toml "github.com/pelletier/go-toml"
  log "github.com/ngmoco/timber"
)

// application configurables
var configFileName string
var targetPath string

func configurePipeline() {
  // configure Git
  TypeResolvers["git"] = GitResolver
  UrlSchemeResolvers["git"] = GitResolver
  UrlHostResolvers["github.com"] = GitResolver 
  InstallIgnorePatterns["git"] = GitIgnorePattern
}

func loadDependencies(filename string) ([]*Dependency, error) {
  tree, err := toml.LoadFile(filename)
  if err != nil {
    return nil, err
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

func installFn(cmd *cobra.Command, args []string) {
  InitLogging()
  configurePipeline()

  // TODO: resolve path for targetPath
  log.Info("installing to: %v", targetPath) 
  if err := os.MkdirAll(targetPath, 0755); err != nil {
    log.Fatal(err)
  }

  // get the dependencies from the config file
  deplist, err := loadDependencies(configFileName)
  if err != nil {
    log.Fatal(err)
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
    log.Fatal(err)
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
    log.Fatal("Cannot open lock file: ", err)
  }
  for _, lib := range libs {
    lib.ToToml(pkgFile)
  }
  
  log.Info("Install complete")
}

func updateFn(cmd *cobra.Command, args []string) {
  InitLogging()
  // Do Stuff Here
  log.Info("Update complete")
}

func infoFn(cmd *cobra.Command, args[]string) {
  InitLogging()
  
  // get the dependencies from the config file
  deplist, err := loadDependencies(configFileName)
  if err != nil {
    log.Error(err)
    return
  }
  for _, dep := range deplist {
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
  
  rootCmd.PersistentFlags().BoolVarP(&LoggingVerbose, "verbose", "v", false, "verbose output")
  rootCmd.PersistentFlags().BoolVarP(&LoggingQuiet, "quiet", "q", false, "quiet output")

  rootCmd.PersistentFlags().StringVarP(&configFileName, "config", "c", "./toml",
    "configuration file")

  rootCmd.PersistentFlags().StringVarP(&targetPath, "target", "t", "./src",
    "where to manage packages")

//  rootCmd.AddCommand(installCmd, updateCmd, infoCmd)
  rootCmd.AddCommand(installCmd)
  rootCmd.Execute()
}
