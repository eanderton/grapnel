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
  "net/url"
  "path"
  "path/filepath"
  "os"
  "io"
  "fmt"
  "go/build"
  util "grapnel/util"
  log "grapnel/log"
)

type Library struct {
  *Dependency
  Import string
  Url *url.URL
  Type string
  Branch string
  Tag string
  *Version
  TempDir string
  Dependencies []*Dependency
}

// TODO: lift this URL into a clone() operation
func NewLibrary(dep *Dependency) *Library {
  var newUrl *url.URL
  if dep.Url != nil {
    newUrl = &url.URL {
      Scheme: dep.Url.Scheme,
      Opaque: dep.Url.Opaque,
      User: dep.Url.User,
      Host: dep.Url.Host,
      Path: dep.Url.Path,
      RawQuery: dep.Url.RawQuery,
      Fragment: dep.Url.Fragment,
    }
  }

  return &Library{
    Dependency: dep,
    Import: dep.Import,
    Url: newUrl,
    Type: dep.Type,
    Tag: dep.Tag,
    Dependencies: make([]*Dependency, 0),
  }
}

func (self *Library) Install(installRoot string, ignorePattern string) error {
  // set up root target dir
  importPath := filepath.Join(installRoot, self.Import)
  if err := os.MkdirAll(importPath, 0755); err != nil {
    log.Info("%s", err.Error())
    return fmt.Errorf("Could not create target directory: '%s'", importPath)
  }

  // move everything over
  if err := util.CopyFileTree(importPath, self.TempDir, ignorePattern); err != nil {
    log.Info("%s", err.Error())
    return fmt.Errorf("Error while walking dependency file tree")
  }
  return nil
}

func (self *Library) Destroy() error {
  return os.Remove(self.TempDir)
}

func (self *Library) AddDependencies() error {
  if self.TempDir == "" {
    return nil  // do nothing if there's nothing to search
  }

  // get dependencies via lockfile or grapnelfile
  if deplist, err := LoadGrapnelDepsfile(
    path.Join(self.TempDir, "grapnel-lock.toml"),
    path.Join(self.TempDir, "grapnel.toml")); err != nil {
    return err
  } else if deplist != nil {
    self.Dependencies = append(self.Dependencies, deplist...)
    return nil
  }

  // attempt get dependencies via raw import statements instead
  pkg, err := build.ImportDir(self.TempDir, 0)
  if err != nil {
    log.Debug("Failed to get go imports for %v", err)
    log.Warn("No Go imports to process for %v", self.Import)
    return nil
  }

  // add all non std libs as dependencies of this lib
  for _, importName := range pkg.Imports {
    if IsStandardDependency(importName) {
      log.Debug("Ignoring import: %v", importName)
    } else {
      log.Warn("Adding secondary import: %v", importName)
      dep, err := NewDependency(importName, "", "")
      if err != nil {
        return err
      }
      self.Dependencies = append(self.Dependencies, dep)
    }
  }
  return nil
}

func (self *Library) ToToml(writer io.Writer) {
  fmt.Fprintf(writer, "\n[[dependencies]]\n")
  if self.Version.Major > 0 {
    fmt.Fprintf(writer, "version = \"==%v\"\n", self.Version)
  } else {
    fmt.Fprintf(writer, "# Unversioned\n")
  }
  if self.Type != "" {
    fmt.Fprintf(writer, "type = \"%s\"\n", self.Type)
  }
  if self.Import != "" {
    fmt.Fprintf(writer, "import = \"%s\"\n", self.Import)
  }
  if self.Url != nil {
    fmt.Fprintf(writer, "url = \"%s\"\n", self.Url.String())
  }
  if self.Branch != "" {
    fmt.Fprintf(writer, "branch = \"%s\"\n", self.Branch)
  }
  if self.Tag != "" {
    // TODO: repair notification
    //if self.Dependency.Tag == "" && self.Version.Major == 0 {
    //  fmt.Fprintf(writer, "# Pinned to recent tip/head of repository\n")
    //}
    fmt.Fprintf(writer, "tag = \"%s\"\n", self.Tag)
  }
}
