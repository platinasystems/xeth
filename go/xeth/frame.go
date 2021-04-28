// Copyright © 2018-2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import (
	"net"

	"github.com/platinasystems/xeth/v3/go/endian"
	"github.com/platinasystems/xeth/v3/go/xeth/internal"
)

type Frame struct{ Buffer }

func (f Frame) Dst() net.HardwareAddr {
	return net.HardwareAddr(f.bytes()[:internal.SizeofEthAddr])
}

func (f Frame) Src() net.HardwareAddr {
	const (
		from = internal.SizeofEthAddr
		to   = 2 * internal.SizeofEthAddr
	)
	return net.HardwareAddr(f.bytes()[from:to])
}

func (f Frame) Loopback(task *Task) {
	task.ExceptionFrame(f.bytes())
}

func (f Frame) Xid(set ...Xid) (xid Xid) {
	b := f.bytes()
	p := EthP(endian.Host.Uint16(b[ETH_P:]))
	if len(set) > 0 {
		xid = set[0]
		// FIXME do we need 8021AD?
		endian.Network.PutUint16(b[ETH_VLAN_TCI:], uint16(xid))
	} else if p == ETH_P_8021AD {
		xid = Xid(endian.Host.Uint16(b[ETH_C_VLAN_TCI:]))
		xid <<= EncapVlanVidBit
		xid |= Xid(endian.Host.Uint16(b[ETH_VLAN_TCI:]))
	} else if p == ETH_P_8021Q {
		xid = Xid(endian.Host.Uint16(b[ETH_VLAN_TCI:]))
	}
	return
}

func (f Frame) EthP() EthP {
	b := f.bytes()
	p := EthP(endian.Host.Uint16(b[ETH_P:]))
	if p == ETH_P_8021AD {
		p = EthP(endian.Host.Uint16(b[ETH_C_VLAN_P:]))
	} else if p == ETH_P_8021Q {
		p = EthP(endian.Host.Uint16(b[ETH_VLAN_P:]))
	} else if p < 0x200 {
		p = ETH_P_802_3
	}
	return p
}
