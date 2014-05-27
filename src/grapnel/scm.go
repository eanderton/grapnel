package grapnel

type SCM interface {
  MatchDependency(spec *Dependency) bool
  ValidateDependency(spec *Dependency) error
  InstallDependency(spec *Dependency, targetPath string) error
}

