// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import (
	"syscall"

	"github.com/platinasystems/xeth/v3/go/endian"
)

const (
	ETH_P = (2 * syscall.ARPHRD_IEEE802) + iota
	_
	ETH_PAYLOAD
)

const (
	ETH_VLAN_TPID = ETH_P + iota
	_
	ETH_VLAN_TCI
	_
	ETH_VLAN_P
	_
	ETH_VLAN_PAYLOAD
)

// inner CTAG
const (
	ETH_C_VLAN_TPID = ETH_VLAN_P + iota
	_
	ETH_C_VLAN_TCI
	_
	ETH_C_VLAN_P
	_
	ETH_C_VLAN_PAYLOAD
)

type EthP uint16

func (p EthP) Network() uint16 { return endian.NetworkUint16(uint16(p)) }

const (
	ETH_P_1588       EthP = syscall.ETH_P_1588
	ETH_P_8021Q      EthP = syscall.ETH_P_8021Q
	ETH_P_8021AD     EthP = 0x88a8
	ETH_P_802_2      EthP = syscall.ETH_P_802_2
	ETH_P_802_3      EthP = syscall.ETH_P_802_3
	ETH_P_AARP       EthP = syscall.ETH_P_AARP
	ETH_P_ALL        EthP = syscall.ETH_P_ALL
	ETH_P_AOE        EthP = syscall.ETH_P_AOE
	ETH_P_ARCNET     EthP = syscall.ETH_P_ARCNET
	ETH_P_ARP        EthP = syscall.ETH_P_ARP
	ETH_P_ATALK      EthP = syscall.ETH_P_ATALK
	ETH_P_ATMFATE    EthP = syscall.ETH_P_ATMFATE
	ETH_P_ATMMPOA    EthP = syscall.ETH_P_ATMMPOA
	ETH_P_AX25       EthP = syscall.ETH_P_AX25
	ETH_P_BPQ        EthP = syscall.ETH_P_BPQ
	ETH_P_CAIF       EthP = syscall.ETH_P_CAIF
	ETH_P_CAN        EthP = syscall.ETH_P_CAN
	ETH_P_CONTROL    EthP = syscall.ETH_P_CONTROL
	ETH_P_CUST       EthP = syscall.ETH_P_CUST
	ETH_P_DDCMP      EthP = syscall.ETH_P_DDCMP
	ETH_P_DEC        EthP = syscall.ETH_P_DEC
	ETH_P_DIAG       EthP = syscall.ETH_P_DIAG
	ETH_P_DNA_DL     EthP = syscall.ETH_P_DNA_DL
	ETH_P_DNA_RC     EthP = syscall.ETH_P_DNA_RC
	ETH_P_DNA_RT     EthP = syscall.ETH_P_DNA_RT
	ETH_P_DSA        EthP = syscall.ETH_P_DSA
	ETH_P_ECONET     EthP = syscall.ETH_P_ECONET
	ETH_P_EDSA       EthP = syscall.ETH_P_EDSA
	ETH_P_FCOE       EthP = syscall.ETH_P_FCOE
	ETH_P_FIP        EthP = syscall.ETH_P_FIP
	ETH_P_HDLC       EthP = syscall.ETH_P_HDLC
	ETH_P_IEEE802154 EthP = syscall.ETH_P_IEEE802154
	ETH_P_IEEEPUP    EthP = syscall.ETH_P_IEEEPUP
	ETH_P_IEEEPUPAT  EthP = syscall.ETH_P_IEEEPUPAT
	ETH_P_IP         EthP = syscall.ETH_P_IP
	ETH_P_IPV6       EthP = syscall.ETH_P_IPV6
	ETH_P_IPX        EthP = syscall.ETH_P_IPX
	ETH_P_IRDA       EthP = syscall.ETH_P_IRDA
	ETH_P_LAT        EthP = syscall.ETH_P_LAT
	ETH_P_LINK_CTL   EthP = syscall.ETH_P_LINK_CTL
	ETH_P_LOCALTALK  EthP = syscall.ETH_P_LOCALTALK
	ETH_P_LOOP       EthP = syscall.ETH_P_LOOP
	ETH_P_MOBITEX    EthP = syscall.ETH_P_MOBITEX
	ETH_P_MPLS_MC    EthP = syscall.ETH_P_MPLS_MC
	ETH_P_MPLS_UC    EthP = syscall.ETH_P_MPLS_UC
	ETH_P_PAE        EthP = syscall.ETH_P_PAE
	ETH_P_PAUSE      EthP = syscall.ETH_P_PAUSE
	ETH_P_PHONET     EthP = syscall.ETH_P_PHONET
	ETH_P_PPPTALK    EthP = syscall.ETH_P_PPPTALK
	ETH_P_PPP_DISC   EthP = syscall.ETH_P_PPP_DISC
	ETH_P_PPP_MP     EthP = syscall.ETH_P_PPP_MP
	ETH_P_PPP_SES    EthP = syscall.ETH_P_PPP_SES
	ETH_P_PUP        EthP = syscall.ETH_P_PUP
	ETH_P_PUPAT      EthP = syscall.ETH_P_PUPAT
	ETH_P_RARP       EthP = syscall.ETH_P_RARP
	ETH_P_SCA        EthP = syscall.ETH_P_SCA
	ETH_P_SLOW       EthP = syscall.ETH_P_SLOW
	ETH_P_SNAP       EthP = syscall.ETH_P_SNAP
	ETH_P_TEB        EthP = syscall.ETH_P_TEB
	ETH_P_TIPC       EthP = syscall.ETH_P_TIPC
	ETH_P_TRAILER    EthP = syscall.ETH_P_TRAILER
	ETH_P_TR_802_2   EthP = syscall.ETH_P_TR_802_2
	ETH_P_WAN_PPP    EthP = syscall.ETH_P_WAN_PPP
	ETH_P_WCCP       EthP = syscall.ETH_P_WCCP
	ETH_P_X25        EthP = syscall.ETH_P_X25
	ETH_P_ATT_PACE   EthP = 0x7373
)
