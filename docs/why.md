Why Use Grapnel?
================

The problem is that 'go get' does a fantastic job of bootstrapping basic 
projects from live repos, but it falls short in the following circumstances:

* A bug is introduced into a library you depend on
* A library author redesigns their API
* The repo is not tagged as 'go1', and is under active development (not stable)
* A library depends on an older release of another library

In all of these cases, 'go get's default behavior is to get the tip of the 
repository.  This means that your project may be composed of unstable code the
instant 'go get' is done with its installation pass.

A bigger problem is that the process is not repeatable, because there's no way
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

Vendoring
=========

Vendoring is one way to solve the problem of engineering a solution around
a specific set of versioned dependencies. However, it comes with some 
consequences that are difficult to work around:

* You are stuck with the author's dependency graph, for better or worse
* There is no manifest that declares when the vendored graph was procured,
or what version/release they are
* There is no good way to modify the graph to navigate around a bug or
insecure feature

Grapnel proposes a different strategy: feel free to vendor libraries for 
ease-of-use, but use a tool to declare your graph for automation.  Grapnel
uses this very strategy for the [go-toml](https://github.com/pelletier/go-toml)
library dependency.  If at some point there is a bug or a problem with the 
go-toml library, this dependency can be moved forwards or backwards in time,
as needed.
