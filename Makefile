# Copyright (c) 2014 Eric Anderton <eric.t.anderton@gmail.com>
# 
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
# 
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
# 
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

VERSION = 0.2
PROGRAM_NAME = grapnel

TESTTARGET := ./foobar
PWD := $(shell pwd)
all: unittest smoketest

# Quick-and-dirty dependency trigger - recompile only if a .go file changes
GOFILES := $(shell find src -type f -name *.go)

clean:
	-rm -f grapnel
	-rm -rf $(TESTTARGET)

# Generates configuration out of data in this Makefile 
emit-config:
	cat src/grapnel/config.tmpl \
	| sed -e 's/%PROGRAM_NAME%/$(PROGRAM_NAME)/' \
	| sed -e 's/%VERSION%/$(VERSION)/' \
	> src/grapnel/config.go

# Normalize 'go test' output to align with 'go build'
go-unittest:
	GOPATH='$(PWD)' go test -v $(TESTPATH) \
	| sed -e 's#	\(.*\).go:#src/$(TESTPATH)/\1.go:#'

# General unittests for each package
unittest:
	make go_unittest TESTPATH=grapnel/flag
	make go_unittest TESTPATH=grapnel/log
	make go_unittest TESTPATH=grapnel/toml
	make go_unittest TESTPATH=grapnel

# Basic command test
smoketest: grapnel
	./grapnel install -c testfiles/smoke.toml -t $(TESTTARGET) -v

# Target command to build
grapnel: $(GOFILES)
	make emit-config
	GOPATH='$(PWD)' go build -o grapnel grapnel/cmd 

.PHONY: all clean emit-config go-unittest smoketest unittest
