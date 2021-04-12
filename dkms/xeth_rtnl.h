/**
 * SPDX-License-Identifier: GPL-2.0
 * Copyright(c) 2018-2020 Platina Systems, Inc.
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */

#ifndef __NET_ETHERNET_XETH_RTNL_H
#define __NET_ETHERNET_XETH_RTNL_H

#include <net/rtnetlink.h>

extern struct rtnl_link_ops xeth_vlan_lnko;

static inline int xeth_rtnl_unlock(int val)
{
	rtnl_unlock();
	return val;
}

#endif /* __NET_ETHERNET_XETH_RTNL_H */
