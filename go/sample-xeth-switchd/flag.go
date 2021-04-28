// Copyright © 2018-2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	flagDebug   = flag.Bool("debug", false, "print debug messages")
	flagDumpFib = flag.Bool("dump-fib", false, "dump fibinfo after ifinfo")
	flagLog     = flag.String("log", "", "print to file instead of stdout")
	flagLicense = flag.Bool("license", false, "print license and exit")
	flagMux     = flag.String("mux", "xeth-mux", "netdev")
	flagVerbose = flag.Bool("verbose", false, "print xeth messages")
)

var (
	log   = os.Stdout
	debug = func(args ...interface{}) (int, error) {
		return 0, nil
	}
	debugf = func(format string, args ...interface{}) (int, error) {
		return 0, nil
	}
	verbose = func(args ...interface{}) (int, error) {
		return 0, nil
	}
	verbosef = func(format string, args ...interface{}) (int, error) {
		return 0, nil
	}
)

func init() {
	flag.Parse()
	if *flagDebug {
		debug = func(args ...interface{}) (int, error) {
			return fmt.Fprintln(log, args...)
		}
		debugf = func(format string, args ...interface{}) (int, error) {
			return fmt.Fprintf(log, format+"\n", args...)
		}
	}
	if *flagVerbose {
		verbose = func(args ...interface{}) (int, error) {
			return fmt.Fprintln(log, args...)
		}
		verbosef = func(format string, args ...interface{}) (int, error) {
			return fmt.Fprintf(log, format+"\n", args...)
		}
	}
	if len(*flagLog) > 0 {
		f, err := os.Create(*flagLog)
		if err != nil {
			panic(err)
		}
		log = f
	}
}
