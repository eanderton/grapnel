package grapnel

import (
  toml "github.com/pelletier/go-toml"
  "fmt"
  "net/url"
  "io"
  log "github.com/ngmoco/timber"
)

type Dependency struct {
  Name string
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
  return nil, log.Error("Cannot reconcile dependencies for '%v'", self.Import)
}

func NewDependencyFromToml(name string, tree *toml.TomlTree) (*Dependency, error) {
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
  dep.Name = name
  dep.Type = tree.GetDefault("type", "").(string)
  dep.Branch = tree.GetDefault("branch", "").(string)
  dep.Tag = tree.GetDefault("tag", "").(string)

  return dep, nil
}

func (self *Dependency) ToToml(writer io.Writer) { 
  fmt.Fprintf(writer, "\n[deps.%s]\n", self.Name)
  fmt.Fprintf(writer, "version = \"%v\"\n", self.VersionSpec)
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
    fmt.Fprintf(writer, "tag = \"%s\"\n", self.Tag)
  }
}

