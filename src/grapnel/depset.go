package grapnel

import (
  "sync"
)

type DepSet struct {
  dict map[string]*Dependency
}

func NewDepSet() *DepSet {
  return &DepSet{ 
    dict: make(map[string]*Dependency),
  }
}

func (self *DepSet) Len() int {
  return len(self.dict)
}

func (self *DepSet) Clear() {
  self.dict = make(map[string]*Dependency)
}

func (self *DepSet) Insert(dep *Dependency) {
  self.dict[dep.Import] = dep
}

func (self *DepSet) Remove(importName string) {
  delete(self.dict, importName)
}

func (self *DepSet) Find(importName string) (*Dependency, bool) {
  name, ok := self.dict[importName]
  return name, ok
}

func (self *DepSet) Each(fn func(*Dependency) bool) {
  for _, dep := range self.dict {
    if !fn(dep) {
      break
    }
  }
}

func (self *DepSet) GoEach(fn func(*Dependency)) {
  var wg sync.WaitGroup
  wg.Add(self.Len())
  for _,dep := range self.dict {
    go func(dep *Dependency) {
      fn(dep)
      wg.Done()
    }(dep)
  }
  wg.Wait()
}

