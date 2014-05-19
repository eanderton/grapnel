package grapnel

import (
  "strings"
  "os"
  "path/filepath"
  "net/url"
  log "github.com/ngmoco/timber"
)


type SCM interface {
  MatchDependencySpec(spec *Spec) bool
  ValidateDependencySpec(spec *Spec) error
  InstallDependency(spec *Spec, targetPath string) error
}

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

// Returns true if the spec is supported by this SCM
func (self *GitSCM) MatchDependencySpec(spec *Spec) bool {
  if spec.Type == "git" {
    return true
  }
  if spec.Url != nil {
    if spec.Url.Scheme == "git" {
      return true
    }
    if _, ok := self.supportedHosts[spec.Url.Host]; ok {
      return true
    }
  } else {
    parts := strings.Split(spec.Import, "/")
    if len(parts) > 0 {
      if _, ok := self.supportedHosts[parts[0]]; ok {
        return true
      }
    }
  }
  return false
}

func (self *GitSCM) ValidateDependencySpec(spec *Spec) error {
  if spec.Tag != "" && spec.Commit != "" {
    return log.Error("Cannot have both a Tag and a Commit specified.")
  }
  return nil
}

func (self *GitSCM) InstallDependency(spec *Spec, targetPath string) error {
  chain := NewFnChain()
  defer chain.Invoke()

  log.Info("Processing %s", spec.Name)

  tempRoot, err := createTempDir(chain)
  if err != nil {
    return err
  }

  // move to the temporary directory and defer returning to that directory 
  if err := pushDir(tempRoot, chain); err != nil {
    return err
  }
  if err := RunCmd("git","init"); err != nil {
    return log.Error("Could not init temporary git repo")
  }

  // use the configured url and acquire the specified branch
  if spec.Branch == "" {
    spec.Branch = "master"
  }
  log.Info("Fetching remote data for %s", spec.Name)
  if spec.Url == nil {
    // try all supported protocols against a URL composed from the import
    for _, protocol := range self.supportedProtocols {
      packageUrl := protocol + "://" + spec.Import
      log.Warn("Synthesizing url from import: '%s'", packageUrl)
       if err := RunCmd("git","fetch", "--tags", packageUrl, spec.Branch); err != nil {
        log.Warn("Failed to fetch: '%s'", packageUrl)
        continue
      }
      spec.Url, _ = url.Parse(packageUrl)  // pin URL
      break
    }
    if err != nil {
      return log.Error("Cannot download dependency: '%s'", spec.Import)
    }
  } else if err := RunCmd("git","fetch", "--tags", spec.Url.String(), spec.Branch); err != nil {
    return log.Error("Cannot download dependency: '%s'", spec.Url.String())
  }

  // advance to head on fetched branch
  if err := RunCmd("git","checkout", "FETCH_HEAD"); err != nil {
    return log.Error("Failed to checkout FETCH_HEAD on branch")
  }
  // optionally check out a specific commit
  if spec.Commit != "" {
    if err := RunCmd("git","checkout", spec.Commit); err != nil {
      return log.Error("Failed to checkout commit: '%s'", spec.Commit)
    }
  } else if spec.Tag != "" {
    if err := RunCmd("git","checkout", spec.Tag); err != nil {
      return log.Error("Failed to checkout tag: '%s'", spec.Tag)
    }
  }

  // claim success here and pin the commit hash
  if data, err := RunCmdOut("git","rev-parse", "HEAD"); err != nil {
    return log.Error("Failed to acquire commit hash for dependency")
  } else {
    spec.Tag = ""
    spec.Commit = strings.TrimSpace(data)
  }
  
  // set up root target dir
  importPath := filepath.Join(targetPath, spec.Import)
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
