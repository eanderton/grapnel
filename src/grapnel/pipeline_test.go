package grapnel

import (
  "testing"
  "net/url"
  log "github.com/ngmoco/timber"
)

func _dummy() {
  log.Info("hello world")
}

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

func TestResolvers(t *testing.T) {
  initTestLogging()

  // set up resolvers and resolver tear-down
  simpleResolver := func(*Dependency) (*Library,error) {
    return &Library{}, nil
  }
  TypeResolvers["test"] = simpleResolver
  UrlSchemeResolvers["foo"] = simpleResolver
  UrlHostResolvers["foobar.com"] = simpleResolver
  defer func(){
    delete(TypeResolvers, "test")
    delete(UrlSchemeResolvers, "test")
    delete(UrlHostResolvers, "test")
  }()

  // test using the test data
  testDeps := getTestPipelineDependencyData()
  for ii, dep := range testDeps {
    if _, err := dep.Resolve(); err != nil {
      t.Errorf("Error resolving dependency %v ('%v'): %v",
        ii, dep.Import, err)
    }
  }
}

func TestDeduplicateDeps(t *testing.T) {
  initTestLogging()

  testDeps := getTestPipelineDependencyData()
  testDeps = append(testDeps, testDeps...)
  results, err := DeduplicateDeps(testDeps)
  if err != nil {
    t.Errorf("%v", err)
  }
  if len(results) != 3 {
    t.Errorf("Expected length of 3; got %v instead", len(testDeps))
  }
}

func TestLibResolveDeps(t *testing.T) {
  initTestLogging()

  resolved := make(map[string]*Library)
  testDeps := getTestPipelineDependencyData()

  // create a pre-resolved entry
  dep := testDeps[0]
  resolved[dep.Import] = &Library {
    Dependency: dep,
    Version: &Version{1, 0, 0},
  }
  // test resolution
  if deps, err := LibResolveDeps(resolved, testDeps); err != nil {
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
  initTestLogging()

  // set up resolvers and resolver tear-down
  simpleResolver := func(dep *Dependency) (*Library,error) {
    return &Library{
      Import: dep.Import,
    }, nil
  }
  TypeResolvers["test"] = simpleResolver
  UrlSchemeResolvers["foo"] = simpleResolver
  UrlHostResolvers["foobar.com"] = simpleResolver
  defer func(){
    delete(TypeResolvers, "test")
    delete(UrlSchemeResolvers, "test")
    delete(UrlHostResolvers, "test")
  }()

  // test using the test data
  testDeps := getTestPipelineDependencyData()
  libs,_ := ResolveDependencies(testDeps)
  if len(libs) != len(testDeps) {
    t.Errorf("Error resolving dependencies. Expected %v entries, got %v instead.",
      len(testDeps), len(libs))
  }
}

