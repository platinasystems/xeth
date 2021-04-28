// Copyright © 2018-2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build ppc64 s390x mips mips64

package endian

import "encoding/binary"

var (
	Host    = binary.BigEndian
	Network = binary.BigEndian
)

func HostUint16(u uint16) uint16 { return u }
func HostUint32(u uint32) uint32 { return u }
func HostUint64(u uint64) uint64 { return u }

func NetworkUint16(u uint16) uint16 { return u }
func NetworkUint32(u uint32) uint16 { return u }
func NetworkUint64(u uint64) uint16 { return u }
