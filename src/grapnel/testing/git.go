package testing

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
	"bufio"
	"bytes"
	log "grapnel/log"
	util "grapnel/util"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var gitDaemon *exec.Cmd

func StartGitDaemon(basePath string) error {
	if gitDaemon != nil {
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	log.Info("Using CWD: %v", basePath)
	cmd := exec.Command("git", "daemon",
		"--reuseaddr",
		"--base-path="+basePath,
		"--port=9999",
		"--export-all",
		"--informative-errors",
		"--verbose")
	var outbuf bytes.Buffer
	cmd.Stdout = bufio.NewWriter(&outbuf)
	cmd.Stderr = cmd.Stdout
	cmd.Dir = cwd

	// start the daemon
	if err = cmd.Start(); err != nil {
		return err
	}
	gitDaemon = cmd

	// wait for the daemon to be ready to use
	for {
		data := outbuf.String()
		if strings.Contains(data, "Ready to rumble") {
			<-time.After(1 * time.Second) // wait a second
			os.Stdout.WriteString(data)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			break
		}
	}

	// wait for it to halt asynchronously
	go func() {
		log.Info("git daemon stopped: %v", gitDaemon.Wait())
		gitDaemon = nil
	}()

	return nil
}

func StopGitDaemon() {
	if gitDaemon == nil {
		return
	}
	gitDaemon.Process.Signal(os.Interrupt)
}

func BuildTestGitRepo(repoName string) string {
	var err error
	var basePath string
	if basePath, err = ioutil.TempDir("", ""); err != nil {
		panic(err)
	}
	repoPath := path.Join(basePath, repoName)
	if err = os.Mkdir(repoPath, 0755); err != nil {
		panic(err)
	}

	cmd := util.NewRunContext(repoPath)
	for _, data := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "you@example.com"},
		{"git", "config", "user.name", "Your Name"},
		{"touch", "README"},
		{"git", "add", "README"},
		{"git", "commit", "-a", "-m", "first commit"},
		{"git", "tag", "v1.0"},
		{"touch", "foo.txt"},
		{"git", "add", "foo.txt"},
		{"git", "commit", "-a", "-m", "second commit"},
		{"git", "tag", "v1.1"},
	} {
		cmd.MustRun(data[0], data[1:]...)
	}
	return basePath
}
