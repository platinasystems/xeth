// Copyright © 2018-2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// +build 386 amd64 arm arm64 ppc64le mips64le mipsle riscv64 wasm

package endian

import "encoding/binary"

var (
	Host    = binary.LittleEndian
	Network = binary.BigEndian
)

func HostUint16(u uint16) uint16 { return NetworkUint16(u) }
func HostUint32(u uint32) uint32 { return NetworkUint32(u) }
func HostUint64(u uint64) uint64 { return NetworkUint64(u) }

func NetworkUint16(u uint16) uint16 {
	return ((u & 0x00ff) << 8) |
		((u & 0xff00) >> 8)

}

func NetworkUint32(u uint32) uint32 {
	return ((u & 0x000000ff) << 24) |
		((u & 0x0000ff00) << 8) |
		((u & 0x00ff0000) >> 8) |
		((u & 0xff000000) >> 24)
}

func NetworkUint64(u uint64) uint64 {
	return ((u & 0x00000000000000ff) << 56) |
		((u & 0x000000000000ff00) << 40) |
		((u & 0x0000000000ff0000) << 24) |
		((u & 0x00000000ff000000) << 8) |
		((u & 0x000000ff00000000) >> 8) |
		((u & 0x0000ff0000000000) >> 24) |
		((u & 0x00ff000000000000) >> 40) |
		((u & 0xff00000000000000) >> 56)
}
