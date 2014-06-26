package grapnel

import (
  "os"
  "testing"
  "os/exec"
  "path/filepath"
  "bytes"
  "bufio"
  "strings"
  "time"
  log "github.com/ngmoco/timber"
)

func getTestDependencyData() []*Dependency {
  return nil
}

var gitDaemon *exec.Cmd

func startGitDaemon() (chan struct{}, error) {
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
  basePath, err := filepath.Abs(cwd + "/../../testfiles")
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

func TestGitResolver(t *testing.T) {
  initTestLogging()
  
  defer stopGitDaemon()
  if ready, err := startGitDaemon(); err != nil {
    t.Error("%v", err)
  } else {
    <- ready  // wait until we're ready
  }

  var err error
  var dep *Dependency
  dep, err = NewDependency("foo/bar/baz", "git://localhost:9999/gitrepo", "1.0")
  if err != nil {
    t.Error("%v", err)
  }

  log.Info("version: %v", dep.VersionSpec.String())

//  var lib *Library
  if _, err = GitResolver(dep); err != nil {
    t.Error("%v", err)
  } 
}
