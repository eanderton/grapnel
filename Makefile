# Dead-simple makefile for Grapnel

TESTTARGET := ./foobar
PWD := $(shell pwd)
all: grapnel

clean:
	-rm -f grapnel
	-rm -rf $(TESTTARGET)

grapnel:
	GOPATH='$(PWD)'; go build -o grapnel grapnel/cmd 

unittest:
	GOPATH='$(PWD)'; go test grapnel

smoketest: grapnel
	./grapnel install -c testfiles/smoke.toml -t $(TESTTARGET) -v

.PHONY: all grapnel clean smoketest unittest
