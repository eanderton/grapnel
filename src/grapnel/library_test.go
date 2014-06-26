package grapnel

import (
  "testing"
)

func TestNewLibrary(t *testing.T) {
  var err error
  var dep *Dependency
  var lib *Library

  dep, err = NewDependency("foo/bar/baz", "http://github.com/foo/bar", ">=1.2.3")
  if err != nil {
    t.Errorf("Error creating Dependency: %v", err)
  }

  lib = NewLibrary(dep)
  if lib.Import != "foo/bar/baz" {
    t.Errorf("Bad value for Import: '%v'. Expected: '%v", lib.Import, "foo/bar/baz")
  }
  if lib.Url.String() != "http://github.com/foo/bar" {
    t.Errorf("Bad value for url: '%v'. Expected: '%v'",
        lib.Url.String(), "http://github.com/foo/bar")
  }
}
