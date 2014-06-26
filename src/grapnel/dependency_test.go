package grapnel

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
  if dep.Oper != OpGte ||
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
  if dep, err = NewDependencyFromToml(name, entry); err != nil {
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
