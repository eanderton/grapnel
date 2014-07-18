package grapnel
/*
Copyright (c) 2014 Eric Anderton <eric.t.anderton@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

import (
  "testing"
  "net/url"
  "regexp"
  log "grapnel/log"
)

func getTestPipelineDependencyData() []*Dependency {
  fooUrl,_ := url.Parse("foo://somewhere.com/baz/gorf")
  foobarUrl,_ := url.Parse("http://foobar.com/baz/gorf")

  return []*Dependency{
    &Dependency{
      Import: "dependency1",
      VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
      Type: "test",
    },
    &Dependency{
      Import: "dependency2",
      Url: fooUrl,
      VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
    },
    &Dependency{
      Import: "dependency3",
      Url: foobarUrl,
      VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
    },
  }
}

type testSCM struct{}

func (self *testSCM) Match(*Dependency) bool {
  return false
}

func (self *testSCM) Resolve(dep *Dependency) (*Library, error) {
  lib := &Library{}
  lib.Dependency = *dep
  return lib, nil
}

func (self *testSCM) ToDSD(*Library) string {
  return ""
}

func newTestResolver() *Resolver {
  return &Resolver {
    LibSources: map[string]LibSource {
      "test": &testSCM{},
    },
    MatchRules: []MatchRule {
      {"scheme", regexp.MustCompile(`foo`), []RewriteRule {
        {"type", nil, "test"},
      },},
      {"host", regexp.MustCompile(`foobar\.com`), []RewriteRule {
        {"type", nil, "test"},
      },},
    },
  }
}

func TestResolver(t *testing.T) {
  log.SetGlobalLogLevel(log.DEBUG)

  testResolver := newTestResolver()

  // test using the test data
  testDeps := getTestPipelineDependencyData()
  for ii, dep := range testDeps {
    if _, err := testResolver.Resolve(dep); err != nil {
      t.Errorf("Error resolving dependency %v ('%v'): %v",
        ii, dep.Import, err)
    }
  }
}

func TestDeduplicateDeps(t *testing.T) {
  log.SetGlobalLogLevel(log.DEBUG)

  testResolver := newTestResolver()

  testDeps := getTestPipelineDependencyData()
  testDeps = append(testDeps, testDeps...)
  results, err := testResolver.DeduplicateDeps(testDeps)
  if err != nil {
    t.Errorf("%v", err)
  }
  if len(results) != 3 {
    t.Errorf("Expected length of 3; got %v instead", len(testDeps))
  }
}

func TestLibResolveDeps(t *testing.T) {
  log.SetGlobalLogLevel(log.DEBUG)

  testResolver := newTestResolver()

  resolved := make(map[string]*Library)
  testDeps := getTestPipelineDependencyData()

  // create a pre-resolved entry
  dep := testDeps[0]
  resolved[dep.Import] = &Library {
    Version: &Version{1, 0, 0},
  }
  // test resolution
  if deps, err := testResolver.LibResolveDeps(resolved, testDeps); err != nil {
    t.Errorf("%v", err)
  } else if len(deps) != 2 {
    t.Error("Expected 2 deps: got %v instead", len(deps))
  } else if deps[0].Import != testDeps[1].Import {
    t.Error("Result 0 should be %v not %v", testDeps[1].Import, deps[0].Import)
  } else if deps[1].Import != testDeps[2].Import {
    t.Error("Result 1 should be %v not %v", testDeps[2].Import, deps[1].Import)
  }
}


func TestResolveDependencies(t *testing.T) {
  log.SetGlobalLogLevel(log.DEBUG)

  testResolver := newTestResolver()

  // test using the test data
  testDeps := getTestPipelineDependencyData()
  libs,_ := testResolver.ResolveDependencies(testDeps)
  if len(libs) != len(testDeps) {
    t.Errorf("Error resolving dependencies. Expected %v entries, got %v instead.",
      len(testDeps), len(libs))
  }
}

