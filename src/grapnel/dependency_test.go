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
  "testing"
  toml "github.com/pelletier/go-toml"
)

var testDependencyEntry = `
[test_import]
import = "foo/bar/baz"
url = "http://github.com/foo/bar"
type = "test"
branch = "master"
tag = "release-1.0"
version = "1.0.*"
`

func TestNewDepdendency(t *testing.T) {
  dep, err := NewDependency("foo/bar/baz", "http://github.com/foo/bar", ">=1.2.3")
  if err != nil {
    t.Errorf("Error creating Dependency: %v", err)
  }
  if dep.Import != "foo/bar/baz" {
    t.Errorf("Bad value for Import: '%v'. Expected: '%v", dep.Import, "foo/bar/baz")
  }
  if dep.Url.String() != "http://github.com/foo/bar" {
    t.Errorf("Bad value for url: '%v'. Expected: '%v'",
        dep.Url.String(), "http://github.com/foo/bar")
  }
  if dep.VersionSpec.Oper != OpGte ||
     dep.VersionSpec.Major != 1 ||
     dep.VersionSpec.Minor != 2 ||
     dep.VersionSpec.Subminor != 3 {
    t.Errorf("Bad value for version: '%v'. Expected: '%v'",
        dep.VersionSpec.String(), ">=1.2.3")
  }
}

func TestNewDependencyFromToml(t *testing.T) {
  var err error
  var tree *toml.TomlTree
  var dep *Dependency

  if tree, err = toml.Load(testDependencyEntry); err != nil {
    t.Errorf("Error parsing TOML data: %v", err)
  }
  name := "test_import"
  entry := tree.Get(name).(*toml.TomlTree)
  if dep, err = NewDependencyFromToml(entry); err != nil {
    t.Errorf("Error building dependency from TOML: %v", err)
  }
  if dep.Import != "foo/bar/baz" {
    t.Errorf("Bad value for import: '%v'. Expected: '%v'", dep.Import, "foo/bar/baz")
  }
  if dep.Type != "test" {
    t.Errorf("Bad value for test: '%v'. Expected: '%v'", dep.Type, "test")
  }
  if dep.Branch != "master" {
    t.Errorf("Bad value for master: '%v'. Expected: '%v'", dep.Branch, "master")
  }
  if dep.Tag != "release-1.0" {
    t.Errorf("Bad value for tag: '%v'. Expected: '%v'", dep.Tag, "release-1.0")
  }
  if dep.Url.String() != "http://github.com/foo/bar" {
    t.Errorf("Bad value for url: '%v'. Expected: '%v'",
        dep.Url.String(), "http://github.com/foo/bar")
  }
  if dep.VersionSpec.Major != 1 ||
     dep.VersionSpec.Minor != 0 ||
     dep.VersionSpec.Subminor != -1 {
    t.Errorf("Bad value for version: '%v'. Expected: '%v'",
        dep.VersionSpec.String(), "1.0.*")
  }
}
