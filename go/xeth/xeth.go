// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package provides a sideband control interface to an XETH driver.
// See sample-xeth-switchd
package xeth

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/platinasystems/xeth/v3/go/xeth/internal"
)

//go:generate sh godef.sh godefs.go godefed.go
//go:generate sh godef.sh internal/godefs.go internal/godefed.go
//go:generate stringer -type=EthP -trimprefix=ETH_P_ ethp.go

const unixpacket = "unixpacket"

type Break struct{}

var (
	Cloned  Counter // cloned received messages
	Parsed  Counter // messages parsed by user
	Dropped Counter // messages that overflowed transmit channel
	Sent    Counter // messages and exception frames sent to driver
	Unknown Counter // LinkOf xid w/o IfInfo
)

type Task struct {
	RxCh <-chan Buffer // cloned msgs received from driver

	WG   *sync.WaitGroup
	Stop <-chan struct{}
	sock *net.UnixConn

	loch chan<- buffer // low priority, leaky-bucket tx channel
	hich chan<- buffer // high priority, unbuffered, no-drop tx channel

	RxErr error // error that stopped the rx service
	TxErr error // error that stopped the tx service

	muxfd int
	muxsa syscall.SockaddrLinklayer
}

// Write provision value to platform device sysfs file
func Provision(dev, val string) error {
	if len(val) == 0 {
		// have to write something to get sysfs store handler to run
		// which will then open the atsock
		val = " "
	}
	sysfsdir := filepath.Join("/sys/bus/platform/devices", dev)
	_, err := os.Stat(sysfsdir)
	if err != nil {
		return err
	}
	sysfsdir, err = filepath.EvalSymlinks(sysfsdir)
	if err != nil {
		return err
	}
	provision := filepath.Join(sysfsdir, "provision")
	err = ioutil.WriteFile(provision, []byte(val), 0644)
	if errors.Is(err, syscall.EBUSY) {
		fmt.Fprintln(os.Stderr, "ignoring provision until xeth reload")
		err = nil
	}
	return err
}

// Connect socket and run channel service routines.
func Start(mux string, wg *sync.WaitGroup,
	stop <-chan struct{}) (task *Task, err error) {
	muxif, err := net.InterfaceByName(mux)
	if err != nil {
		return
	}

	atsockaddr, err := net.ResolveUnixAddr(unixpacket, "@"+mux)
	if err != nil {
		return
	}

	muxfd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW,
		int(ETH_P_ALL.Network()))
	if err != nil {
		return
	}
	if err = syscall.BindToDevice(muxfd, mux); err != nil {
		syscall.Close(muxfd)
		return
	}

	atsock, err := func(a *net.UnixAddr) (*net.UnixConn, error) {
		t := time.NewTicker(100 * time.Millisecond)
		defer t.Stop()
		for {
			s, err := net.DialUnix(unixpacket, nil, atsockaddr)
			if err == nil {
				return s, nil
			} else if !isEAGAIN(err) &&
				!errors.Is(err, syscall.ECONNREFUSED) {
				return nil, err
			}
			select {
			case <-stop:
				return nil, io.EOF
			case <-t.C:
			}
		}
	}(atsockaddr)

	loch := make(chan buffer, 4)
	hich := make(chan buffer, 4)
	rxch := make(chan Buffer, 1024)

	task = &Task{
		RxCh:  rxch,
		WG:    wg,
		Stop:  stop,
		sock:  atsock,
		loch:  loch,
		hich:  hich,
		muxfd: muxfd,
		muxsa: syscall.SockaddrLinklayer{
			Protocol: syscall.ETH_P_ARP,
			Ifindex:  muxif.Index,
			Hatype:   syscall.ARPHRD_ETHER,
		},
	}

	task.WG.Add(4)
	go task.goRx(rxch)
	go task.goTx(loch, hich)
	go task.goRawRx(rxch)
	go task.goClose()

	return
}

func kind(buf buffer) uint8 {
	return (*internal.MsgHeader)(buf.pointer()).Kind
}

// parse driver message and cache ifinfo in xid maps.
func Parse(buf Buffer) interface{} {
	defer Parsed.Inc()
	for _, b := range buf.bytes()[:14] {
		if b != 0 {
			return Frame{buf}
		}
	}
	defer buf.pool()
	switch k := kind(buf); k {
	case internal.MsgKindBreak:
		return Break{}
	case internal.MsgKindChangeUpperXid:
		msg := (*internal.MsgChangeUpperXid)(buf.pointer())
		lower := Xid(msg.Lower)
		upper := Xid(msg.Upper)
		if msg.Linking != 0 {
			return lower.join(upper)
		} else {
			return lower.quit(upper)
		}
	case internal.MsgKindEthtoolFlags:
		msg := (*internal.MsgEthtoolFlags)(buf.pointer())
		return Xid(msg.Xid).RxEthtoolFlags(msg.Flags)
	case internal.MsgKindEthtoolLinkModesSupported:
		msg := (*internal.MsgEthtoolLinkModes)(buf.pointer())
		return Xid(msg.Xid).RxSupported(msg.Modes)
	case internal.MsgKindEthtoolLinkModesAdvertising:
		msg := (*internal.MsgEthtoolLinkModes)(buf.pointer())
		return Xid(msg.Xid).RxAdvertising(msg.Modes)
	case internal.MsgKindEthtoolLinkModesLPAdvertising:
		msg := (*internal.MsgEthtoolLinkModes)(buf.pointer())
		return Xid(msg.Xid).RxLPAdvertising(msg.Modes)
	case internal.MsgKindEthtoolSettings:
		msg := (*internal.MsgEthtoolSettings)(buf.pointer())
		return Xid(msg.Xid).RxEthtoolSettings(msg)
	case internal.MsgKindFibEntry:
		msg := (*internal.MsgFibEntry)(buf.pointer())
		return fib4(msg)
	case internal.MsgKindFib6Entry:
		msg := (*internal.MsgFib6Entry)(buf.pointer())
		return fib6(msg)
	case internal.MsgKindIfa:
		msg := (*internal.MsgIfa)(buf.pointer())
		xid := Xid(msg.Xid)
		if msg.Event == internal.IFA_ADD {
			return xid.RxIP4Add(msg.Address, msg.Mask)
		} else {
			return xid.RxIP4Del(msg.Address, msg.Mask)
		}
	case internal.MsgKindIfa6:
		msg := (*internal.MsgIfa6)(buf.pointer())
		xid := Xid(msg.Xid)
		if msg.Event == internal.IFA_ADD {
			addr := []byte(msg.Address[:])
			length := int(msg.Length)
			return xid.RxIP6Add(addr, length)
		} else {
			return xid.RxIP6Del(msg.Address[:])
		}
	case internal.MsgKindIfInfo:
		msg := (*internal.MsgIfInfo)(buf.pointer())
		xid := Xid(msg.Xid)
		switch msg.Reason {
		case internal.IfInfoReasonNew:
			return RxIfInfo(msg)
		case internal.IfInfoReasonDump:
			return RxIfInfo(msg)
		case internal.IfInfoReasonDel:
			return RxDelete(xid)
		case internal.IfInfoReasonUp:
			return xid.RxUp()
		case internal.IfInfoReasonDown:
			return xid.RxDown()
		case internal.IfInfoReasonReg:
			return xid.RxReg(NetNs(msg.Net), msg.Ifindex)
		case internal.IfInfoReasonUnreg:
			return xid.RxUnreg(msg.Ifindex)
		case internal.IfInfoReasonFeatures:
			return xid.RxFeatures(msg.Features)
		}
	case internal.MsgKindNeighUpdate:
		msg := (*internal.MsgNeighUpdate)(buf.pointer())
		return neighbor(msg)
	case internal.MsgKindNetNsAdd:
		msg := (*internal.MsgNetNs)(buf.pointer())
		return NetNsAdd{NetNs(msg.Net)}
	case internal.MsgKindNetNsDel:
		msg := (*internal.MsgNetNs)(buf.pointer())
		return NetNsDel{NetNs(msg.Net)}
	}
	return nil
}

func Pool(v interface{}) {
	if method, found := v.(interface{ Pool() }); found {
		method.Pool()
	}
	if method, found := v.(interface{ pool() }); found {
		method.pool()
	}
}

// request fib dump
func (task *Task) DumpFib() {
	buf := newBuffer(internal.SizeofMsgDumpFibInfo)
	msg := (*internal.MsgHeader)(buf.pointer())
	msg.Set(internal.MsgKindDumpFibInfo)
	task.hich <- buf
}

// request ifinfo dump
func (task *Task) DumpIfInfo() {
	buf := newBuffer(internal.SizeofMsgDumpIfInfo)
	msg := (*internal.MsgHeader)(buf.pointer())
	msg.Set(internal.MsgKindDumpIfInfo)
	task.hich <- buf
}

// Send an exception frame to driver through raw socket.
func (task *Task) ExceptionFrame(b []byte) {
	// set priority so that the xeth will forward to the
	// respective upper device rather than it's port
	b[ETH_VLAN_TCI] |= VlanPrioMask >> 8
	syscall.Sendto(task.muxfd, b, 0, &task.muxsa)
}

// Send carrier change to driver through hi-priority channel.
func (task *Task) SetCarrier(xid Xid, on bool) {
	buf := newBuffer(internal.SizeofMsgCarrier)
	msg := (*internal.MsgCarrier)(buf.pointer())
	msg.Header.Set(internal.MsgKindCarrier)
	msg.Xid = uint32(xid)
	if on {
		msg.Flag = internal.CarrierOn
	} else {
		msg.Flag = internal.CarrierOff
	}
	task.hich <- buf
}

// Send ethtool stat change to driver through leaky-bucket channel.
func (task *Task) SetEthtoolStat(xid Xid, stat uint32, n uint64) {
	task.setStat(internal.MsgKindEthtoolStat, xid, stat, n)
}

// Send link stat change to driver through leaky-bucket channel.
func (task *Task) SetLinkStat(xid Xid, stat uint32, n uint64) {
	task.setStat(internal.MsgKindLinkStat, xid, stat, n)
}

func (task *Task) setStat(kind uint8, xid Xid, stat uint32, n uint64) {
	buf := newBuffer(internal.SizeofMsgStat)
	msg := (*internal.MsgStat)(buf.pointer())
	msg.Header.Set(kind)
	msg.Xid = uint32(xid)
	msg.Index = stat
	msg.Count = n
	task.hich <- buf
}

// Send speed change to driver through hi-priority channel.
func (task *Task) SetSpeed(xid Xid, mbps uint32) {
	buf := newBuffer(internal.SizeofMsgSpeed)
	msg := (*internal.MsgSpeed)(buf.pointer())
	msg.Header.Set(internal.MsgKindSpeed)
	msg.Xid = uint32(xid)
	msg.Mbps = mbps
	task.hich <- buf
}

// Wait for stop signal then shutdown and close socket
func (task *Task) goClose() {
	defer task.WG.Done()

	<-task.Stop

	if task.muxfd > 0 {
		syscall.Close(task.muxfd)
	}
	if task.sock == nil {
		return
	}
	sock := task.sock
	task.sock = nil
	f, err := sock.File()
	if err == nil {
		syscall.Shutdown(int(f.Fd()), SHUT_RDWR)
	}
	sock.Close()
}

func (task *Task) goRawRx(rxch chan<- Buffer) {
	defer task.WG.Done()

	rxbuf := make([]byte, internal.SizeofJumboFrame)
	for {
		select {
		case <-task.Stop:
			return
		default:
		}
		// FIXME timeout through sockopt or select
		n, from, err := syscall.Recvfrom(task.muxfd, rxbuf, 0)
		if err != nil {
			e, ok := err.(*os.SyscallError)
			if !ok || e.Err.Error() != "EOF" {
				task.RxErr = err
			}
			return
		}
		sa, ok := from.(*syscall.SockaddrLinklayer)
		if ok && sa.Ifindex == task.muxsa.Ifindex {
			rxch <- cloneBuffer(rxbuf[:n])
			Cloned.Inc()
		}
	}
}

func (task *Task) goRx(rxch chan<- Buffer) {
	defer task.WG.Done()

	const minrxto = 10 * time.Millisecond
	const maxrxto = 320 * time.Millisecond

	rxto := minrxto
	rxbuf := make([]byte, PageSize, PageSize)
	rxoob := make([]byte, PageSize, PageSize)
	ptr := unsafe.Pointer(&rxbuf[0])
	h := (*internal.MsgHeader)(ptr)

	defer func() {
		rxbuf = rxbuf[:0]
		rxoob = rxoob[:0]
		close(rxch)
	}()

	for {
		select {
		case <-task.Stop:
			return
		default:
		}
		task.RxErr = task.sock.SetReadDeadline(time.Now().Add(rxto))
		if task.RxErr != nil {
			break
		}
		n, noob, flags, addr, err := task.sock.ReadMsgUnix(rxbuf, rxoob)
		_ = noob
		_ = flags
		_ = addr
		select {
		case <-task.Stop:
			return
		default:
		}
		if n == 0 || isTimeout(err) {
			if rxto < maxrxto {
				rxto *= 2
			}
		} else if err != nil {
			e, ok := err.(*os.SyscallError)
			if !ok || e.Err.Error() != "EOF" {
				task.RxErr = err
			}
			break
		} else if task.RxErr = h.Validate(rxbuf[:n]); task.RxErr != nil {
			break
		} else {
			rxto = minrxto
			rxch <- cloneBuffer(rxbuf[:n])
			Cloned.Inc()
		}
	}
}

func (task *Task) goTx(loch, hich <-chan buffer) {
	defer task.WG.Done()
	for task.TxErr == nil {
		select {
		case <-task.Stop:
			return
		case buf, ok := <-hich:
			if ok {
				task.TxErr = task.tx(buf, 0)
			} else {
				return
			}
		case buf, ok := <-loch:
			if ok {
				task.TxErr = task.tx(buf, 10*time.Millisecond)
			} else {
				return
			}
		}
	}
}

// Send through low-priority, leaky-bucket.
func (task *Task) queue(buf buffer) {
	select {
	case task.loch <- buf:
	default:
		buf.pool()
		Dropped.Inc()
	}
}

func (task *Task) tx(buf buffer, timeout time.Duration) error {
	var oob []byte
	var dl time.Time
	defer buf.pool()
	if task.sock == nil {
		return io.EOF
	}
	if timeout != time.Duration(0) {
		dl = time.Now().Add(timeout)
	}
	err := task.sock.SetWriteDeadline(dl)
	if err != nil {
		return err
	}
	_, _, err = task.sock.WriteMsgUnix(buf.bytes(), oob, nil)
	if err == nil {
		Sent.Inc()
		if kind(buf) == internal.MsgKindCarrier {
			msg := (*internal.MsgCarrier)(buf.pointer())
			xid := Xid(msg.Xid)
			l := LinkOf(xid)
			if l != nil {
				l.LinkUp(msg.Flag == internal.CarrierOn)
			}
		}
	}
	return err
}

func isEAGAIN(err error) bool {
	if err != nil {
		if operr, ok := err.(*net.OpError); ok {
			if oserr, ok := operr.Err.(*os.SyscallError); ok {
				if oserr.Err == syscall.EAGAIN {
					return true
				}
			}
		}
	}
	return false
}

func isTimeout(err error) bool {
	if err != nil {
		if op, ok := err.(*net.OpError); ok {
			return op.Timeout()
		}
	}
	return false
}
