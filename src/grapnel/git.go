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
  "os"
  "path"
  "fmt"
  "strings"
  "regexp"
  "text/template"
  "io/ioutil"
  url "grapnel/url"
  util "grapnel/util"
  log "grapnel/log"
)

var GitRewriteRules = RewriteRuleArray {
  // rewrite rules for misc git resolvers
  SimpleRewriteRule("scheme", `git`,           "type", `git`),
  SimpleRewriteRule("path",   `.*\.git`,       "type", `git`),
  SimpleRewriteRule("import", `github.com/.*`, "type", `git`),
  SimpleRewriteRule("host",   `github.com`,    "type", `git`),

  // rewrite rules for gopkg.in
  &RewriteRule{
    Matches: MatchMap{
      "host": regexp.MustCompile(`gopkg\.in`),
      "path": regexp.MustCompile(`^/[^/]+$`),
    },
    Replacements: ReplaceMap{
      "branch": template.Must(RewriteTemplate(`{{replace .path "^.*\\.(.*)$" "$1"}}`)),
      "path":   template.Must(RewriteTemplate(`{{replace .path "^/(.*)\\..*$" "/go-$1/$1"}}`)),
      "host":   template.Must(RewriteTemplate(`github.com`)),
      "type":   template.Must(RewriteTemplate(`git`)),
    },
  },
  &RewriteRule{
    Matches: MatchMap{
      "host": regexp.MustCompile(`gopkg\.in`),
      "path": regexp.MustCompile(`^.+/.+$`),
    },
    Replacements: ReplaceMap{
      "branch": template.Must(RewriteTemplate(`{{replace .path "^.*\\.(.*)$" "$1"}}`)),
      "path":   template.Must(RewriteTemplate(`{{replace .path "^(.*)\\..*$" "$1"}}`)),
      "host":   template.Must(RewriteTemplate(`github.com`)),
      "type":   template.Must(RewriteTemplate(`git`)),
    },
  },
}

type GitSCM struct{}

func stripGitRepo(baseDir string) {
  os.RemoveAll(path.Join(baseDir,".git"))
}

func (self *GitSCM) Resolve(dep *Dependency) (*Library, error) {
  lib := NewLibrary(dep)

  // fix the tag, and default branch
  if lib.Branch == "" {
    lib.Branch = "master"
  }
  if lib.Tag == "" {
    lib.Tag = "HEAD"
  }

  log.Info("Fetching Git Dependency: '%s'", lib.Import)

  // create a dedicated directory and a context for commands
  tempRoot, err := ioutil.TempDir("","")
  if err != nil {
    return nil, err
  }
  lib.TempDir = tempRoot
  cmd := util.NewRunContext(tempRoot)

  // use the configured url and acquire the depified branch
  log.Info("Fetching remote data for %s", lib.Import)
  if lib.Url == nil {
    // try all supported protocols against a URL composed from the import
    for _, protocol := range []string{"http", "https", "git", "ssh"} {
      packageUrl := protocol + "://" + lib.Import
      log.Warn("Synthesizing url from import: '%s'", packageUrl)
       if err := cmd.Run("git","clone", packageUrl, "-b", lib.Branch, tempRoot); err != nil {
        log.Warn("Failed to fetch: '%s'", packageUrl)
        continue
      }
      lib.Url, _ = url.Parse(packageUrl)  // pin URL
      break
    }
    if err != nil {
      return nil, fmt.Errorf("Cannot download dependency: '%s'", lib.Import)
    }
  } else if err := cmd.Run("git","clone", lib.Url.String(), "-b", lib.Branch, tempRoot); err != nil {
    return nil, fmt.Errorf("Cannot download dependency: '%s'", lib.Url.String())
  }

  // move to the specified commit/tag/hash
  // check out a depific commit - may be a tag, commit hash or HEAD
  if err := cmd.Run("git","checkout", lib.Tag); err != nil {
    return nil, fmt.Errorf("Failed to checkout tag: '%s'", lib.Tag)
  }

  // Pin the Tag to a commit hash if we just have "HEAD" as the 'Tag'
  if lib.Tag == "HEAD" {
    if err := cmd.Run("git","rev-list", "--all", "--max-count=1"); err != nil {
      return nil, fmt.Errorf("Failed to checkout tag: '%s'", lib.Tag)
    } else {
      lib.Tag = strings.TrimSpace(cmd.CombinedOutput)
    }
  }

  // Stop now if we have no semantic version information
  if lib.VersionSpec.IsUnversioned() {
    lib.Version = NewVersion(-1,-1,-1)
    log.Warn("Resolved: %v (unversioned)", lib.Import)
    stripGitRepo(lib.TempDir)
    return lib, nil
  }

  // find latest version match
  if err := cmd.Run("git", "for-each-ref", "refs/tags", "--sort=taggerdate",
      "--format=%(refname:short)"); err != nil {
    return nil, fmt.Errorf("Failed to acquire ref list for depenency")
  } else {
    for _,line := range strings.Split(cmd.CombinedOutput, "\n") {
      log.Debug("%v", line)
      if ver, err := ParseVersion(line); err == nil {
        log.Debug("ver: %v", ver)
        if dep.VersionSpec.IsSatisfiedBy(ver) {
          lib.Tag = line
          lib.Version = ver
          // move to this tag in the history
          if err := cmd.Run("git","checkout", lib.Tag); err != nil {
            return nil, fmt.Errorf("Failed to checkout tag: '%s'", lib.Tag)
          }
          break
        }
      } else {
        log.Debug("Parse git tag err: %v", err)
      }
    }
  }

  // fail if the tag cannot be determined.
  if lib.Version == nil {
    return nil, fmt.Errorf("Cannot find a tag for dependency version specification: %v.", lib.VersionSpec)
  }

  log.Info("Resolved: %s %v", lib.Import, lib.Version)
  stripGitRepo(lib.TempDir)
  return lib, nil
}

func (self *GitSCM) ToDSD(*Library) string {
  return ""
}
