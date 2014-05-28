package grapnel

type Resolver interface {
  MatchDependency(dep *Dependency) bool
  ValidateDependency(dep *Dependency) error
  FetchDependency(dep *Dependency) error
  InstallDependency(dep *Dependency, targetPath string) error
}

