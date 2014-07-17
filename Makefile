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

VERSION = 0.3
PROGRAM_NAME = grapnel

TESTTARGET := ./foobar
PWD := $(shell pwd)

# Quick-and-dirty dependency trigger - recompile only if a .go file changes
GOFILES := $(shell find src -type f -name *.go)

# Default target - used by vim quickfix and travis-ci
all: unittest smoketest

# Start over
clean:
	-rm -f grapnel
	-rm -rf $(TESTTARGET)
	-rm -f coverage.tmp coverage.out 

# Generates configuration out of data in this Makefile 
emit-config:
	@cat src/grapnel/config.tmpl \
	| sed -e 's/%PROGRAM_NAME%/$(PROGRAM_NAME)/' \
	| sed -e 's/%VERSION%/$(VERSION)/' \
	> src/grapnel/config.go

# Normalize 'go test' output to align with 'go build'
# NOTE: this is a tremendous help for vim's 'quickfix' feature
# Also concatenate coverage reports to 'coverage.out'
go-unittest:
	@GOPATH='$(PWD)' go test -v \
		-coverprofile=coverage.tmp \
		$(TESTPATH) \
	| sed -e 's#	\(.*\).go:#src/$(TESTPATH)/\1.go:#'
	@tail -n +2 coverage.tmp >> coverage.out

# General unittests for each package
unittest:
	# reset coverage data with a single mode line
	-rm coverage.out
	@echo 'mode: set' > coverage.out
	# run unittests and coverage analysis
	make go-unittest TESTPATH=grapnel/flag
	make go-unittest TESTPATH=grapnel/log
	make go-unittest TESTPATH=grapnel/util
	make go-unittest TESTPATH=grapnel
	# generate coverage reports
	@GOPATH='$(PWD)' go tool cover -func=coverage.out

# Interactive coverage report
htmlcover: unittest
	@GOPATH='$(PWD)' go tool cover -html=coverage.out

# Basic command test
smoketest: grapnel
	./grapnel update -p testfiles/smoke.toml -t $(TESTTARGET) --debug

# Target command to build
grapnel: $(GOFILES)
	make emit-config
	GOPATH='$(PWD)' go build -o grapnel grapnel/cmd 

.PHONY: all clean emit-config go-unittest smoketest unittest htmlcover
