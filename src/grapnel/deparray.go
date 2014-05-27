/*

*/

package grapnel

import (
  "sync"
)

type DepArray struct {
  array []*Dependency
  mutex sync.RWMutex
}

func NewDepArray() *DepArray {
  return &DepArray{}
}

func (self *DepArray) Clear() {
  self.mutex.Lock()
  defer self.mutex.Unlock()
  self.array = make([]*Dependency, 0)
}

func (self *DepArray) Len() int {
  self.mutex.RLock()
  defer self.mutex.RUnlock()
  return len(self.array)
}

func (self *DepArray) Push(dep *Dependency) {
  self.mutex.Lock()
  defer self.mutex.Unlock()
  self.array = append(self.array, dep)
}

/*
func (self *DepArray) Pull() *Dependency, err {
  if Len(self.array) == 0 {
    return nil, errors.New("Attempted to pull from empty array") 
  }
  self.mutex.Lock()
  defer self.mutex.Unlock()
  result := self.array[0]
  self.array = self.array[1:]
  return result, nil
}
*/

type EachFn func(dep *Dependency) bool

func (self *DepArray) Each(fn EachFn) {
  self.mutex.RLock()
  for ii := 0; ii < len(self.array); ii++ {
    dep := self.array[ii]
    self.mutex.RUnlock()
    if !fn(dep) {
      break
    }
    self.mutex.RLock()
  }
  self.mutex.RUnlock()
}


type GoEachFn func(dep *Dependency)

func (self *DepArray) GoEach(fn GoEachFn) {
  self.Each(func(dep *Dependency) bool {
    go func() {
      fn(dep)
    }()
    return true
  })
}

