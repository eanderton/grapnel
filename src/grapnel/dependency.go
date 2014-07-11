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
  toml "github.com/pelletier/go-toml"
  "fmt"
  "net/url"
)

type Dependency struct {
  Import string
  Url *url.URL
  Type string
  Branch string
  Tag string  // alased to: commit and revision
  *VersionSpec
}

func NewDependency(importStr string, urlStr string, versionStr string) (*Dependency, error) {
  var err error
  dep := &Dependency {
    Import: importStr,
  }

  if urlStr == "" {
    dep.Url = nil
  } else if dep.Url, err = url.Parse(urlStr); err != nil {
    return nil, err
  } else if dep.Url.Scheme == "" {
    return nil, fmt.Errorf("Url must have a scheme specified.")
  }

  if versionStr == "" {
    dep.VersionSpec = NewVersionSpec(OpEq, -1, -1, -1)
  } else if dep.VersionSpec, err = ParseVersionSpec(versionStr); err != nil {
    return nil, err
  }

  // figure out import from URL if not set
  if dep.Import == "" {
    if dep.Url == nil {
      return nil, fmt.Errorf("Must have an 'import' or 'url' specified")
    } else {
      dep.Import = dep.Url.Host + "/" + dep.Url.Path
    }
  }

  return dep, nil
}

func (self *Dependency) Reconcile(other *Dependency) (*Dependency, error) {
  if self.VersionSpec.Outranks(other.VersionSpec) {
    return self, nil
  } else if other.VersionSpec.Outranks(self.VersionSpec) {
    return other, nil
  }
  return nil, fmt.Errorf("Cannot reconcile dependencies for '%v'", self.Import)
}

func NewDependencyFromToml(tree *toml.TomlTree) (*Dependency, error) {
  var err error = nil
  var dep *Dependency

  dep, err = NewDependency(
    tree.GetDefault("import", "").(string),
    tree.GetDefault("url", "").(string),
    tree.GetDefault("version", "").(string),
  )
  if err != nil {
    return nil, err
  }
  dep.Type = tree.GetDefault("type", "").(string)
  dep.Branch = tree.GetDefault("branch", "").(string)
  dep.Tag = tree.GetDefault("tag", "").(string)

  return dep, nil
}
