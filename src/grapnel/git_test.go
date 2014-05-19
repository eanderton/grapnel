package grapnel

import (
  "testing"
)

func TestMatchDependencySpec(t *testing.T) {
  if !gitMatchDependencySpec(Spec{
    Import: "github.com/username/project",
  }) {
    t.Error("Failed supported host: github.com")
  }
  
  if gitMatchDependencySpec(Spec{
    Import: "foobar.com/username/project",
  }) {
    t.Error("Failed unsupported host: foobar.com")
  }
  
  if !gitMatchDependencySpec(Spec{
    Type: "git",
  }) {
    t.Error("Failed supported type: git")
  }
  
  if gitMatchDependencySpec(Spec{
    Type: "foobar",
  }) {
    t.Error("Failed unsupported type: foobar")
  }

  if !gitMatchDependencySpec(Spec{
    Url: "git://github.com/username/project",
  }) {
    t.Error("Failed git protocol")
  }
  
  if !gitMatchDependencySpec(Spec{
    Url: "https://github.com/username/project",
  }) {
    t.Error("Failed supported host with protocol: https://github.com")
  }
} 
