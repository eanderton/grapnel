package grapnel

import (
  "testing"
)

func TestLen(t *testing.T) {
  obj := NewDepArray()
  if obj.Len() != 0 {
    t.Error("New DepArray is not of length 0:", obj.Len())
  }

  dep := &Dependency{} 
  obj.Push(dep)
  if obj.Len() != 1 {
    t.Error("New DepArray is not of length 1:", obj.Len())
  }
}
