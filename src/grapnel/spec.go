package grapnel

import (
  "net/url"
  "errors"
  "io"
  "strings"
  "fmt"
)

type Spec struct {
  Name string
  Import string
  Url *url.URL
  Type string
  Branch string
  Commit string
  Tag string
}

var validSpecTypes = map[string]bool{
  "git": true, "hg": true, "svn": true,
} 

func getString(config map[string]interface{}, key string) string {
  if value, ok := config[key]; ok {
    return strings.TrimSpace(value.(string))
  }
  return ""
}

func NewSpec(name string, config map[string]interface{}) (*Spec, error) {
  spec := &Spec{
    Name: name,
    Import: getString(config, "import"),
    Type: getString(config, "type"),
    Branch: getString(config, "branch"),
    Commit: getString(config, "commit"),
    Tag: getString(config, "tag"),
  }

  // validate url and import
  urlValue := getString(config, "url")
  if urlValue != "" {
    if url, err := url.Parse(urlValue); err == nil {
      spec.Url = url 
    } else {
      return nil, err
    }
    if spec.Import == "" {
      spec.Import = spec.Url.Host + "/" + spec.Url.Path
    }
  } else if spec.Import == "" {
    return nil, errors.New("Must have an 'import' or 'url'")
  }

  // validate type
  if spec.Type != "" {
    if _, ok := validSpecTypes[spec.Type]; !ok {
        return nil, errors.New("Invalid type: '" + spec.Type + "'")
    }
  }

  return spec, nil
}

// Serializes the specification to a writer in TOML format
func (self *Spec) ToToml(writer io.Writer) {
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
