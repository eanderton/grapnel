package grapnel

import (
  "testing"
  "net/url"
)

func TestMatchDependency(t *testing.T) {
  var urlValue *url.URL
  git := NewGitSCM()

  if !git.MatchDependency(&Dependency{
    Import: "github.com/username/project",
  }) {
    t.Error("Failed supported host: github.com")
  }
  
  if git.MatchDependency(&Dependency{
    Import: "foobar.com/username/project",
  }) {
    t.Error("Failed unsupported host: foobar.com")
  }
  
  if !git.MatchDependency(&Dependency{
    Type: "git",
  }) {
    t.Error("Failed supported type: git")
  }
  
  if git.MatchDependency(&Dependency{
    Type: "foobar",
  }) {
    t.Error("Failed unsupported type: foobar")
  }

  urlValue, _ = url.Parse("git://github.com/username/project")
  if !git.MatchDependency(&Dependency{
    Url: urlValue,
  }) {
    t.Error("Failed git protocol")
  }
  
  urlValue, _ = url.Parse("https://github.com/username/project")
  if !git.MatchDependency(&Dependency{
    Url: urlValue,
  }) {
    t.Error("Failed supported host with protocol: https://github.com")
  }
} 

func TestValidatehDependency(t *testing.T) {
  git := NewGitSCM()

  if err := git.ValidateDependency(&Dependency{
    Tag: "foo",
    Commit: "bar",
  }); err == nil {
    t.Error("Failed disallowing conflicing details: tag + commit")
  }
}
