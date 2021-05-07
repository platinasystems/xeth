// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import "fmt"

func (xid Xid) Format(w fmt.State, c rune) {
	if l := LinkOf(xid); l != nil {
		fmt.Fprint(w, l.IfInfoName())
	} else if xid > VlanNVid {
		fmt.Fprintf(w, "(%d, %d)", xid&VlanVidMask, xid/VlanNVid)
	} else {
		fmt.Fprint(w, uint32(xid))
	}
}

func (Break) String() string { return "break" }

func (dev DevNew) Format(w fmt.State, c rune) {
	xid := Xid(dev)
	fmt.Fprint(w, "new ", xid)
	if l := LinkOf(xid); l != nil {
		fmt.Fprint(w, " features ", l.IfInfoFeatures())
	}
}

func (dev DevDel) Format(w fmt.State, c rune) {
	fmt.Fprint(w, "del ", Xid(dev))
}

func (dev DevUp) Format(w fmt.State, c rune) {
	fmt.Fprint(w, Xid(dev), " up")
}

func (dev DevDown) Format(w fmt.State, c rune) {
	fmt.Fprint(w, Xid(dev), " down")
}

func (dev DevDump) Format(w fmt.State, c rune) {
	fmt.Fprint(w, Xid(dev), " dump")
}

func (reg DevReg) Format(w fmt.State, c rune) {
	xid := Xid(reg)
	fmt.Fprint(w, xid, " reg")
	if l := LinkOf(xid); l != nil {
		fmt.Fprint(w, " ", l.IfInfoNetNs())
	}
}

func (dev DevUnreg) Format(w fmt.State, c rune) {
	fmt.Fprint(w, Xid(dev), " unreg")
}

func (dev DevFeatures) Format(w fmt.State, c rune) {
	var f IfInfoFeatures
	xid := Xid(dev)
	if l := LinkOf(xid); l != nil {
		f = l.IfInfoFeatures()
	}
	fmt.Fprint(w, xid, " features ", f)
}

func (dev *DevAddIPNet) Format(w fmt.State, c rune) {
	fmt.Fprint(w, dev.Xid, " add ", dev.IPNet)
}

func (dev *DevDelIPNet) Format(w fmt.State, c rune) {
	fmt.Fprint(w, dev.Xid, " del ", dev.Prefix)
}

func (dev *DevJoin) Format(w fmt.State, c rune) {
	fmt.Fprint(w, dev.Lower, " join ", dev.Upper)
}

func (dev *DevQuit) Format(w fmt.State, c rune) {
	fmt.Fprint(w, dev.Lower, " quit ", dev.Upper)
}

func (dev *DevEthtoolFlags) Format(w fmt.State, c rune) {
	fmt.Fprint(w, dev.Xid, " ethtool flags <", dev.EthtoolFlagBits, ">")
}

func (dev DevEthtoolSettings) Format(w fmt.State, c rune) {
	xid := Xid(dev)
	fmt.Fprint(w, xid)
	if l := LinkOf(xid); l != nil {
		fmt.Fprint(w, " speed ", l.EthtoolSpeed(), " (mbps)")
		fmt.Fprint(w, " autoneg ", l.EthtoolAutoNeg())
		fmt.Fprint(w, " duplex ", l.EthtoolDuplex())
		fmt.Fprint(w, " port ", l.EthtoolDevPort())
	} else {
		fmt.Fprint(w, " ethtool settings")
	}
}

func (dev DevLinkModesSupported) Format(w fmt.State, c rune) {
	xid := Xid(dev)
	fmt.Fprint(w, xid)
	if l := LinkOf(xid); l != nil {
		fmt.Fprint(w, " <", l.LinkModesSupported(), ">")
	} else {
		fmt.Fprint(w, " <supported link modes>")
	}
}

func (dev DevLinkModesAdvertising) Format(w fmt.State, c rune) {
	xid := Xid(dev)
	fmt.Fprint(w, xid)
	if l := LinkOf(xid); l != nil {
		fmt.Fprint(w, " <", l.LinkModesAdvertising(), ">")
	} else {
		fmt.Fprint(w, " <advertising link modes>")
	}
}

func (dev DevLinkModesLPAdvertising) Format(w fmt.State, c rune) {
	xid := Xid(dev)
	fmt.Fprint(w, xid)
	if l := LinkOf(xid); l != nil {
		fmt.Fprint(w, " <", l.LinkModesLPAdvertising(), ">")
	} else {
		fmt.Fprint(w, " <link partner advertising link modes>")
	}
}

func (bits EthtoolFlagBits) Format(w fmt.State, c rune) {
	if bits == 0 {
		w.Write([]byte("none"))
	} else {
		fmt.Fprintf(w, "0b%b", uint32(bits))
	}
}

func (bits EthtoolLinkModeBits) Format(w fmt.State, c rune) {
	sep := ""
	for i, s := range []string{
		"10baseT-half",
		"10baseT-full",
		"100baseT-half",
		"100baseT-full",
		"1000baseT-half",
		"1000baseT-full",
		"Autoneg",
		"TP",
		"AUI",
		"MII",
		"FIBRE",
		"BNC",
		"10000baseT-full",
		"Pause",
		"Asym-Pause",
		"2500baseX-full",
		"Backplane",
		"1000baseKX-full",
		"10000baseKX4-full",
		"10000baseKR-full",
		"10000baseR-FEC",
		"20000baseMLD2-full",
		"20000baseKR2-full",
		"40000baseKR4-full",
		"40000baseCR4-full",
		"40000baseSR4-full",
		"40000baseLR4-full",
		"56000baseKR4-full",
		"56000baseCR4-full",
		"56000baseSR4-full",
		"56000baseLR4-full",
		"25000baseCR-full",
		"25000baseKR-full",
		"25000baseSR-full",
		"50000baseCR2-full",
		"50000baseKR2-full",
		"100000baseKR4-full",
		"100000baseSR4-full",
		"100000baseCR4-full",
		"100000baseLR4-ER4-full",
		"50000baseSR2-full",
		"1000baseX-full",
		"10000baseCR-full",
		"10000baseSR-full",
		"10000baseLR-full",
		"10000baseLRM-full",
		"10000baseER-full",
		"2500baseT-full",
		"5000baseT-full",
		"fec-none",
		"fec-rs",
		"fec-baser",
	} {
		if bits.Test(uint(i)) {
			fmt.Fprint(w, sep, s)
			sep = ", "
		}
	}
	if len(sep) == 0 {
		fmt.Fprint(w, "none")
	}
}

func (msg *FibEntry) Format(w fmt.State, c rune) {
	fmt.Fprint(w, msg.FibEntryEvent)
	fmt.Fprint(w, " netns ", msg.NetNs)
	fmt.Fprint(w, " table ", msg.RtTable)
	fmt.Fprint(w, " type ", msg.Rtn)
	fmt.Fprint(w, " ", &msg.IPNet)
	if len(msg.NHs) == 1 {
		fmt.Fprint(w, " nexthop ", msg.NHs[0])
	} else {
		fmt.Fprint(w, " nexthops [")
		sep := ""
		for _, nh := range msg.NHs {
			fmt.Fprint(w, sep, nh)
			sep = ", "
		}
		fmt.Fprint(w, "]")
	}
}

func (msg *Neighbor) Format(w fmt.State, c rune) {
	fmt.Fprint(w, "neighbor")
	fmt.Fprint(w, " netns ", msg.NetNs)
	fmt.Fprint(w, " ", msg.Xid)
	fmt.Fprint(w, " ", msg.IP)
	fmt.Fprint(w, " ", msg.HardwareAddr)
}

func (msg NetNsAdd) Format(w fmt.State, c rune) {
	fmt.Fprint(w, "netns add ", msg.NetNs)
}

func (msg NetNsDel) Format(w fmt.State, c rune) {
	fmt.Fprint(w, "netns del ", msg.NetNs)
}

func (nh *NH) Format(w fmt.State, c rune) {
	fmt.Fprint(w, "{")
	fmt.Fprint(w, nh.IP)
	fmt.Fprint(w, " ", nh.Xid)
	fmt.Fprint(w, " weight ", nh.Weight)
	fmt.Fprint(w, " flags <", nh.RtnhFlags, ">")
	fmt.Fprint(w, " scope ", nh.RtScope)
	fmt.Fprint(w, "}")
}

func (event FibEntryEvent) String() string {
	s, found := map[FibEntryEvent]string{
		FIB_EVENT_ENTRY_REPLACE: "replace",
		FIB_EVENT_ENTRY_APPEND:  "append",
		FIB_EVENT_ENTRY_ADD:     "add",
		FIB_EVENT_ENTRY_DEL:     "del",
		FIB_EVENT_RULE_ADD:      "rule-add",
		FIB_EVENT_RULE_DEL:      "rule-del",
		FIB_EVENT_NH_ADD:        "nh-add",
		FIB_EVENT_NH_DEL:        "nh-del",
	}[event]
	if !found {
		s = fmt.Sprint("unknown-", uint8(event))
	}
	return s
}

func (stat LinkStat) String() string {
	s, found := map[LinkStat]string{
		LinkStatRxPackets:         "rx-packets",
		LinkStatTxPackets:         "tx-packets",
		LinkStatRxBytes:           "rx-bytes",
		LinkStatTxBytes:           "tx-bytes",
		LinkStatRxErrors:          "rx-errors",
		LinkStatTxErrors:          "tx-errors",
		LinkStatRxDropped:         "rx-dropped",
		LinkStatTxDropped:         "tx-dropped",
		LinkStatMulticast:         "multicast",
		LinkStatCollisions:        "collisions",
		LinkStatRxLengthErrors:    "rx-length-errors",
		LinkStatRxOverErrors:      "rx-over-errors",
		LinkStatRxCrcErrors:       "rx-crc-errors",
		LinkStatRxFrameErrors:     "rx-frame-errors",
		LinkStatRxFifoErrors:      "rx-fifo-errors",
		LinkStatRxMissedErrors:    "rx-missed-errors",
		LinkStatTxAbortedErrors:   "tx-aborted-errors",
		LinkStatTxCarrierErrors:   "tx-carrier-errors",
		LinkStatTxFifoErrors:      "tx-fifo-errors",
		LinkStatTxHeartbeatErrors: "tx-heartbeat-errors",
		LinkStatTxWindowErrors:    "tx-window-errors",
		LinkStatRxCompressed:      "rx-compressed",
		LinkStatTxCompressed:      "tx-compressed",
		LinkStatRxNohandler:       "rx-nohandler",
	}[stat]
	if found {
		return s
	}
	return "invalid-link-stat"
}

func (autoneg AutoNeg) String() string {
	s, found := map[AutoNeg]string{
		AUTONEG_DISABLE: "disabled",
		AUTONEG_ENABLE:  "enabled",
	}[autoneg]
	if !found {
		if autoneg == AUTONEG_UNKNOWN {
			s = "unknown"
		} else {
			s = fmt.Sprint("unknown-", uint8(autoneg))
		}
	}
	return s
}

func (kind DevKind) String() string {
	s, found := map[DevKind]string{
		DevKindUnspec: "unspecified",
		DevKindPort:   "port",
		DevKindVlan:   "vlan",
		DevKindBridge: "bridge",
		DevKindLag:    "lag",
	}[kind]
	if !found {
		s = fmt.Sprint("unknown-", uint8(kind))
	}
	return s
}

func (port DevPort) String() string {
	s, found := map[DevPort]string{
		PORT_TP:    "tp",
		PORT_AUI:   "aui",
		PORT_MII:   "mii",
		PORT_FIBRE: "fibre",
		PORT_BNC:   "bnc",
		PORT_DA:    "da",
	}[port]
	if !found {
		if port == PORT_NONE {
			s = "none"
		} else if port == PORT_OTHER {
			s = "other"
		} else {
			s = fmt.Sprint("unknown", uint8(port))
		}
	}
	return s
}

func (duplex Duplex) String() string {
	s, found := map[Duplex]string{
		DUPLEX_HALF: "half",
		DUPLEX_FULL: "full",
	}[duplex]
	if !found {
		if duplex == DUPLEX_UNKNOWN {
			s = "unknown"
		} else {
			s = fmt.Sprint("uknown-", uint8(duplex))
		}
	}
	return s
}

func (rtn Rtn) String() string {
	s, found := map[Rtn]string{
		RTN_UNSPEC:      "unspec",
		RTN_UNICAST:     "unicast",
		RTN_LOCAL:       "local",
		RTN_BROADCAST:   "broadcast",
		RTN_ANYCAST:     "anycast",
		RTN_MULTICAST:   "multicast",
		RTN_BLACKHOLE:   "blackhole",
		RTN_UNREACHABLE: "unreachable",
		RTN_PROHIBIT:    "prohibit",
		RTN_THROW:       "throw",
		RTN_NAT:         "nat",
		RTN_XRESOLVE:    "xresolve",
	}[rtn]
	if !found {
		s = fmt.Sprint("unknown-", uint8(rtn))
	}
	return s
}

func (flags RtnhFlags) Format(w fmt.State, c rune) {
	sep := ""
	for _, x := range []struct {
		flag RtnhFlags
		name string
	}{
		{RTNH_F_DEAD, "dead"},
		{RTNH_F_PERVASIVE, "pervasive"},
		{RTNH_F_ONLINK, "on-link"},
		{RTNH_F_OFFLOAD, "off-load"},
		{RTNH_F_LINKDOWN, "link-down"},
		{RTNH_F_UNRESOLVED, "unresolved"},
	} {
		if flags&x.flag == x.flag {
			fmt.Fprint(w, sep, x.name)
			sep = ", "
		}
	}
	if len(sep) == 0 {
		fmt.Fprint(w, "none")
	}
}

func (rtt RtTable) String() string {
	s, found := map[RtTable]string{
		RT_TABLE_UNSPEC:  "unspec",
		RT_TABLE_COMPAT:  "compat",
		RT_TABLE_DEFAULT: "default",
		RT_TABLE_MAIN:    "main",
		RT_TABLE_LOCAL:   "local",
		RT_TABLE_MAX:     "max",
	}[rtt]
	if !found {
		s = fmt.Sprint(uint32(rtt))
	}
	return s
}

func (scope RtScope) String() string {
	s, found := map[RtScope]string{
		RT_SCOPE_UNIVERSE: "universe",
		RT_SCOPE_SITE:     "site",
		RT_SCOPE_LINK:     "link",
		RT_SCOPE_HOST:     "host",
		RT_SCOPE_NOWHERE:  "nowhere",
	}[scope]
	if !found {
		s = fmt.Sprint("undefined-", uint8(scope))
	}
	return s
}

func (f Frame) Format(w fmt.State, c rune) {
	fmt.Fprint(w, f.Dst())
	fmt.Fprint(w, " ", f.Src())
	if xid := f.Xid(); xid != 0 {
		fmt.Fprint(w, " {", xid, ")")
	}
	fmt.Fprint(w, " ", f.EthP())
}

func (f IfInfoFeatures) Format(w fmt.State, c rune) {
	var sep string
	for _, flag := range []struct {
		f    IfInfoFeatures
		name string
	}{
		{NetIfHwL2FwdOffload, "l2-fwd-offload"},
	} {
		if f.Has(flag.f) {
			fmt.Fprint(w, sep, flag.name)
			sep = ", "
		}
	}
	if len(sep) == 0 {
		fmt.Fprint(w, "off")
	}
}
