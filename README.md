Grapnel
=======

A dependency management solution for the Go Programming Language.  At the moment
only the basic 'install' mode is supported for github repositories.  Support for
other SCM types, and other major repo sites, is planned.

```
$ ./grapnel 

Usage: 
  grapnel [command]

Available Commands: 
  install                   :: Ensure that dependencies are installed and ready for use.
  help [command]            :: Help about any command

 Available Flags:
  -c, --config="./toml": configuration file
  -q, --quiet: quiet output
  -t, --target="./src": where to manage packages
  -v, --verbose: verbose output

Use "grapnel help [command]" for more information about that command.
```

About
=====

Grapnel is a work in progress. Release v0.1 features a very rudimentary take on
how the tool is designed to configure dependencies from outside your source.

Grapnel's approach is modeled on Ruby's Bundler, with some inspiration from
CPAN, SetupUtils, Dub, and Crate.

Motivation
==========

The problem is that 'go get' does a fantastic job of bootstrapping basic 
projects from live repos, but it falls short in the following circumstances:

* A bug is introduced into a library you depend on
* A library author redesigns their API
* The repo is not tagged as 'go1', and is under active development (not stable)
* A library depends on an older release of another library

In all of these cases, 'go get's default behavior is to get the tip of the 
repository.  This means that your project may be composed of unstable code the
instant 'go get' is done with its installation pass.

A bigger problem is that the process is not repatable, because there's no way
to inform go of specifically which version, branch, tag, or commit to install.

This would encourage developers to manually resolve the issue in any number 
of ways, and/or freeze a reliable dependency graph in a repository somewhere.
This establishes an artificial barrier to tracking releases and other updates
that may be security-related in nature.  An automated tool can more readily
solve the problem of upgrades, with less error, and in less time.

Finally, 'go get' has no understanding of semantic versioning.  So even the
best maintained repo can only pin 'go1' on one release in particular as
their annointed version to use.  It provides no way to freeze a moment in
time where a particular dependency graph was proven to be stable, and is 
at the whim of project mantainers, if and when they tag a paricular changeset
as 'go1'.

Grapnel attempts to solve all of these problems, by providing a way to
"dial in" a specific point in time, within any given repository's published
commit history.  It also understands semantic versioning in tags on a repo,
so basic version requirements can be applied to any given dependency.

Roadmap
=======

Grapnel will eventually cover support for Subversion, Mercurial, and Bazaar.

Support for web proxies, local repository mirrors, artifactories, recursive
dependency reconciliation, are all planned.  Please be patient while we
add the features you're wating for.
