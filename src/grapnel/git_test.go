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
  "testing"
  "os/exec"
  _ "path/filepath"
  "bytes"
  "bufio"
  "strings"
  "time"
  "path"
  "io/ioutil"
  log "github.com/ngmoco/timber"
)

func getTestDependencyData() []*Dependency {
  return nil
}

var gitDaemon *exec.Cmd

func startGitDaemon(basePath string) (chan struct{}, error) {
  ready := make(chan struct{})
  if gitDaemon != nil {
    go func(){
      ready <- struct{}{}
    }()
    return ready, nil
  }
  cwd, err := os.Getwd()
  if err != nil {
    return nil, err
  }
  log.Info("Using CWD: %v", basePath)
  cmd := exec.Command("git", "daemon",
    "--reuseaddr",
    "--base-path=" + basePath,
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
    return nil, err
  }
  gitDaemon = cmd

  // signal that the daemon is ready to use
  go func() {
    for {
      data := outbuf.String()
      if strings.Contains(data, "Ready to rumble") {
        <- time.After(1*time.Second)  // wait a second
        os.Stdout.WriteString(data)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        ready <- struct{}{}
        return
      }
    }
  }()

  // wait for it to halt asynchronously
  go func() {
    log.Info("git daemon stopped: %v", gitDaemon.Wait())
    gitDaemon = nil
  }()

  return ready, nil
}

func stopGitDaemon() {
  if gitDaemon == nil {
    return
  }
  gitDaemon.Process.Signal(os.Interrupt)
}

func BuildTestGitRepo(repoName string) string {
  var err error
  var basePath string
  if basePath, err = ioutil.TempDir("",""); err != nil {
    panic(err)
  }
  repoPath := path.Join(basePath, repoName)
  if err = os.Mkdir(repoPath,0755); err != nil{
    panic(err)
  }

  cmd := NewRunContext(repoPath)
  for _, data := range [][]string {
    {"git", "init"},
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


func TestGitResolver(t *testing.T) {
  initTestLogging()
  
  // construct a repo
  basePath := BuildTestGitRepo("gitrepo")  
  defer os.Remove(basePath)

  // start a daemon to serve the repo
  defer stopGitDaemon()
  if ready, err := startGitDaemon(basePath); err != nil {
    t.Error("%v", err)
  } else {
    <- ready  // wait until we're ready
  }

  // map a dependency to the repo
  var err error
  var dep *Dependency
  dep, err = NewDependency("foo/bar/baz", "git://localhost:9999/gitrepo", "1.0")
  if err != nil {
    t.Error("%v", err)
  }

  log.Info("version: %v", dep.VersionSpec.String())

  // test the resolver
  if _, err = GitResolver(dep); err != nil {
    t.Error("%v", err)
  } 
}
