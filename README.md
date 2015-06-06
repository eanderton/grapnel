Grapnel
=======
[![Build Status](https://travis-ci.org/eanderton/grapnel.svg)](https://travis-ci.org/eanderton/grapnel)

A dependency management solution for the Go Programming Language.

Grapnel is designed to solve a host of dependency management corner-cases, with a focus on
repeatable builds without need of a vendored code graph.

Grapnel's approach draws heavy inspiration from Bundler, CPAN, SetupUtils, Dub, and Cargo.


Installation
============

Clone this repository and compile the program.

```bash
make all

# or

go install grapnel
```

Copy the `bin/grapnel` binary to the location of your choice, preferably somewhere on the executable path.


How to Use
==========

### 1. Add Dependencies

Add your dependencies  to `grapnel.toml`.  Dependencies support tagging, branching, semantic
versioning, and url overriding, repository type overriding, and more.

```toml
# my grapnel file

[[dependencies]]
import = "gopkg.in/inconshreveable/log15.v2"
version = "2.8"  # install a specific version, not just "latest" v2

[[dependencies]]                                                                                      
import = "gopkg.in/mgo.v2"                                                                            
version = "2015.01.24"  #mgo uses dates for release numbers

[[dependencies]]
import = "github.com/spf13/viper" 

[[dependencies]]
import = "github.com/spf13/cobra"
```

### 2. Update

Run `grapnel update` to update the project dependencies in grapnel.toml.  This 'updates' the 
project with the latest versions of all dependencies that match the specifications within 
`grapnel.toml`.

```bash
$ grapnel update
```

Grapnel will also install any additional dependencies it finds within the cited imports, just like 
`go get`.  If there is a collision or a conflict within the dependency graph, Grapnel will stop
and tell you - it won't write anything to the src directory until it can resolve the entire graph.

In addition to installing a dependency graph, `grapnel update` generates a lockfile: 
`grapnel-lock.toml`.  This file contains the "pinned" state of everything that was installed, 
down to the commit hash for unversioned entries.  This includes any additional dependencies that
were discovered.

```toml
# example lockfile snippet for spf13/cobra - your lockfile may contain many such sections

[[dependencies]]                                                                                      
# Unversioned
type = "git"
import = "github.com/spf13/cobra"
url = "http://github.com/spf13/cobra"
branch = "master"
tag = "f8e1ec56bdd7494d309c69681267859a6bfb7549"
```


### 3. Code and Distribute

Make sure to publish the `grapnel.toml` file, and the `grapnel-lock.toml` file with your project, so other 
users of grapnel can reproduce your build by running `grapnel install` - this will install the exact 
dependency graph cited in the lockfile.. 


### 4. Maintainence

When a new version of a depdendency is available, simply modify the `grapnel.toml` file to
point at the new code.  This may involve settting the `version`, `tag`, or `branch` keys for the
dependency in question.  Then, run `grapnel update` to get the latest code, and follow up with
a unit-test run to make sure the upgrade was successful.


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
pinpoint a specific moment in time, within any given repository's published
commit history.  It also understands semantic versioning in tags on a repo,
so basic version requirements can be applied to any given dependency.


Roadmap
=======

Grapnel will eventually cover support for Subversion, Mercurial, and Bazaar.
At the moment, only git repositories are supported.
