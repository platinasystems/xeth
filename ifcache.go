/* Copyright(c) 2018 Platina Systems, Inc.
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms and conditions of the GNU General Public License,
 * version 2, as published by the Free Software Foundation.
 *
 * This program is distributed in the hope it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin St - Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * The full GNU General Public License is included in this distribution in
 * the file called "COPYING".
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */

package xeth

import (
	"bytes"
	"fmt"
	"net"
	"sort"
	"sync"
)

type NoValue struct{}
type Associates struct {
	mu sync.RWMutex
	m  map[int32]NoValue
}

type InterfaceEntry struct {
	Ifinfo
	EthtoolPrivFlags
	EthtoolSettings
	IPNets []*net.IPNet
	Uppers Associates
	Lowers Associates
}

type ifMap struct {
	mu    sync.RWMutex
	index map[int32]*InterfaceEntry
}

func (i *ifMap) validate() {
	i.mu.Lock()
	if i.index == nil {
		i.index = make(map[int32]*InterfaceEntry)
	}
	i.mu.Unlock()
}

func (i *ifMap) get(ifindex int32) (entry *InterfaceEntry, found bool) {
	i.mu.RLock()
	entry, found = i.index[ifindex]
	i.mu.RUnlock()
	return
}

func (i *ifMap) set(ifindex int32, entry *InterfaceEntry) {
	i.validate()
	i.mu.Lock()
	i.index[ifindex] = entry
	i.mu.Unlock()
}

func (i *ifMap) unset(ifindex int32) {
	i.mu.Lock()
	delete(i.index, ifindex)
	i.mu.Unlock()
}

func (i *ifMap) unsetAll() {
	i.mu.Lock()
	for ifindex := range i.index {
		delete(i.index, ifindex)
	}
	i.index = nil
	i.mu.Unlock()
}

type dirMap struct {
	mu    sync.RWMutex
	index map[string]*InterfaceEntry
}

func (i *dirMap) validate() {
	i.mu.Lock()
	if i.index == nil {
		i.index = make(map[string]*InterfaceEntry)
	}
	i.mu.Unlock()
}

func (i *dirMap) get(name string) (entry *InterfaceEntry, found bool) {
	i.mu.RLock()
	entry, found = i.index[name]
	i.mu.RUnlock()
	return
}

func (i *dirMap) set(name string, entry *InterfaceEntry) {
	i.validate()
	i.mu.Lock()
	i.index[name] = entry
	i.mu.Unlock()
}

func (i *dirMap) unset(name string) {
	i.mu.Lock()
	delete(i.index, name)
	i.mu.Unlock()
}

func (i *dirMap) unsetAll() {
	i.mu.Lock()
	for name := range i.index {
		delete(i.index, name)
	}
	i.index = nil
	i.mu.Unlock()
}

type Ifcache struct {
	indexes []int32
	index   ifMap
	// only map XETH_DEVTYPE_XETH_PORT by name
	dir dirMap
}

var Interface Ifcache

func (c *Ifcache) Indexed(ifindex int32) *InterfaceEntry {
	if entry, found := c.index.get(ifindex); found {
		return entry
	}
	if p, err := net.InterfaceByIndex(int(ifindex)); err == nil {
		return c.cache(int32(p.Index), p)
	}
	return nil
}

// Call given function with each cached interface entry ceasing on error.
func (c *Ifcache) Iterate(f func(*InterfaceEntry) error) error {
	sort.Slice(c.indexes, func(i, j int) bool {
		return c.indexes[i] < c.indexes[j]
	})
	for _, ifindex := range c.indexes {
		if entry, found := c.index.get(ifindex); found {
			if err := f(entry); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Ifcache) Named(name string) *InterfaceEntry {
	entry, _ := c.dir.get(name)
	return entry
}

func (c *Ifcache) cache(ifindex int32, args ...interface{}) *InterfaceEntry {
	entry, found := c.index.get(ifindex)
	if !found {
		entry = c.newEntry(ifindex)
	}
	entry.cache(args...)
	return entry
}

func (c *Ifcache) del(ifindex int32) {
	if entry := c.Indexed(ifindex); entry != nil {
		if len(entry.IPNets) > 0 {
			entry.IPNets = entry.IPNets[:0]
		}
		c.index.unset(ifindex)
		c.dir.unset(entry.String())
		for i := range c.indexes {
			if c.indexes[i] == ifindex {
				copy(c.indexes[i:], c.indexes[i+1:])
				c.indexes = c.indexes[:len(c.indexes)-1]
				break
			}
		}
	}
}

func (c *Ifcache) newEntry(ifindex int32) *InterfaceEntry {
	entry := new(InterfaceEntry)
	entry.Index = ifindex
	c.index.set(ifindex, entry)
	c.indexes = append(c.indexes, ifindex)
	return entry
}

func (entry *InterfaceEntry) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprint(buf, entry.Ifinfo.Index, ": ", entry.Ifinfo.Name)
	if entry.Ifinfo.Link > 0 {
		fmt.Fprint(buf, "@", Interface.Indexed(entry.Link).Ifinfo.Name)
	}
	fmt.Fprint(buf, ":")
	if entry.Ifinfo.Flags != 0 {
		fmt.Fprint(buf, " <", entry.Ifinfo.Flags, ">")
	}
	fmt.Fprint(buf, " reason ", entry.Ifinfo.Reason)
	if entry.Ifinfo.Id != 0 {
		fmt.Fprint(buf, " id ", entry.Ifinfo.Id)
	}
	if entry.Ifinfo.Port >= 0 {
		fmt.Fprint(buf, " port ", entry.Ifinfo.Port)
	}
	if entry.Ifinfo.Subport >= 0 {
		fmt.Fprint(buf, " subport ", entry.Ifinfo.Subport)
	}
	if entry.Ifinfo.Netns != DefaultNetns {
		fmt.Fprint(buf, " netns ", entry.Ifinfo.Netns)
	}
	fmt.Fprint(buf, "\n    link/", entry.Ifinfo.DevType)
	fmt.Fprint(buf, " ", entry.Ifinfo.HardwareAddr())
	if entry.EthtoolPrivFlags != 0 {
		fmt.Fprint(buf, " <", entry.EthtoolPrivFlags, ">")
	}
	if entry.EthtoolSettings.Speed != 0 {
		fmt.Fprint(buf, " speed ", entry.EthtoolSettings.Speed)
		fmt.Fprint(buf, " autoneg ", entry.EthtoolSettings.Autoneg)
	}
	if entry.Uppers.NotEmpty() {
		fmt.Fprint(buf, " uppers [", entry.Uppers, "]")
	}
	if entry.Lowers.NotEmpty() {
		fmt.Fprint(buf, " lowers [", entry.Lowers, "]")
	}
	for _, ipnet := range entry.IPNets {
		fmt.Fprint(buf, "\n    ")
		if ipnet.IP.To4() != nil {
			fmt.Fprint(buf, "inet ", ipnet)
		} else {
			fmt.Fprint(buf, "inet6 ", ipnet)
		}
	}
	return buf.String()
}

func (entry *InterfaceEntry) cache(args ...interface{}) {
	for _, v := range args {
		switch t := v.(type) {
		case string:
			entry.dub(t)
		case *net.Interface:
			// don't override Index set by newEntry
			entry.dub(t.Name)
			entry.Link = -1
			entry.Netns = DefaultNetns
			copy(entry.addr[:], t.HardwareAddr)
			entry.Flags = t.Flags
			entry.DevType = XETH_DEVTYPE_LINUX_UNKNOWN
			entry.Reason = XETH_IFINFO_REASON_NEW
			entry.Id = 0
			entry.Port = -1
			entry.Subport = -1
		case *MsgChangeUpper:
			upper := Interface.Indexed(t.Upper)
			entry.Uppers.validate()
			entry.Lowers.validate()
			if t.Linking > 0 {
				entry.Uppers.Add(t.Upper)
				upper.Lowers.Add(t.Lower)
			} else {
				entry.Uppers.Del(t.Upper)
				entry.Uppers.Del(t.Lower)
			}
		case *MsgIfinfo:
			entry.dub((*Ifname)(&t.Ifname).String())
			entry.Link = t.Iflinkindex
			entry.Netns = Netns(t.Net)
			copy(entry.addr[:], t.Addr[:])
			entry.Ifinfo.Flags = net.Flags(t.Flags)
			entry.DevType = DevType(t.Devtype)
			entry.Reason = IfinfoReason(t.Reason)
			entry.Id = t.Id
			entry.Port = t.Portindex
			entry.Subport = t.Subportindex
		case IfinfoReason:
			entry.Reason = t
		case net.HardwareAddr:
			copy(entry.addr[:], t)
		case DevType:
			entry.DevType = t
		case net.Flags:
			entry.Flags = t
		case Netns:
			entry.Netns = t
		case *MsgIfa:
			switch t.Event {
			case IFA_ADD:
				entry.IPNets = append(entry.IPNets, t.IPNet())
			case IFA_DEL:
				ipnet := t.IPNet()
				n := len(entry.IPNets)
				for i, x := range entry.IPNets {
					if x.IP.Equal(ipnet.IP) {
						copy(entry.IPNets[i:],
							entry.IPNets[i+1:])
						entry.IPNets[n-1] = nil
						entry.IPNets =
							entry.IPNets[:n-1]
						break
					}
				}
			}
		case *MsgEthtoolFlags:
			entry.EthtoolPrivFlags.cache(t)
		case EthtoolPrivFlags:
			entry.EthtoolPrivFlags.cache(t)
		case *MsgEthtoolSettings:
			entry.EthtoolSettings.cache(t)
		case Mbps:
			entry.EthtoolSettings.cache(t)
		case Duplex:
			entry.EthtoolSettings.cache(t)
		case DevPort:
			entry.EthtoolSettings.cache(t)
		case Autoneg:
			entry.EthtoolSettings.cache(t)
		}
	}
}

func (entry *InterfaceEntry) dub(name string) {
	if entry.Name == name {
		return
	}
	if entry.DevType == XETH_DEVTYPE_XETH_PORT {
		if len(entry.Name) > 0 {
			Interface.dir.unset(entry.Name)
		}
		Interface.dir.set(name, entry)
	}
	entry.Name = name
}

func (associates *Associates) validate() {
	associates.mu.Lock()
	if associates.m == nil {
		associates.m = make(map[int32]NoValue)
	}
	associates.mu.Unlock()
}

func (associates *Associates) NotEmpty() bool {
	associates.mu.RLock()
	notEmpty := associates.m != nil && len(associates.m) > 0
	associates.mu.RUnlock()
	return notEmpty
}

func (associates *Associates) Add(ifindex int32) {
	associates.validate()
	associates.mu.Lock()
	associates.m[ifindex] = NoValue{}
	associates.mu.Unlock()
}

func (associates *Associates) Del(ifindex int32) {
	associates.mu.Lock()
	delete(associates.m, ifindex)
	associates.mu.Unlock()
}

func (associates *Associates) String() string {
	buf := new(bytes.Buffer)
	sep := ""
	associates.mu.RLock()
	for ifindex := range associates.m {
		fmt.Fprint(buf, sep, Interface.Indexed(ifindex).Ifinfo.Name)
		sep = ", "
	}
	associates.mu.RUnlock()
	return buf.String()
}

func (associates *Associates) ForeachKey(f func(key int32)) {
	associates.mu.RLock()
	for key, _ := range associates.m {
		f(key)
	}
	associates.mu.RUnlock()
}
