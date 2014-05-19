package grapnel

import (
  "testing"
  "net/url"
)

func TestMatchDependencySpec(t *testing.T) {
  var urlValue *url.URL
  git := NewGitSCM()

  if !git.MatchDependencySpec(&Spec{
    Import: "github.com/username/project",
  }) {
    t.Error("Failed supported host: github.com")
  }
  
  if git.MatchDependencySpec(&Spec{
    Import: "foobar.com/username/project",
  }) {
    t.Error("Failed unsupported host: foobar.com")
  }
  
  if !git.MatchDependencySpec(&Spec{
    Type: "git",
  }) {
    t.Error("Failed supported type: git")
  }
  
  if git.MatchDependencySpec(&Spec{
    Type: "foobar",
  }) {
    t.Error("Failed unsupported type: foobar")
  }

  urlValue, _ = url.Parse("git://github.com/username/project")
  if !git.MatchDependencySpec(&Spec{
    Url: urlValue,
  }) {
    t.Error("Failed git protocol")
  }
  
  urlValue, _ = url.Parse("https://github.com/username/project")
  if !git.MatchDependencySpec(&Spec{
    Url: urlValue,
  }) {
    t.Error("Failed supported host with protocol: https://github.com")
  }
} 

func TestValidatehDependencySpec(t *testing.T) {
  git := NewGitSCM()

  if err := git.ValidateDependencySpec(&Spec{
    Tag: "foo",
    Commit: "bar",
  }); err == nil {
    t.Error("Failed disallowing conflicing details: tag + commit")
  }
}
