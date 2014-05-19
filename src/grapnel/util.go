package grapnel

import (
  "os"
  so "stackoverflow"
  "io/ioutil"
  "os/exec"
  "path/filepath"
  "regexp"
  log "github.com/ngmoco/timber"
)

func RunCmdOut(cmd string, args... string) (string, error) {
  out, err := exec.Command(cmd, args...).CombinedOutput()
  if err != nil {
    if _, ok := err.(*exec.ExitError); ok {
      log.Info("%s", out)
    } else {
      log.Info("%s", err.Error())
    }
  }
  return string(out[:]), err
}

func RunCmd(cmd string, args... string) (error) {
  _, err := RunCmdOut(cmd, args...)
  return err
}

// Function chain for deferred callback registration
// Designed to be used with 'defer', to allow closures and
// called functions to register defer/cleanup functions
// within a parent scope
type FnChain struct {
  chain []func()
}

// FnChain constructor
func NewFnChain() *FnChain {
  return &FnChain{}
}

// Adds a func to the chain
// TODO: change to LIFO order
func (self *FnChain) Add(fn func()) {
  self.chain = append(self.chain, fn)
}

// Calls all functions on the chain
// TODO: handle panic/recover to allow all chain elements to be called
func (self *FnChain) Invoke() {
  for _, fn := range self.chain {
    fn()
  }
}

// Creates a temporary directory
// Destroys the temporary directory when the chain is invoked
func createTempDir(deferChain *FnChain) (string, error) {
  tempRoot, err := ioutil.TempDir("","")
  if err != nil {
    log.Info("%s", err.Error())
    return "", log.Error("Could not create temporary directory: '%s'", tempRoot)
  }
  deferChain.Add(func() {
    log.Info("Cleaning up '%s'", tempRoot)
    os.RemoveAll(tempRoot)
  })
  return tempRoot, err
}

// TODO: tighten up this implementation
func pushDir(newDir string, deferChain *FnChain) error {
  // save the current working directory
  oldDir, err := os.Getwd()
  if err != nil {
    return err
  }
  // change to new directory
  if err := os.Chdir(newDir); err != nil {
    return err
  }
  deferChain.Add(func() {
    os.Chdir(oldDir)
  })
  return nil
}

// Copies a file tree from src to dest
func CopyFileTree(dest string, src string, ignore string) error {
  // create a callback for filtering
  var ignoreFn func(name string) bool
  if ignore == "" {
    ignoreFn = func(string) bool {
      return false
    }
  } else {
    if ignoreRegex, err := regexp.Compile(ignore); err != nil {
      return log.Error("Failed to compile ignore regex")
    } else {
      ignoreFn = func(name string) bool {
        return ignoreRegex.MatchString(name)
      }
    }
  }

  return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
    if err != nil {
      log.Info("%s", err.Error())
      return log.Error("Error while walking file tree")
    }
    relativePath, _ := filepath.Rel(src, path)
    destPath := filepath.Join(dest, relativePath)
    if info.IsDir() {
      if ignoreFn(info.Name()) {
        return filepath.SkipDir
      }
      dir := filepath.Base(destPath)
      if err := os.MkdirAll(dir, 0755); err != nil {
        return log.Error("Could not create directory: '%s'", dir)
      }
    } else { 
      if ignoreFn(info.Name()) {
        return nil  // skip file
      }
      log.Info("Copying: %s", relativePath)
      if err := so.CopyFileContents(path, destPath); err != nil {
        return log.Error("Could not copy file '%s' to '%s'", path, destPath)
      }
    }
    return nil 
  })
}

