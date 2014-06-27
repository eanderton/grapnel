package grapnel

import (
  "net/url"
  "path/filepath"
  "os"
  "io"
  "fmt"
  log "github.com/ngmoco/timber"
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

func (self *Library) AddDependencies(deps... *Dependency) {
  self.Dependencies = append(self.Dependencies, deps...)
}

func (self *Library) Install(installRoot string, ignorePattern string) error {
  // set up root target dir
  importPath := filepath.Join(installRoot, self.Import)
  if err := os.MkdirAll(importPath, 0755); err != nil {
    log.Info("%s", err.Error())
    return log.Error("Could not create target directory: '%s'", importPath)
  }

  // move everything over
  if err := CopyFileTree(importPath, self.TempDir, ignorePattern); err != nil {
    log.Info("%s", err.Error())
    return log.Error("Error while walking dependency file tree")
  }
  return nil
}

func (self *Library) Destroy() error {
  return os.Remove(self.TempDir)
}

func (self *Library) ToToml(writer io.Writer) {
  fmt.Fprintf(writer, "\n[%s]\n", self.Dependency.Name)
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
    if self.Dependency.Tag == "" {
      fmt.Fprintf(writer, "# Pinned to recent tip/head of repository\n")
    }
    fmt.Fprintf(writer, "tag = \"%s\"\n", self.Tag)
  }
}
