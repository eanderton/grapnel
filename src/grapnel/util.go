package grapnel

import (
  "os"
  so "stackoverflow"
  "os/exec"
  "path/filepath"
  "regexp"
  log "github.com/ngmoco/timber"
)

type RunContext struct {
  WorkingDirectory string
  CombinedOutput string
}

func NewRunContext(workingDirectory string) *RunContext {
  return &RunContext {
    WorkingDirectory: workingDirectory,
  }
}

func (self *RunContext) Run(cmd string, args... string) error {
  cmdObj := exec.Command(cmd, args...)
  cmdObj.Dir = self.WorkingDirectory
  log.Debug("%v %v", cmd, args)
  out, err := cmdObj.CombinedOutput()
  self.CombinedOutput = string(out)
  if err != nil {
    if _, ok := err.(*exec.ExitError); ok {
      log.Info("%s", out)
    } else {
      log.Info("%s", err.Error())
    }
  }
  return err
}

func (self *RunContext) Start(cmd string, args... string) (*exec.Cmd, error) {
  cmdObj := exec.Command(cmd, args...)
  cmdObj.Dir = self.WorkingDirectory
  err := cmdObj.Start()
  return cmdObj, err 
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
