
# Dependency Rewrite Rules

Grapnel supports "rewrite rules" to help with situations that can't be handled
directly with Grapnel's default behavior.  In general, they are how we compose
additional _implicit_ rules for handling dependencies. These get added to the 
Grapnel config file, `.grapnelrc`, so that one or more projects might take
advantage of them.

Rewrite rules come in handy under the following situations:

* When a dpendency's repository is moved to another site, directory, or URL
* A dependency on a repo that Grapnel may support but doesn't quite understand
* Dependencies that exist on locally hosted/cached repository mirrors
* Dependencies that need to be replaced by a public or private fork

In all these cases, the typical practice would involve bulk modification of 
the project sourcecode, in order to use a completely different import path.  Most
of the time, this isn't a practical way to download and evaluate new Go programs,
libraries, and tools.

Instead, Grapnel provides support for re-writing a dependency's details, before
Grapnel attempts to download dependencies.

## Rewrite Rules in the Config File

Grapnel Rewrite Rules are placed in `.grapnelrc`.  This is done so that the 
rules may be global across all projects, local to the user, or local to the 
project.

# Composing a Rewrite Rule

A rewrite rule is defined by a `[[rewrite]]` section that contains
`[match]` and `[replace]` subsections.  Within each of those, we define
regular expressions to match on various aspects of dependencies, and how to
re-define the dependency using Go templates.

```
[[rewrite]]
  [match]
    # things to match go here
  [replace]
    # things to replace go here
```

## Valid Dependency Aspects

The following are supported:

* import = The dependency import path (as it would appear in Go code)
* type = The type of dependency
* branch = The branch the dependency is located on
* tag = The tag the dependency is located at
* url = The full URL for the dependency
* scheme = The URL scheme
* host = The URL host
* path = The URL path
* port = The URL port

It should be noted that the URL parts (scheme, host, path, and port) map to the 
'url' aspect, and are provided for ease of use.  When modifying these aspects,
it will also alter the value of the 'url', and vice-versa. 


## Simple Matching

The most simple, and common, rewrite rule, involves a single match expression
and a single replacement:

```
[[rewrite]]
  [match]
    host = `^.*foobar\.com$`
  [replace]
    type = `git`
```

The above maps 'foobar.com' repositories to git.  This is useful if you have 
a local mirror on a corporate Gitlab server, or want to enable some new Git 
hosting service.

## Complex Matching

It's not uncommon to want to match more than one aspect of a dependency
before re-writing it.  This is usually done to assert that only poorly
defined dependencies are modified, and is a best practice.

Let's take the previous example, and make sure we only modify dependencies
that do not have a `type` defined.

```
[[rewrite]]
  [match]
    type = `^$`
    host = `^.*foobar\.com$`
  [replace]
    type = `git`
```

The regular expression `^$` is used to match a completely empty type, and
this rule will only be used if and only if both the `type` and `host` 
expressions above match.


## Complex Replacement

Not only can Rewrite Rules replace multiple aspects on a match, but they
can do so with [Go Templates](https://golang.org/pkg/text/template/).  Each
replacement value is a template, 

All of the Dependency Aspects are available as variables within the template.

In addition, a convenience function, `replace`, is provided for sed-like string
replacement.  It's incredibly useful for working with URL paths.  Overall,
it acts like a proxy for [regex.ReplaceAllString](https://golang.org/pkg/regexp/#Regexp.ReplaceAllString).

```
[[rewrite]]
  [match]
    type = `^$`
    host = `^.*foobar\.com$`
  [replace]
    type = `git`
    host = `{{ replace .host "^.*\\.(.*)$" "$1" }}`
```

The call to `replace`, inside the `host` template, takes three arguments: a
variable to inspect, a regular expression to match, and a replacement expression.
In this case, it strips the leading piece of the `host`, and leaves the 'foobar.com'
part.


# Built-in Rewrite Rules 

Rewrite Rules are not an afterthough or bolt-on feature to Grapnel.  In fact, 
Grapnel is built around this feature, and ships with the following rules
pre-configured:

* A dependency URL that ends with '.git', that has 'github.com' or 'gopkg.in' as 
the host, or has a scheme of 'git://', is considered of type 'git'
* A dependency without an explicit URL is synthesized from the dependency Import 
* A dependency without an explicit Import is synthesized from the dependency URL
* All 'gopkg.in' dependencies are re-mapped to the equivalent 'github.com' settings
* All 'golang.org/x' dependencies are re-mapped to the equivalent 'github.com' settings

Examples of these can be seen in the sourcecode:

* [src/grapnel/lib/git.go](../src/grapnel/lib/git.go#L35)
* [src/grapnel/lib/rewrite.go](../src/grapnel/lib/rewrite.go#L210)
