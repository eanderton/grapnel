package lib

import (
	"fmt"
	"io"
	"strings"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	log "grapnel/log"
)

// https://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
// LinkFile links a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files.
func LinkFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	return
}

// CopyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func CopyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// http://stackoverflow.com/a/12527546
// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// https://stackoverflow.com/questions/17609732/expand-tilde-to-home-directory
// Returns an absolute path for 'path'; delegates to filepath.Abs for
// most of the heavy lifting. Expands '~/' in a path to the current user's
// home directory
func AbsolutePath(path string) (string, error) {
	// promote path to absolute path
	result, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	// Check in case of paths like "/something/~/something/"
	if result[:2] == "~/" {
		// attempt to get user information
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		dir := usr.HomeDir
		result = strings.Replace(result, "~/", dir, 1)
	}
	return result, nil
}

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
			if !Exists(destPath) {
				if err := os.MkdirAll(destPath, 0755); err != nil {
					return err
				}
			}
		} else {
			log.Debug("Copying: %s", destPath)
			if err := CopyFileContents(path, destPath); err != nil {
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
