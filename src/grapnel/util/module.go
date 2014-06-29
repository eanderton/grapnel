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
  "os"
  so "grapnel/stackoverflow"
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
      log.Error("%s", out)
    } else {
      log.Error("%s", err.Error())
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

func (self *RunContext) MustRun(cmd string, args... string) {
  if err := self.Run(cmd, args...); err != nil {
    log.Fatal(self.CombinedOutput)
  }
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
      //dir := filepath.Dir(destPath)
      //log.Info("Making directory: %v : %v", dir, destPath)
      //if err := os.MkdirAll(dir, 0755); err != nil {
      //  return log.Error("Could not create directory: '%s'", dir)
      //}
    } else { 
      if ignoreFn(info.Name()) {
        return nil  // skip file
      }
      log.Debug("Copying: %s", destPath)
      if err := so.CopyFileContents(path, destPath); err != nil {
        return log.Error("Could not copy file '%s' to '%s'", path, destPath)
      }
    }
    return nil 
  })
}
