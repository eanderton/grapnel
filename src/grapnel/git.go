package grapnel

import (
  "strings"
//  "path/filepath"
  "net/url"
  "io/ioutil"
  log "github.com/ngmoco/timber"
)

var (
  GitIgnorePattern string = `\.git|\.gitignore`
)

func GitResolver(dep *Dependency) (*Library, error) {
  lib := NewLibrary(dep)

  // fix the type, import, tag, and default branch
  if lib.Type == "" {
    lib.Type = "git"
  }
  if lib.Import == "" {
    lib.Import = dep.Url.Host + dep.Url.Path
  }
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
  cmd := NewRunContext(tempRoot)

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
      return nil, log.Error("Cannot download dependency: '%s'", lib.Import)
    }
  } else if err := cmd.Run("git","clone", lib.Url.String(), "-b", lib.Branch, tempRoot); err != nil {
    return nil, log.Error("Cannot download dependency: '%s'", lib.Url.String())
  }

  // move to the specified commit/tag/hash
  // check out a depific commit - may be a tag, commit hash or HEAD
  if err := cmd.Run("git","checkout", lib.Tag); err != nil {
    return nil, log.Error("Failed to checkout tag: '%s'", lib.Tag)
  }

  // Pin the Tag to a commit hash if we just have "HEAD" as the 'Tag'
  if lib.Tag == "HEAD" {
    if err := cmd.Run("git","rev-list", "--all", "--max-count=1"); err != nil {
      return nil, log.Error("Failed to checkout tag: '%s'", lib.Tag)
    } else {
      lib.Tag = strings.TrimSpace(cmd.CombinedOutput)
    }
  }
 
  // Stop now if we have no semantic version information 
  if lib.VersionSpec.IsUnversioned() {
    log.Warn("Dependency is resolved, but unversioned.")
    lib.Version = NewVersion(-1,-1,-1)
    return lib, nil
  }

  // find latest version match
  if err := cmd.Run("git", "for-each-ref",
      "refs/tags", "--sort=taggerdate", "--format=%(refname:short)"); err != nil {
    return nil, log.Error("Failed to acquire commit hash for dependency")
  } else {
    for _,line := range strings.Split(cmd.CombinedOutput, "\n") {
      log.Info("%v", line)
      if ver, err := ParseVersion(line); err == nil {
        log.Info("ver: %v", ver)
        if dep.VersionSpec.IsSatisfiedBy(ver) {
          lib.Tag = line
          lib.Version = ver
          // move to this tag in the history
          if err := cmd.Run("git","checkout", lib.Tag); err != nil {
            return nil, log.Error("Failed to checkout tag: '%s'", lib.Tag)
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
    return nil, log.Error("Cannot find a tag for dependency version specification: %v.", lib.Dependency.VersionSpec)
  }

  return lib, nil
}



