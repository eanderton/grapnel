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
	log "grapnel/log"
	"grapnel/util"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var ArchiveRewriteRules = RewriteRuleArray{
	TypeResolverRule("path", `^.*\.zip$`, `archive`),
	TypeResolverRule("path", `^.*\.tar.gz$`, `archive`),
	TypeResolverRule("path", `^.*\.tar$`, `archive`),
}

type ArchiveSCM struct{}

func (self *ArchiveSCM) Resolve(dep *Dependency) (*Library, error) {
	lib := NewLibrary(dep)

	// create a dedicated directory and a context for commands
	tempRoot, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	lib.TempDir = tempRoot
	cmd := util.NewRunContext(tempRoot)

	// prep archive file for write
	filename := filepath.Join(tempRoot, filepath.Base(lib.Dependency.Url.Path))
	file, err := os.OpenFile(filename, os.O_CREATE, 0)
	if err != nil {
		return nil, fmt.Errorf("Cannot open archive for writing: %v", err)
	}
	defer file.Close()

	// get the targeted archive
	response, err := http.Get(lib.Dependency.Url.String())
	if err != nil {
		return nil, fmt.Errorf("Cannot download archive: %v", err)
	}
	if err := response.Write(file); err != nil {
		return nil, fmt.Errorf("Cannot write archive: %v", err)
	}
	file.Close()
	log.Info("Wrote: %s", filename)

	// extract the file
	// TODO: change to using built-in libraries whenever possible.
	switch filepath.Ext(filename) {
	case "zip":
		cmd.Run("unzip", filename)
	case "tar.gz":
		cmd.Run("tar", "xzf", filename)
	case "tar":
		cmd.Run("tar", "xf", filename)
	}
	os.Remove(filename)

	// Stop now if we have no semantic version information
	if lib.VersionSpec.IsUnversioned() {
		lib.Version = NewVersion(-1, -1, -1)
		log.Warn("Resolved: %v (unversioned)", lib.Import)
		return lib, nil
	}

	// get the version number from the filename
	if ver, err := ParseVersion(filepath.Base(filename)); err == nil {
		log.Debug("ver: %v", ver)
		if dep.VersionSpec.IsSatisfiedBy(ver) {
			lib.Version = ver
		}
	} else {
		log.Debug("Parse archive version err: %v", err)
	}

	// fail if the tag cannot be determined.
	if lib.Version == nil {
		return nil, fmt.Errorf("Cannot find a version specification on archive: %v.", lib.VersionSpec)
	}

	log.Info("Resolved: %s %v", lib.Import, lib.Version)
	return lib, nil
}

func (self *ArchiveSCM) ToDSD(*Library) string {
	return ""
}
