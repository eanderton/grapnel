package util

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
	log "grapnel/log"
	so "grapnel/stackoverflow"
	"os"
	"os/exec"
	"path/filepath"
)

type RunContext struct {
	WorkingDirectory string
	CombinedOutput   string
}

func NewRunContext(workingDirectory string) *RunContext {
	return &RunContext{
		WorkingDirectory: workingDirectory,
	}
}

func (self *RunContext) Run(cmd string, args ...string) error {
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

func (self *RunContext) Start(cmd string, args ...string) (*exec.Cmd, error) {
	cmdObj := exec.Command(cmd, args...)
	cmdObj.Dir = self.WorkingDirectory
	err := cmdObj.Start()
	return cmdObj, err
}

func (self *RunContext) MustRun(cmd string, args ...string) {
	if err := self.Run(cmd, args...); err != nil {
		log.Fatal(self.CombinedOutput)
	}
}

// Copies a file tree from src to dest
func CopyFileTree(dest string, src string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Info("%s", err.Error())
			return fmt.Errorf("Error while walking file tree")
		}
		relativePath, _ := filepath.Rel(src, path)
		destPath := filepath.Join(dest, relativePath)
		if info.IsDir() {
			// create target directory if it's not already there
			if !so.Exists(destPath) {
				if err := os.MkdirAll(destPath, 0755); err != nil {
					return err
				}
			}
		} else {
			log.Debug("Copying: %s", destPath)
			if err := so.CopyFileContents(path, destPath); err != nil {
				return fmt.Errorf("Could not copy file '%s' to '%s'", path, destPath)
			}
		}
		return nil
	})
}

func GetDirectories(src string) ([]string, error) {
	results := []string{}

	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Info("%s", err.Error())
			return fmt.Errorf("Error while walking file tree")
		}
		relativePath, _ := filepath.Rel(src, path)
		if info.IsDir() && relativePath != "." {
			results = append(results, relativePath)
		}
		return nil
	})
	return results, err
}
