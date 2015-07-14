Grapnel
=======

[![Join the chat at https://gitter.im/eanderton/grapnel](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/eanderton/grapnel?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[Version 0.4](https://github.com/eanderton/grapnel/releases)

[![Build Status](https://travis-ci.org/eanderton/grapnel.svg)](https://travis-ci.org/eanderton/grapnel)

A dependency management solution for the Go Programming Language.

Grapnel is designed to solve a host of dependency management corner-cases, with a focus on
repeatable builds without need of a vendored code graph.

Grapnel's approach draws heavy inspiration from Bundler and Cargo. Features include verision/commit 
pinning, semantic versioning, and rewriting of import handling without touching your code.

For more information about the problems Grapnel solves, see: [Why Use Grapnel?](docs/why.md).

Installation
============

Clone this repository and compile the program.

```bash
make all

# or

go install grapnel
```

Copy the `bin/grapnel` binary to the location of your choice, preferably somewhere on the executable path.


How to Use Grapnel - A Basic Guide
==================================

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

More on writing Dependencies into grapnel.toml [here](docs/dependency.md)

More on TOML syntax [here](https://github.com/toml-lang/toml/tree/4f9760fe0ad59163194b837d0b31fcf08323bef3).

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
dependency graph cited in the lockfile.


### 4. Maintainence

When a new version of a depdendency is available, simply modify the `grapnel.toml` file to
point at the new code.  This may involve settting the `version`, `tag`, or `branch` keys for the
dependency in question.  Then, run `grapnel update` to get the latest code, and follow up with
a unit-test run to make sure the upgrade was successful.


### 6. Feedback

Grapnel is a work in progress.  If you have any ideas, suggestions, or complaints,
please feel free to [file an issue](https://github.com/eanderton/grapnel/issues)!
No software is perfect, but together, we can make Grapnel more perfect than
when you found it.  


Advanced Use
============

### 1. The Configuration File

Grapnel doesn't need a config file, but it will complain if you don't have one.  At
the minimum, add an empty `.grapnelrc` file to your system:

`touch /etc/.grapnelrc`

Grapnel looks in the following locations for `.grapnelrc`, in order:

* ./.grapnelrc     # project-level configuration
* ~/.grapnelrc     # user-level configuration
* /etc/.grapnelrc  # system-level configuration

In general it is a best practice to create a project local .grapnelrc, and then migrate 
the contents out to `/etc/.grapnelrc` when the contents are made final.

The .grapnelrc should also be added to your .gitignore file, or equivalent for your 
repository type.  In general, the contents of the file aren't meant for distribution
(like rewrite rules).  If you desire to share this information, consider publishing a
`grapnelrc` file, or placing the contents in your project documentation instead.

### 2. Dependency Rewrite Rules

Sometimes you need more leverage than what Grapnel gives you out of the box. For that
Grapnel supports [Dependency Rewrite Rules](docs/rewrite.md) which 
can come in handy under the following situations:

* When a repository is moved to another site, directory, or URL
* A dependency on a repo that Grapnel may support but doesn't quite understand
* Locally hosted/cached repository mirrors
* A public or private fork of a well-used library

More about rewrite rules [here](docs/rewrite.md).


Roadmap
=======

Grapnel will eventually cover support for Subversion, Mercurial, and Bazaar.
At the moment, only git repositories are supported.
