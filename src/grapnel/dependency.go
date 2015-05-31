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
	"fmt"
	toml "github.com/pelletier/go-toml"
	so "grapnel/stackoverflow"
	url "grapnel/url"
)

type Dependency struct {
	//TODO: parent dependency for error reporting
	Import      string
	Url         *url.URL
	Type        string
	Branch      string
	Tag         string // alased to: commit and revision
	VersionSpec *VersionSpec
}

func NewDependency(importStr string, urlStr string, versionStr string) (*Dependency, error) {
	var err error
	dep := &Dependency{
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

func (self *Dependency) Flatten() map[string]string {
	results := map[string]string{}
	results["import"] = self.Import
	results["type"] = self.Type
	results["branch"] = self.Branch
	results["tag"] = self.Tag
	if self.Url != nil {
		results["scheme"] = self.Url.Scheme
		results["host"] = self.Url.Host
		results["port"] = self.Url.Port
		results["path"] = self.Url.Path
		results["url"] = self.Url.String()
	} else {
		results["scheme"] = ""
		results["host"] = ""
		results["port"] = ""
		results["path"] = ""
		results["url"] = ""
	}
	return results
}

func (self *Dependency) SetValues(valueMap map[string]string) error {
	for _, key := range []string{"import", "type", "branch", "tag",
		"url", "scheme", "host", "path", "port"} {
		value, ok := valueMap[key]
		if !ok {
			continue // value not in map
		}

		// set the value
		switch key {
		case "import":
			self.Import = value
		case "type":
			self.Type = value
		case "branch":
			self.Branch = value
		case "tag":
			self.Tag = value
		case "url":
			if urlValue, err := url.Parse(value); err != nil {
				return fmt.Errorf("Error setting dependency url: %v", err)
			} else {
				self.Url = urlValue
			}
		}
		if self.Url != nil {
			switch key {
			case "scheme":
				self.Url.Scheme = value
			case "host":
				self.Url.Host = value
			case "path":
				self.Url.Path = value
			case "port":
				self.Url.Port = value
			}
		} else {
			switch key {
			case "scheme":
				self.Url = &url.URL{Scheme: value}
			case "host":
				self.Url = &url.URL{Host: value}
			case "path":
				self.Url = &url.URL{Path: value}
			case "port":
				self.Url = &url.URL{Port: value}
			}
		}
	}
	return nil
}

func (self *Dependency) Reconcile(other *Dependency) (*Dependency, error) {
	if self.VersionSpec.Outranks(other.VersionSpec) {
		return self, nil
	} else if other.VersionSpec.Outranks(self.VersionSpec) {
		return other, nil
	}
	return nil, fmt.Errorf("Cannot reconcile dependencies for '%v'", self.Import)
}

func (self *Dependency) Equal(other *Dependency) bool {
	if self.Import == other.Import &&
		self.Type == other.Type &&
		self.Branch == other.Branch &&
		self.Tag == other.Tag &&
		self.VersionSpec == other.VersionSpec {
		if self.Url != nil && other.Url != nil {
			return self.Url.Equal(other.Url)
		} else {
			return self.Url == nil && other.Url == nil
		}
	}
	return false
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

func loadDependencies(filename string) ([]*Dependency, error) {
	tree, err := toml.LoadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%s %s", filename, err)
	}

	items := tree.Get("dependencies").([]*toml.TomlTree)
	if items == nil {
		return nil, fmt.Errorf("No dependencies to process")
	}

	deplist := make([]*Dependency, 0)
	for idx, item := range items {
		if dep, err := NewDependencyFromToml(item); err != nil {
			return nil, fmt.Errorf("In dependency #%d: %v", idx, err)
		} else {
			deplist = append(deplist, dep)
		}
	}

	return deplist, nil
}

func LoadGrapnelDepsfile(searchFiles ...string) ([]*Dependency, error) {
	for _, filename := range searchFiles {
		if so.Exists(filename) {
			if deplist, err := loadDependencies(filename); err != nil {
				return nil, err
			} else {
				return deplist, nil
			}
		}
	}
	return nil, nil
}
