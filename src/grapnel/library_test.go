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
	"testing"
)

func TestNewLibrary(t *testing.T) {
	var err error
	var dep *Dependency
	var lib *Library

	dep, err = NewDependency("foo/bar/baz", "http://github.com/foo/bar", ">=1.2.3")
	if err != nil {
		t.Errorf("Error creating Dependency: %v", err)
	}

	lib = NewLibrary(dep)
	if lib.Import != "foo/bar/baz" {
		t.Errorf("Bad value for Import: '%v'. Expected: '%v", lib.Import, "foo/bar/baz")
	}
	if lib.Url.String() != "http://github.com/foo/bar" {
		t.Errorf("Bad value for url: '%v'. Expected: '%v'",
			lib.Url.String(), "http://github.com/foo/bar")
	}
}
