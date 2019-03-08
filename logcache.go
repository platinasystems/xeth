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
	"github.com/platinasystems/elib/cli"
	"github.com/platinasystems/xeth/dbgxeth"

	"fmt"
	"time"
)

const MaxEventHistory = 1000

var eventHistory []string

type previous struct {
	event string
	count uint64
}

func (p *previous) repeat(s string) {
	p.event = s
	p.count++
}
func (p *previous) new(s string) {
	p.event = s
	p.count = 0
}

var prevEvent previous

func LogEvent(s string) {
	if len(eventHistory) > MaxEventHistory {
		eventHistory = eventHistory[1:]
	}
	prevEvent.new(s)
	s = fmt.Sprintf("%v ", time.Now().Format(time.UnixDate)) + s
	dbgxeth.Syslog.Log(s)
	eventHistory = append(eventHistory, s)
}

func LogReplace(s string) {
	var s2 string
	if len(eventHistory) > MaxEventHistory {
		eventHistory = eventHistory[1:]
	}
	if prevEvent.event == s {
		eventHistory = eventHistory[:len(eventHistory)-1]
		prevEvent.repeat(s)
		s2 = fmt.Sprintf("%v %s repeated %dx", time.Now().Format(time.UnixDate), s, prevEvent.count+1)
	} else {
		prevEvent.new(s)
		s2 = fmt.Sprintf("%v ", time.Now().Format(time.UnixDate)) + s
	}
	dbgxeth.Syslog.Log(s2)
	eventHistory = append(eventHistory, s2)
}

func ShowLastEvents(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	for _, line := range eventHistory {
		fmt.Fprintln(w, line)
	}
	return nil
}
