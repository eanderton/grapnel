package cmd

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
	"fmt"
	. "grapnel/flag"
	. "grapnel/lib"
	log "grapnel/log"
	"os"
)

// TODO: if no lock file can be found, then fail over to update instead
// TODO: move lock file writing to updateFn()

func installFn(cmd *Command, args []string) error {
	configureLogging()

	if len(args) > 0 {
		return fmt.Errorf("Too many arguments for 'install'")
	}

	// set unset paramters to the defaults
	if lockFileName == "" {
		lockFileName = defaultLockFileName
	}
	if targetPath == "" {
		targetPath = defaultTargetPath
	}

	log.Debug("lock file: %v", lockFileName)
	log.Debug("target path: %v", targetPath)

	// get dependencies from the lockfile
	deplist, err := LoadGrapnelDepsfile(lockFileName)
	if err != nil {
		return err
	} else if deplist == nil {
		// TODO: fail over to update instead?
		return fmt.Errorf("Cannot open lock file: '%s'", lockFileName)
	}
	log.Info("loaded %d dependency definitions", len(deplist))

	log.Info("installing to: %v", targetPath)
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return err
	}

	libs := []*Library{}
	// cleanup
	defer func() {
		for _, lib := range libs {
			lib.Destroy()
		}
	}()

	// resolve all the dependencies
	resolver, err := getResolver()
	if err != nil {
		return err
	}
	libs, err = resolver.ResolveDependencies(deplist)
	if err != nil {
		return err
	}

	// install all the dependencies
	log.Info("Resolved %v dependencies. Installing.", len(libs))
	resolver.InstallLibraries(targetPath, libs)

	log.Info("Install complete")
	return nil
}

var installCmd = Command{
	Desc: "Downloads and installs locked dependencies.",
	Help: " Installs packages at 'targetPath', from configured lock file.\n" +
		"\nDefaults:\n" +
		"  Lock file = " + defaultLockFileName + "\n" +
		"  Target path = " + defaultTargetPath + "\n",
	Flags: FlagMap{
		"lockfile": &Flag{
			Alias:   "l",
			Desc:    "Grapnel lock file",
			ArgDesc: "[filename]",
			Fn:      StringFlagFn(&lockFileName),
		},
		"target": &Flag{
			Alias:   "t",
			Desc:    "Target installation path",
			ArgDesc: "[target]",
			Fn:      StringFlagFn(&targetPath),
		},
	},
	Fn: installFn,
}
