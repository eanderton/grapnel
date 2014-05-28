package grapnel

import (
  "net/url"
  "errors"
  "io"
  "os"
  "strings"
  "fmt"
)

type Dependency struct {
  Name string
  Import string
  Url *url.URL        `toml:"-"`
  RawUrl string       `toml:"url"`
  Type string
  Branch string
  Commit string
  Tag string
  Resolver Resolver   `toml:"-"`
  TempRoot string     `toml:"-"`
}

func getString(config map[string]interface{}, key string) string {
  if value, ok := config[key]; ok {
    return strings.TrimSpace(value.(string))
  }
  return ""
}

func (self *Dependency) Destroy() {
  if self.TempRoot != "" {
    os.RemoveAll(self.TempRoot)
    self.TempRoot = ""
  }
}

func (self *Dependency) Init() error {
  // validate url and import
  if self.RawUrl != "" {
    if url, err := url.Parse(self.RawUrl); err == nil {
      self.Url = url 
    } else {
      return err
    }
    if self.Import == "" {
      self.Import = self.Url.Host + "/" + self.Url.Path
    }
  } else if self.Import == "" {
    return errors.New("Must have an 'import' or 'url'")
  }
  return nil
}

// Serializes the specification to a writer in TOML format
func (self *Dependency) ToToml(writer io.Writer) {
  // don't write a dependency entry for standard imports
  if self.Type == "std" {
    return
  }
  fmt.Fprintf(writer, "\n[deps.%s]\n", self.Name)
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
  if self.Commit != "" {
    fmt.Fprintf(writer, "commit = \"%s\"\n", self.Commit)
  }
}
