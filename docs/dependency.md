# Dependency Rules

The core responsibility of Grapnel is to help the developer maintain a manifest
of dependencies from which we can reliably re-create a dependency graph. To do
this, we need to know how to write Dependency Rules

The project `grapnel.toml` file contains one or more `[[dependencies]]` toml 
dictionary array sections.  Each of these contains one or more keys that define
the dependency to download and install (typically in ./src, unless configured
otherwise).

# Valid Dependency Aspects

All of the following are optional, but at least the `import` or `url` must appear
for the entry to represent a valid dependency:

* import = The import as it would appear in Go code
* url = A URL where the import is stored
* version = A semantic version matching expression (more below)
* type = The type of the repository
* branch = A branch within the repository
* tag = A tag within the repository

Each dependency is made up of, at least, information that describes where to
obtain the code for the dependency itself.  In addition, we may provide data
that describes _where in time_ on a project repository we might find the
correct version of that dependency.

# Basic Use: Imitating go get

The most basic expression of a Dependency Rule is to imitate the behavior of
`go get`.  In general, this means downloading the leading commit (`HEAD`) on a 
repository:

```
[[dependencies]]
import = `github.com/pelletier/go-toml`
```

Alternately, we can specify a URL for Grapnel to use instead:

```
[[dependencies]]
url = `https://github.com/pelletier/go-toml`
```

Grapnel is smart enough to use the URL to re-construct what the import path
is supposed to be; these two examples are equivalent.  As a best practice,
do adhere to using import paths, and only use the `url` aspect to override
where Grapnel looks to download a dependency.

In most cases, this is the minimum amount of information necessary to 
allow Grapnel to download your dependencies and generate a lockfile.  However,
both the lockfile and the `grapnel.toml` file will lack metadata vital to
maintenance of the project.  For that we use: Semantic Versioning.


# Using a Version Number

One of the more powerful features of Grapnel is the use of [Semantic Versioning](http://semver.org/)
when describing a dependency.  Grapnel will attempt to match a provided
semantic version expression, with the _latest_ tag or branch within a 
repository that matches.

Granted, not all Go authors publish their software with release version
numbers, but some do!

```
[[dependencies]]
import = `github.com/pelletier/go-toml`
version = "0.2.*"
```

In this example, we specify that `go-toml` should be supplied by any release with
a version number of '0.2.0' or greater.  In practice, most version expressions will
look like this one.  If more leverage is desired, see the sections below:

## Semantic Version Expressions

```
version := [oper] major ['.' minor [ '.' subminor ] ]
```

The version expression is any match operator, followed by a major number, and optional
period-separated minor and subminor numbers.  Asterisks may be used in place of any number,
to match any version number at that level.  The default operator is '=', and the other
relational operators only apply to numbers in the expression: asterisks always match any
version number. Whitespace is allowed in between any expression parts.

### Valid Operators:
* < Match versions less than specified
* <= Match versions less than or equal to specified
* = Match versions equal to specified (default)
* >= Match versions greater than specified
* > Match versions greater than specified

Grapnel's behavior is to match the _latest_ such matching version, in all cases.

### Examples:

```
version = `1.0.0`    # matches exactly version 1.0.0
version = `1.0`      # matches exactly version 1.0 (any subminor)
version = `1.0.*`    # same
version = `=1.0.*`   # same
version = `>=22.2.*` # matches at least version 22.2.0
version = '<10.*.8`  # matches the latest version before 10 with any minor and a subminor of 8
```

# Indicating a Repository Tag

When semantic versioning isn't supported on a dependency's repo, consider indicating
the tag or branch instead.  This way, if a project maintainer is using a different
scheme for organizing versions (like gopkg.in-style projects), you can still specify
how Grapnel should use that kind of information.

Let's take Spf13's Viper library, as an example.  This project has no semantic versioning,
so all we have to go by are commit hashes:

```
[[dependencies]]
import = `github.com/spf13/viper`
tag = `be5ff3e4840cf692388bde7a057595a474ef379e`  # try to use the full hash if you have it

```

This will pin the version to the specified commit hash.



# Advanced: Dissecting the Lockfile

After running `grapnel update`, Grapnel will discover all the intermediate imports
not declared in `grapnel.toml`.  It will then publish them all in the 'lockfile': 
`grapnel-lock.toml`.

It is a best practice to go back through the lockfile and attempt to explicitly
document any undocumented transitive dependencies in `grapnel.toml`.  This way, 
the project doesn't completely depend on the lockfile behavior and the user might
better understand what versions of various dependencies are really important.

In general, any and all of the data in `grapnel-lock.toml` can be copied to
`grapnel.toml`, but you may want to be _less_ specific than the commit-level
pinning that goes on the lockfile.  It is recommended that judicious use be made
of using commit hashes to track dependencies in `grapnel.toml`, and instead, 
provide more semantic information like versions or other release tags.
