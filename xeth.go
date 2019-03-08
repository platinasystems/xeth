/* XETH driver sideband control.
 *
 * Copyright(c) 2018 Platina Systems, Inc.
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
	"github.com/platinasystems/xeth/dbgxeth"

	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const netname = "unixpacket"

var (
	Count struct {
		Tx struct {
			Sent, Dropped, Retried uint64
		}
	}
	// Receive message channel feed from sock by gorx
	RxCh <-chan []byte

	xeth struct {
		name string
		addr *net.UnixAddr
		sock *net.UnixConn

		rxch chan []byte
		txch chan []byte
	}
)

// Connect to @xeth socket and run channel service routines
// driver :: XETH driver name (e.g. "platina-mk1")
func Start(driver string) error {
	var err error
	xeth.name = driver
	xeth.addr, err = net.ResolveUnixAddr(netname, "@xeth")
	if err != nil {
		return err
	}
	for {
		xeth.sock, err = net.DialUnix(netname, nil, xeth.addr)
		if err == nil {
			break
		}
		if !isEAGAIN(err) {
			return err
		}
	}
	xeth.rxch = make(chan []byte, 4)
	xeth.txch = make(chan []byte, 4)
	Interface.index.validate()
	Interface.dir.validate()
	RxCh = xeth.rxch
	go gorx()
	go gotx()

	// load Interface cache
	DumpIfinfo()
	UntilBreak(func(buf []byte) error {
		return nil
	})
	dbgxeth.Chan.Log("Done")

	return nil
}

// Close @xeth socket and shutdown service routines
func Stop() {
	const (
		SHUT_RD = iota
		SHUT_WR
		SHUT_RDWR
	)
	if xeth.sock == nil {
		return
	}
	close(xeth.txch)
	sock := xeth.sock
	xeth.sock = nil
	if f, err := sock.File(); err == nil {
		syscall.Shutdown(int(f.Fd()), SHUT_RDWR)
	}
	sock.Close()
	Interface.indexes = Interface.indexes[:0]
	Interface.index.unsetAll()
	Interface.dir.unsetAll()
}

// Return driver name (e.g. "platina-mk1")
func String() string { return xeth.name }

// Send carrier state change message
func Carrier(ifindex int32, flag uint8) error {
	buf := Pool.Get(SizeofMsgCarrier)
	defer Pool.Put(buf)
	msg := ToMsgCarrier(buf)
	msg.Kind = uint8(XETH_MSG_KIND_CARRIER)
	msg.Ifindex = ifindex
	msg.Flag = flag
	return tx(buf, 0)
}

// Send DumpFib request
func DumpFib() error {
	dbgxeth.Chan.Log("DumpFib")
	buf := Pool.Get(SizeofMsgDumpFibinfo)
	defer Pool.Put(buf)
	msg := ToMsg(buf)
	msg.Kind = XETH_MSG_KIND_DUMP_FIBINFO
	return tx(buf, 0)
}

// Send DumpIfinfo request then flush RxCh until break to cache ifinfos
func CacheIfinfo() {
	if err := DumpIfinfo(); err == nil {
		UntilBreak(func(buf []byte) error { return nil })
	}
}

// Send DumpIfinfo request
func DumpIfinfo() error {
	dbgxeth.Chan.Log("DumpIfinfo")
	buf := Pool.Get(SizeofMsgDumpIfinfo)
	defer Pool.Put(buf)
	msg := ToMsg(buf)
	msg.Kind = XETH_MSG_KIND_DUMP_IFINFO
	return tx(buf, 0)
}

// Send stat update message
func SetStat(ifindex int32, stat string, count uint64) error {
	var statindex uint64
	var kind uint8
	if linkstat, found := LinkStatOf(stat); found {
		kind = uint8(XETH_MSG_KIND_LINK_STAT)
		statindex = uint64(linkstat)
	} else if ethtoolstat, found := EthtoolStatOf(stat); found {
		kind = uint8(XETH_MSG_KIND_ETHTOOL_STAT)
		statindex = uint64(ethtoolstat)
	} else {
		return fmt.Errorf("%q unknown", stat)
	}
	buf := Pool.Get(SizeofMsgStat)
	defer Pool.Put(buf)
	msg := ToMsgStat(buf)
	msg.Kind = kind
	msg.Ifindex = ifindex
	msg.Index = statindex
	msg.Count = count
	return tx(buf, 10*time.Millisecond)
}

// Send speed change message
func Speed(index int, count uint64) error {
	buf := Pool.Get(SizeofMsgSpeed)
	defer Pool.Put(buf)
	msg := ToMsgSpeed(buf)
	msg.Kind = uint8(XETH_MSG_KIND_SPEED)
	msg.Ifindex = int32(index)
	msg.Mbps = uint32(count)
	return tx(buf, 0)
}

// Send through leaky bucket
func Tx(buf []byte) {
	msg := Pool.Get(len(buf))
	copy(msg, buf)
	kind := KindOf(msg)
	if kind == XETH_MSG_KIND_CARRIER {
		wait := true
		for wait {
			select {
			case xeth.txch <- msg:
				Count.Tx.Sent++
				wait = false
			case <-time.After(1 * time.Second):
				Count.Tx.Retried++
			}
		}
	} else {
		select {
		case xeth.txch <- msg:
			Count.Tx.Sent++
		default:
			Count.Tx.Dropped++
			Pool.Put(msg)
		}
	}
}

func UntilBreak(f func([]byte) error) error {
	dbgxeth.Chan.Logf("start UntilBreak len(RxCh)=%v", len(RxCh))
	for buf := range RxCh {
		if KindOf(buf) == XETH_MSG_KIND_BREAK {
			Pool.Put(buf)
			break
		}
		err := f(buf)
		Pool.Put(buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func UntilSig(sig <-chan os.Signal, f func([]byte) error) error {
	for {
		select {
		case <-sig:
			return nil
		case buf, ok := <-RxCh:
			if !ok {
				return nil
			}
			err := f(buf)
			Pool.Put(buf)
			if err != nil {
				return err
			}
		}
	}
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

func isEOF(err error) bool {
	if err != nil {
		if op, ok := err.(*net.OpError); ok {
			return op.Err == io.EOF
		}
	}
	return false
}

func gorx() {
	dbgxeth.Chan.Log("start")
	const minrxto = 10 * time.Millisecond
	const maxrxto = 320 * time.Millisecond
	rxto := minrxto
	rxbuf := Pool.Get(PageSize)
	defer Pool.Put(rxbuf)
	rxoob := Pool.Get(PageSize)
	defer Pool.Put(rxoob)
	defer close(xeth.rxch)

	start := time.Now()
	connected := true
	for xeth.sock != nil {
		err := xeth.sock.SetReadDeadline(time.Now().Add(rxto))
		if err != nil {
			fmt.Fprintln(os.Stderr, "xeth set rx deadline", err)
			break
		}
		n, noob, flags, addr, err :=
			xeth.sock.ReadMsgUnix(rxbuf, rxoob)
		_ = noob
		_ = flags
		_ = addr
		if dbgxeth.Chan > 0 {
			if isEOF(err) {
				if connected {
					LogEvent(fmt.Sprintf("gorx() lost connection to socket"))
					connected = false
				}
			} else {
				if !connected {
					LogEvent(fmt.Sprintf("gorx() reconnected to socket"))
					connected = true
				}
			}
		}
		if n == 0 || isTimeout(err) {
			if rxto < maxrxto {
				rxto *= 2
			}
			if dbgxeth.Chan > 0 {
				if time.Since(start) > 10*time.Second {
					LogReplace(fmt.Sprintf("RxCh length = %v, read socket timed out for last 10 seconds", len(xeth.rxch)))
					start = time.Now()
				}
			}
		} else if err == nil {
			rxto = minrxto
			kind := KindOf(rxbuf[:n])
			if err = kind.validate(rxbuf[:n]); err != nil {
				fmt.Fprintln(os.Stderr, "xeth rx", err)
				break
			}
			kind.cache(rxbuf[:n])
			msg := Pool.Get(n)
			copy(msg, rxbuf[:n])
			if dbgxeth.Chan > 0 {
				LogEvent(fmt.Sprintf("From kernel to RxCh (len %v) msg kind: %v", len(xeth.rxch), KindOf(msg)))
				start = time.Now()
			}
			done := false
			for !done {
				select {
				case xeth.rxch <- msg:
					done = true
				case <-time.After(1 * time.Second):
					if dbgxeth.Chan > 0 {
						LogEvent(fmt.Sprint("wait for chan to free up"))
					}
				}
			}
		} else {
			e, ok := err.(*os.SyscallError)
			if !ok || e.Err.Error() != "EOF" {
				fmt.Fprintln(os.Stderr, "xeth rx", err)
			}
			break
		}
	}
	dbgxeth.Chan.Log("quit")
}

func gotx() {
	dbgxeth.Chan.Log("start")
	for msg := range xeth.txch {
		tx(msg, 10*time.Millisecond)
		Pool.Put(msg)
	}
}

func tx(buf []byte, timeout time.Duration) error {
	var oob []byte
	var dl time.Time
	if xeth.sock == nil {
		return io.EOF
	}
	if timeout != time.Duration(0) {
		dl = time.Now().Add(timeout)
	}
	if dbgxeth.Chan > 0 {
		if KindOf(buf) == XETH_MSG_KIND_CARRIER {
			updown := "down"
			ptr := unsafe.Pointer(&buf[0])
			msg := (*MsgCarrier)(ptr)
			xethif := Interface.Indexed(msg.Ifindex)
			ifname := xethif.Ifinfo.Name
			if msg.Flag > 0 {
				updown = "up"
			}
			LogEvent(fmt.Sprintf("Write to kernel msg %v %v %v", KindOf(buf), ifname, updown))
		}
	}
	err := xeth.sock.SetWriteDeadline(dl)
	if err != nil {
		if dbgxeth.Chan > 0 {
			LogEvent(fmt.Sprintf("SetWriteDeadLine %v error %v", KindOf(buf), err))
		}
		return err
	}
	_, _, err = xeth.sock.WriteMsgUnix(buf, oob, nil)
	if dbgxeth.Chan > 0 {
		if err != nil {
			LogEvent(fmt.Sprintf("Write to kernel %v error %v", KindOf(buf), err))
		}
	}
	return err
}
