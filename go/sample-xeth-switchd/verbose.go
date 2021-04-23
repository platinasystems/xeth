// Copyright © 2018-2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

// Println(args...) if flagVerbose
func verbose(args ...interface{}) (n int, err error) {
	if *flagVerbose {
		n, err = fmt.Fprintln(log, args...)
	}
	return
}

// Printf(format, args...) w/ newline if flagVerbose
func verbosef(format string, args ...interface{}) (n int, err error) {
	if *flagVerbose {
		n, err = fmt.Fprintf(log, format+"\n", args...)
	}
	return
}
