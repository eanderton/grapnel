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
	log "grapnel/log"
	url "grapnel/url"
	"testing"
)

type testSCM struct{}

func (self *testSCM) Resolve(dep *Dependency) (*Library, error) {
	lib := &Library{}
	lib.Dependency = *dep
	return lib, nil
}

func (self *testSCM) ToDSD(*Library) string {
	return ""
}

func TestResolver(t *testing.T) {
	log.SetGlobalLogLevel(log.DEBUG)

	// construct a test for a basic resolver for type 'test'
	resolver := &Resolver{
		LibSources: map[string]LibSource{
			"test": &testSCM{},
		},
	}

	// positive tests
	for ii, dep := range []*Dependency{
		&Dependency{
			Import:      "dependency1",
			VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
			Type:        "test",
		},
	} {
		if _, err := resolver.Resolve(dep); err != nil {
			t.Errorf("Error resolving dependency %v: %v", ii, err)
			t.Log("Dep: ", dep.Flatten())
		}
	}

	// negative tests
	for ii, dep := range []*Dependency{
		&Dependency{
			Import:      "dependency2",
			Url:         url.MustParse("foo://somewhere.com/baz/gorf"),
			VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
		},
	} {
		if _, err := resolver.Resolve(dep); err == nil {
			t.Errorf("Error ignoring dependency %v", ii)
			t.Log("Dep: ", dep.Flatten())
		}
	}
}

func TestDeduplicateDeps(t *testing.T) {
	log.SetGlobalLogLevel(log.DEBUG)

	// construct a test for a basic resolver for type 'test'
	resolver := &Resolver{
		LibSources: map[string]LibSource{
			"test": &testSCM{},
		},
	}

	// dependency set for test
	testDeps := []*Dependency{
		&Dependency{
			Import:      "dependency1",
			VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
			Type:        "test",
		},
		&Dependency{
			Import:      "dependency2",
			Url:         url.MustParse("foo://somewhere.com/baz/gorf"),
			VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
		},
		&Dependency{
			Import:      "dependency3",
			Url:         url.MustParse("http://foobar.com/baz/gorf"),
			VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
		},
	}
	targetLen := len(testDeps)

	// ... duplicated
	testDeps = append(testDeps, testDeps...)

	// positive test
	results, err := resolver.DeduplicateDeps(testDeps)
	if err != nil {
		t.Errorf("%v", err)
	}
	if len(results) != targetLen {
		t.Errorf("Expected length of %v; got %v instead", targetLen, len(results))
	}
}

func TestLibResolveDeps(t *testing.T) {
	log.SetGlobalLogLevel(log.DEBUG)

	// construct a test for a basic resolver for type 'test'
	resolver := &Resolver{
		LibSources: map[string]LibSource{
			"test": &testSCM{},
		},
	}

	// dependency set for test
	testDeps := []*Dependency{
		&Dependency{
			Import:      "dependency1",
			VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
			Type:        "test",
		},
		&Dependency{
			Import:      "dependency2",
			Url:         url.MustParse("foo://somewhere.com/baz/gorf"),
			VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
		},
		&Dependency{
			Import:      "dependency3",
			Url:         url.MustParse("http://foobar.com/baz/gorf"),
			VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
		},
	}

	// target library set
	resolved := make(map[string]*Library)

	// create a pre-resolved entry
	dep := testDeps[0]
	resolved[dep.Import] = &Library{
		Version: &Version{1, 0, 0},
	}
	// test resolution
	if deps, err := resolver.LibResolveDeps(resolved, testDeps); err != nil {
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

	// construct a test for a basic resolver for type 'test'
	resolver := &Resolver{
		LibSources: map[string]LibSource{
			"test": &testSCM{},
		},
	}

	// dependency set for test
	testDeps := []*Dependency{
		&Dependency{
			Import:      "dependency1",
			VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
			Type:        "test",
		},
		&Dependency{
			Import:      "dependency2",
			VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
			Type:        "test",
		},
		&Dependency{
			Import:      "dependency3",
			VersionSpec: NewVersionSpec(OpEq, 1, 0, -1),
			Type:        "test",
		},
	}

	// test using the test data
	libs, err := resolver.ResolveDependencies(testDeps)
	if err != nil {
		t.Errorf("Error resolving dependencies: %v", err)
	}
	if len(libs) != len(testDeps) {
		t.Errorf("Error resolving dependencies. Expected %v entries, got %v instead.",
			len(testDeps), len(libs))
	}
}
