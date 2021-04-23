// Copyright © 2018-2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "flag"

var (
	flagDebug   = flag.Bool("debug", false, "print debug messages")
	flagDumpFib = flag.Bool("dump-fib", false, "dump fibinfo after ifinfo")
	flagLog     = flag.String("log", "", "print to file instead of stdout")
	flagLicense = flag.Bool("license", false, "print license and exit")
	flagMux     = flag.String("mux", "xeth-mux", "netdev")
	flagVerbose = flag.Bool("verbose", false, "print xeth messages")
)

func init() { flag.Parse() }
