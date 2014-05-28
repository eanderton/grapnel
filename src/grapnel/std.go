package grapnel

import (
  "go/build"
  _ "path/filepath"
  log "github.com/ngmoco/timber"
)

type StdResolver struct {
  context build.Context
}

func NewStdResolver() *StdResolver {
  self := &StdResolver{
    context: build.Default,
  }
  // reduce resolution paths to point only at root
  self.context.GOPATH = self.context.GOROOT

  return self
}

func (self *StdResolver) MatchDependency(dep *Dependency) bool {
  if dep.Type == "std" {
    return true
  }
  if dep.Import == "" || dep.Url != nil {
    return false
  }
  // attempt to resolve the location of the import
  if _, err := self.context.Import(dep.Import, "", build.FindOnly); err == nil {
    return true
  }
  return false
}

func (self *StdResolver) ValidateDependency(dep *Dependency) error {
  return nil
}

func (self *StdResolver) FetchDependency(dep *Dependency) error {
  dep.Type = "std"
  if _, err := self.context.Import(dep.Import, "", 0); err != nil {
    log.Error("", err)
    return log.Error("Cannot load standard import: '%s'", dep.Import)
  }
  return nil
}

func (self *StdResolver) InstallDependency(dep *Dependency, targetPath string) error {
  // do nothing
  return nil
}
