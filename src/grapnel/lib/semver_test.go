package lib

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
	log "grapnel/log"
	"testing"
)

func TestParseVersionSpec(t *testing.T) {
	log.SetGlobalLogLevel(log.DEBUG)

	// positive tests
	for k, v := range map[string]*VersionSpec{
		">1.0":         NewVersionSpec(OpGt, 1, 0, -1),
		"=1.1":         NewVersionSpec(OpEq, 1, 1, -1),
		"<1.1.1":       NewVersionSpec(OpLt, 1, 1, 1),
		" <= 1.*.* ":   NewVersionSpec(OpLte, 1, -1, -1),
		" >= 100.0.* ": NewVersionSpec(OpGte, 100, 0, -1),
		" 5 ":          NewVersionSpec(OpEq, 5, -1, -1),
	} {
		version, err := ParseVersionSpec(k)
		if err != nil {
			t.Errorf("Error parsing version: '%v': %v", k, err)
		}
		if v.Oper != version.Oper {
			t.Errorf("Operators don't match for '%v': %v vs %v", k, v.Oper, version.Oper)
		}
		if v.Major != version.Major {
			t.Errorf("Major values don't match for '%v': %v vs %v", k, v.Major, version.Major)
		}
		if v.Minor != version.Minor {
			t.Errorf("Minor values don't match for '%v': %v vs %v", k, v.Minor, version.Minor)
		}
		if v.Subminor != version.Subminor {
			t.Errorf("Subminor values don't match for '%v': %v vs %v", k, v.Subminor, version.Subminor)
		}
	}

	// negative tests
	for _, item := range []string{
		"v1.0", "1.0xyz", "1.1.1.1",
	} {
		if _, err := ParseVersionSpec(item); err == nil {
			t.Errorf("Bad version parsed okay: %v", item)
		}
	}
}

func TestParseVersion(t *testing.T) {
	log.SetGlobalLogLevel(log.DEBUG)

	// positive tests
	for k, v := range map[string]*Version{
		"1.0":            NewVersion(1, 0, -1),
		"1.1":            NewVersion(1, 1, -1),
		"1.1.1":          NewVersion(1, 1, 1),
		" v1.5 ":         NewVersion(1, 5, -1),
		" release100.0 ": NewVersion(100, 0, -1),
		"release.r60 ":   NewVersion(60, -1, -1),
		"r2015.01.24":    NewVersion(2015, 1, 24),
		" 5 ":            NewVersion(5, -1, -1),
	} {
		version, err := ParseVersion(k)
		if err != nil {
			t.Errorf("Error parsing version: '%v': %v", k, err)
		}
		if v.Major != version.Major {
			t.Errorf("Major values don't match for '%v': %v vs %v", k, v.Major, version.Major)
		}
		if v.Minor != version.Minor {
			t.Errorf("Minor values don't match for '%v': %v vs %v", k, v.Minor, version.Minor)
		}
		if v.Subminor != version.Subminor {
			t.Errorf("Subminor values don't match for '%v': %v vs %v", k, v.Subminor, version.Subminor)
		}
	}

	// negative tests
	for _, item := range []string{
		"1.1.1.1",
		"7dbad25113954256a925a5a1f7348b92f196b295",
	} {
		if _, err := ParseVersion(item); err == nil {
			t.Errorf("Bad version parsed okay: %v", item)
		}
	}
}

type vsRankTest struct {
	A, B             string
	ResultA, ResultB bool
}

func TestVersionSpecRank(t *testing.T) {
	log.SetGlobalLogLevel(log.DEBUG)

	for _, item := range []vsRankTest{
		vsRankTest{"=1", "=5", false, false},
		vsRankTest{">1", "=5", false, true},
		vsRankTest{"<7", "=5", false, true},
		vsRankTest{"=1", "=1", true, true},
		vsRankTest{">6", "<4", false, false},
		vsRankTest{">2", ">4", false, true},
		vsRankTest{">2.0", ">4.0", false, true},
		vsRankTest{">2.0.0", ">4.0.0", false, true},
	} {
		var err error
		var vsA, vsB *VersionSpec
		if vsA, err = ParseVersionSpec(item.A); err != nil {
			t.Errorf("Error parsing version spec: '%v': %v", item.A, err)
		}
		if vsB, err = ParseVersionSpec(item.B); err != nil {
			t.Errorf("Error parsing version spec: '%v': %v", item.B, err)
		}
		if vsA.Outranks(vsB) != item.ResultA {
			t.Errorf("'%v' outrank '%v' == %v, expected %v", item.A, item.B, !item.ResultA, item.ResultA)
		}
		if vsB.Outranks(vsA) != item.ResultB {
			t.Errorf("'%v' outrank '%v' == %v, expected %v", item.B, item.A, !item.ResultB, item.ResultB)
		}
	}
}

type vsSatisfyTest struct {
	Vspec, Ver string
	Result     bool
}

func TestVersionSatisfaction(t *testing.T) {
	log.SetGlobalLogLevel(log.DEBUG)

	for _, item := range []vsSatisfyTest{
		vsSatisfyTest{"=1", "5", false},
		vsSatisfyTest{">7", "5", false},
		vsSatisfyTest{">1", "5", true},
		vsSatisfyTest{"<6", "2", true},
		vsSatisfyTest{">=2", "2", true},
		vsSatisfyTest{">=1.*.*", "2", true},
		vsSatisfyTest{">=1.0.*", "2.1", true},
		vsSatisfyTest{">=1.0.*", "2.1.10", true},
		vsSatisfyTest{"=1.0.*", "1.0.10", true},
		vsSatisfyTest{"=1.0.*", "1.0", true},
		vsSatisfyTest{"1.0", "1.0", true},
		vsSatisfyTest{"1.0", "v1.0", true},
	} {
		var err error
		var vs *VersionSpec
		var ver *Version
		if vs, err = ParseVersionSpec(item.Vspec); err != nil {
			t.Errorf("Error parsing version spec: '%v': %v", item.Vspec, err)
		}
		if ver, err = ParseVersion(item.Ver); err != nil {
			t.Errorf("Error parsing version: '%v': %v", item.Ver, err)
		}
		if vs.IsSatisfiedBy(ver) != item.Result {
			t.Errorf("'%v' satisfied by '%v' == %v, expected %v",
				item.Vspec, item.Ver, !item.Result, item.Result)
		}
	}
}
