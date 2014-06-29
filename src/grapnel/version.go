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
  "regexp"
  "strconv"
  "math"
  "fmt"
)

type VersionSpec struct {
  Oper int
  Major int
  Minor int
  Subminor int
  minMajor int
  maxMajor int
  minMinor int
  maxMinor int
  minSubminor int
  maxSubminor int
}

type Version struct {
  Major int
  Minor int
  Subminor int
}

const (
  OpEq = iota
  OpLt
  OpLte
  OpGt
  OpGte
)

func (self *VersionSpec) String() string {
  var op string
  switch self.Oper {
  case OpLt:
    op = "<"
  case OpLte:
    op = "<="
  case OpEq:
    op = "="
  case OpGte:
    op = ">="
  case OpGt:
    op = ">"
  }
  var minor string
  if self.Minor == -1 {
    minor = "*"
  } else {
    minor = strconv.Itoa(self.Minor)
  }
  var subminor string
  if self.Subminor == -1 {
    subminor = "*"
  } else {
    subminor = strconv.Itoa(self.Subminor)
  }
  return fmt.Sprintf("%s %v.%v.%v", op, self.Major, minor, subminor)
}

func getMinMax(oper int, value int) (int, int) {
  var min, max int
  if value == -1 {
    min = -1
    max = math.MaxInt32
  } else {
    switch oper {
    case OpLt:
      min = -1
      max = value - 1
    case OpLte:
      min = -1
      max = value
    case OpEq:
      min = value
      max = value
    case OpGte:
      min = value
      max = math.MaxInt32
    case OpGt:
      min = value + 1
      max = math.MaxInt32
    }
  }
  return min,max
}

func NewVersionSpec(oper, major, minor, subminor int) *VersionSpec {
  version := &VersionSpec{
    Oper: oper,
    Major: major,
    Minor: minor,
    Subminor: subminor,
  }
  version.minMajor, version.maxMajor = getMinMax(oper, major)
  version.minMinor, version.maxMinor = getMinMax(oper, minor)
  version.minSubminor, version.maxSubminor = getMinMax(oper, subminor)
  return version
}

func NewVersion(major, minor, subminor int) *Version {
  return &Version{
    Major: major,
    Minor: minor,
    Subminor: subminor,
  }
}

var (
  any = `[^\d]*`
  sp = `\s*`
  opsTok = `(<|<=|=|>=|>)`
  numTok = `(\d+)`
  dotTok = `\.`
  wildNumTok = `(\d+|\*)`
  parseVersionSpec = regexp.MustCompile("^" + 
    sp + opsTok + "?" + sp + numTok + 
    "(" + sp + dotTok + sp + wildNumTok + ")?" +
    "(" + sp + dotTok + sp + wildNumTok + ")?" + 
    sp + "$")
  parseVersion = regexp.MustCompile("^" +
    any + numTok +
    "(" + sp + dotTok + sp + numTok + ")?" +
    "(" + sp + dotTok + sp + numTok + ")?" +
    sp + any + "$")
)

func ParseVersionSpec(src string) (*VersionSpec, error) {
  var oper, major, minor, subminor int
  matches := parseVersionSpec.FindStringSubmatch(src)
  if len(matches) == 0 {
    return nil, fmt.Errorf("Cannot parse version spec: '%s'", src)
  }
  switch matches[1] {
  case "<":
    oper = OpLt
  case "<=":
    oper = OpLte
  case "=":
    oper = OpEq
  case ">=":
    oper = OpGte
  case ">":
    oper = OpGt
  case "":
    oper = OpEq  // default to equals
  }
  major, _ = strconv.Atoi(matches[2])
  if matches[4] != "*" && matches[4] != "" {
    minor, _ = strconv.Atoi(matches[4])
  } else {
    minor = -1
  }
  if matches[6] != "*" && matches[6] != "" {
    subminor, _ = strconv.Atoi(matches[6])
  } else {
    subminor = -1
  }
  return NewVersionSpec(oper, major, minor, subminor), nil
}

func ParseVersion(src string) (*Version, error) {
  var major, minor, subminor int
  matches := parseVersion.FindStringSubmatch(src)
  if len(matches) == 0 {
    return nil, fmt.Errorf("Cannot parse version: '%s'", src)
  }
  major, _ = strconv.Atoi(matches[1])
  if matches[3] != "" {
    minor, _ = strconv.Atoi(matches[3])
  } else {
    minor = -1
  }
  if matches[5] != "" {
    subminor, _ = strconv.Atoi(matches[5])
  } else {
    subminor = -1
  }
  return NewVersion(major, minor, subminor), nil
} 

// Returns true if 'other' is more specific and/or more recent than self.
func (self *VersionSpec) Outranks(other *VersionSpec) bool {
  return self.minMajor >= other.minMajor && self.maxMajor <= other.maxMajor &&
    self.minMinor >= other.minMinor && self.maxMinor <= other.maxMinor &&
    self.minSubminor >= other.minSubminor && self.maxSubminor <= other.maxSubminor
}

// Compares the range of possible valid versions in the spec to a specific version
// Returns true if 'version' satisfies the specification
func (self *VersionSpec) IsSatisfiedBy(version *Version) bool {
  return self.minMajor <= version.Major && self.maxMajor >= version.Major &&
    self.minMinor <= version.Minor && self.maxMinor >= version.Minor &&
    self.minSubminor <= version.Subminor && self.maxSubminor >= version.Subminor
}

func (self *VersionSpec) IsUnversioned() bool {
  return self.Major == -1
}

