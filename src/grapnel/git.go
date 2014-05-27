package grapnel

import (
  "strings"
  "os"
  "path/filepath"
  "net/url"
  "io/ioutil"
  log "github.com/ngmoco/timber"
)

type GitSCM struct {
  supportedHosts map[string]string
  supportedProtocols []string
}

func NewGitSCM() *GitSCM {
  self := &GitSCM{}

  // TODO: configure from config file
  self.supportedHosts = map[string]string {
    "github.com": "github.com",
  }
  self.supportedProtocols = []string {
    "https", "http", "git",
  }

  return self
}

// Returns true if the dep is supported by this SCM
func (self *GitSCM) MatchDependency(dep *Dependency) bool {
  if dep.Type == "git" {
    return true
  }
  if dep.Url != nil {
    if dep.Url.Scheme == "git" {
      return true
    }
    if _, ok := self.supportedHosts[dep.Url.Host]; ok {
      return true
    }
  } else {
    parts := strings.Split(dep.Import, "/")
    if len(parts) > 0 {
      if _, ok := self.supportedHosts[parts[0]]; ok {
        return true
      }
    }
  }
  return false
}

func (self *GitSCM) ValidateDependency(dep *Dependency) error {
  if dep.Tag != "" && dep.Commit != "" {
    return log.Error("Cannot have both a Tag and a Commit depified.")
  }
  return nil
}

func (self *GitSCM) InstallDependency(dep *Dependency, targetPath string) error {
  log.Info("Processing %s", dep.Name)

  // create a dedicated directory and a context for commands
  tempRoot, err := ioutil.TempDir("","")
  if err != nil {
    return err
  }
  defer os.RemoveAll(tempRoot) 
  cmd := NewRunContext(tempRoot)

  if err := cmd.Run("git","init"); err != nil {
    return log.Error("Could not init temporary git repo")
  }

  // use the configured url and acquire the depified branch
  if dep.Branch == "" {
    dep.Branch = "master"
  }
  log.Info("Fetching remote data for %s", dep.Name)
  if dep.Url == nil {
    // try all supported protocols against a URL composed from the import
    for _, protocol := range self.supportedProtocols {
      packageUrl := protocol + "://" + dep.Import
      log.Warn("Synthesizing url from import: '%s'", packageUrl)
       if err := cmd.Run("git","fetch", "--tags", packageUrl, dep.Branch); err != nil {
        log.Warn("Failed to fetch: '%s'", packageUrl)
        continue
      }
      dep.Url, _ = url.Parse(packageUrl)  // pin URL
      break
    }
    if err != nil {
      return log.Error("Cannot download dependency: '%s'", dep.Import)
    }
  } else if err := cmd.Run("git","fetch", "--tags", dep.Url.String(), dep.Branch); err != nil {
    return log.Error("Cannot download dependency: '%s'", dep.Url.String())
  }

  // advance to head on fetched branch
  if err := cmd.Run("git","checkout", "FETCH_HEAD"); err != nil {
    return log.Error("Failed to checkout FETCH_HEAD on branch")
  }
  // optionally check out a depific commit
  if dep.Commit != "" {
    if err := cmd.Run("git","checkout", dep.Commit); err != nil {
      return log.Error("Failed to checkout commit: '%s'", dep.Commit)
    }
  } else if dep.Tag != "" {
    if err := cmd.Run("git","checkout", dep.Tag); err != nil {
      return log.Error("Failed to checkout tag: '%s'", dep.Tag)
    }
  }

  // claim success here and pin the commit hash
  if err := cmd.Run("git","rev-parse", "HEAD"); err != nil {
    return log.Error("Failed to acquire commit hash for dependency")
  } else {
    dep.Tag = ""
    dep.Commit = strings.TrimSpace(cmd.CombinedOutput)
  }
  
  // set up root target dir
  importPath := filepath.Join(targetPath, dep.Import)
  if err := os.MkdirAll(importPath, 0755); err != nil {
    log.Info("%s", err.Error())
    return log.Error("Could not create target directory: '%s'", importPath)
  }
  if err := CopyFileTree(importPath, tempRoot, `\.git|\.gitignore`); err != nil {
    log.Info("%s", err.Error())
    return log.Error("Error while walking dependency file tree")
  }
  return nil
}
