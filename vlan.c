/* XETH driver's VLAN encapsulations.
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

#include <linux/etherdevice.h>
#include <linux/if_vlan.h>
#include <net/rtnetlink.h>

#include "xeth.h"
#include "debug.h"

static inline bool xeth_vlan_is_8021X(__be16 proto)
{
	return  proto == cpu_to_be16(ETH_P_8021Q) ||
		proto == cpu_to_be16(ETH_P_8021AD);
}

/* If tagged, pop dev index  and skb priority from outer VLAN;
 * otherwise, RX_HANDLER_PASS through to upper protocols.
 */
static rx_handler_result_t xeth_vlan_rx(struct sk_buff **pskb)
{
	struct sk_buff *skb = *pskb;
	struct net_device *nd;
	u16 vid;
	int res;
	
	if (!xeth_vlan_is_8021X(skb->vlan_proto))
		return xeth_debug_netdev_val(skb->dev,
					     "%d, vlan_proto: 0x%04hx",
					     RX_HANDLER_PASS,
					     be16_to_cpu(skb->vlan_proto));
	skb->priority =
		(typeof(skb->priority))(skb->vlan_tci >> VLAN_PRIO_SHIFT);
	vid = skb->vlan_tci & VLAN_VID_MASK;
	nd = to_xeth_nd(vid);
	if (nd == NULL) {
		kfree_skb(skb);
		xeth_debug_netdev_val(skb->dev, "%d unknown", vid);
		return RX_HANDLER_CONSUMED;
	}
	if (skb->protocol == cpu_to_be16(ETH_P_8021Q)) {
		struct vlan_hdr *iv = (struct vlan_hdr *)skb->data;
		skb->vlan_proto = skb->protocol;
		skb->vlan_tci = VLAN_TAG_PRESENT | be16_to_cpu(iv->h_vlan_TCI);
		skb->protocol = iv->h_vlan_encapsulated_proto;
		skb_pull_rcsum(skb, VLAN_HLEN);
		/* make DST, SRC address precede encapsulated protocol */
		memcpy(skb->data-ETH_HLEN, skb->data-(ETH_HLEN+VLAN_HLEN),
		       2*ETH_ALEN);
	} else {
		skb->vlan_proto = 0;
		skb->vlan_tci = 0;
	}
	skb_push(skb, ETH_HLEN);
	skb->dev = nd;
	xeth_debug_hex_dump(skb);
	res = xeth_debug_netdev_true_val(nd, "%d", dev_forward_skb(nd, skb));
	if (res == NET_RX_DROP)
		atomic_long_inc(&nd->rx_dropped);
	return RX_HANDLER_CONSUMED;
}

static ssize_t xeth_vlan_sb(struct sk_buff *skb)
{
	struct vlan_ethhdr *ve = (struct vlan_ethhdr *)skb->data;
	u16 vid;
	char eaddrs[2*ETH_ALEN];
	ssize_t n = skb->len;

	if (!xeth_debug_false_val("%d", xeth_vlan_is_8021X(ve->h_vlan_proto)))
		return -EINVAL;
	vid = be16_to_cpu(ve->h_vlan_TCI) & VLAN_VID_MASK;
	skb->dev = to_xeth_nd(vid);
	if (skb->dev == NULL) {
		xeth_debug_val("%d unknown", vid);
		return -ENOENT;
	}
	memcpy(eaddrs, skb->data, 2*ETH_ALEN);
	if (xeth_vlan_is_8021X(ve->h_vlan_encapsulated_proto)) {
		const size_t vesz = sizeof(struct vlan_ethhdr);
		struct vlan_hdr *iv = (struct vlan_hdr *)(skb->data + vesz);
		skb->vlan_proto = ve->h_vlan_encapsulated_proto;
		skb->vlan_tci = VLAN_TAG_PRESENT | be16_to_cpu(iv->h_vlan_TCI);
		skb_pull(skb, 2*VLAN_HLEN);
	} else {
		skb_pull(skb, VLAN_HLEN);
	}
	/* restore mac addrs to beginning of de-encapsulated frame */
	memcpy(skb->data, eaddrs, 2*ETH_ALEN);
	xeth_debug_hex_dump(skb);
	skb->protocol = eth_type_trans(skb, skb->dev);
	skb_postpull_rcsum(skb, eth_hdr(skb), ETH_HLEN);
	return (xeth_debug_netdev_true_val(skb->dev, "%d", netif_rx(skb))
		== NET_RX_DROP) ? -EBUSY : n;
}

/* Push outer VLAN tag with xeth's vid and skb's priority. */
static netdev_tx_t xeth_vlan_tx(struct sk_buff *skb, struct net_device *nd)
{
	struct xeth_priv *priv = netdev_priv(nd);
	struct net_device *iflink = xeth_priv_iflink(priv);
	u16 tpid = xeth_vlan_is_8021X(skb->protocol)
		? cpu_to_be16(ETH_P_8021AD)
		: cpu_to_be16(ETH_P_8021Q);
	u16 pcp = (u16)(skb->priority) << VLAN_PRIO_SHIFT;
	u16 tci = pcp | priv->id;

	if (iflink == NULL) {
		kfree_skb(skb);
		atomic_long_inc(&nd->tx_dropped);
		return xeth_debug_netdev_val(nd, "0x%02x, no iflink",
					     NETDEV_TX_OK);
	}
	skb = vlan_insert_tag_set_proto(skb, tpid, tci);
	if (!skb) {
		atomic_long_inc(&nd->tx_dropped);
		return xeth_debug_val("%d, couldn't insert tag", NETDEV_TX_OK);
	}
	xeth_debug_hex_dump(skb);
	skb->dev = iflink;
	return xeth_debug_netdev_true_val(iflink, "%d", dev_queue_xmit(skb));
}

void xeth_vlan_init(void)
{
	xeth.ops.rx_handler         = xeth_vlan_rx;
	xeth.ops.side_band_rx       = xeth_vlan_sb;
	xeth.ops.ndo.ndo_start_xmit = xeth_vlan_tx;
}

void xeth_vlan_exit(void)
{
	xeth.ops.rx_handler         = NULL;
	xeth.ops.side_band_rx       = NULL;
	xeth.ops.ndo.ndo_start_xmit = NULL;
}