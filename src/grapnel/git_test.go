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
  . "grapnel/testing"
  "os"
  "testing"
  log "grapnel/log"
)

func TestGitSource(t *testing.T) {
  InitTestLogging()

  // construct a repo
  basePath := BuildTestGitRepo("gitrepo")
  defer os.Remove(basePath)

  // start a daemon to serve the repo
  defer StopGitDaemon()
  if err := StartGitDaemon(basePath); err != nil {
    t.Error("%v", err)
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
  libsrc := &GitSCM{}
  if _, err = libsrc.Resolve(dep); err != nil {
    t.Error("%v", err)
  }
}
