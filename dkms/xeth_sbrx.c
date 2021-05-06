/**
 * SPDX-License-Identifier: GPL-2.0
 * Copyright(c) 2018-2019 Platina Systems, Inc.
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */

static bool xeth_sbrx_is_msg(void *data)
{
	struct xeth_msg *msg = data;

	return	msg->header.z64 == 0 &&
		msg->header.z32 == 0 &&
		msg->header.z16 == 0;
}

static void xeth_sbrx_carrier(struct net_device *mux,
			      struct xeth_msg_carrier *msg)
{
	struct xeth_proxy *proxy = xeth_mux_proxy_of_xid(mux, msg->xid);
	if (proxy && proxy->kind == XETH_DEV_KIND_PORT)
		xeth_mux_change_carrier(mux, proxy->nd,
					msg->flag == XETH_CARRIER_ON);
	else
		xeth_mux_inc_sbrx_invalid(mux);
}

static void xeth_sbrx_et_stat(struct net_device *mux,
			      struct xeth_msg_stat *msg)
{
	struct xeth_proxy *proxy = xeth_mux_proxy_of_xid(mux, msg->xid);
	if (proxy && proxy->kind == XETH_DEV_KIND_PORT)
		xeth_port_ethtool_stat(proxy->nd, msg->index, msg->count);
}

static void xeth_sbrx_link_stat(struct net_device *mux,
				struct xeth_msg_stat *msg)
{
	struct xeth_proxy *proxy = xeth_mux_proxy_of_xid(mux, msg->xid);
	xeth_proxy_link_stat(proxy->nd, msg->index, msg->count);
}

static void xeth_sbrx_speed(struct net_device *mux,
			    struct xeth_msg_speed *msg)
{
	struct xeth_proxy *proxy = xeth_mux_proxy_of_xid(mux, msg->xid);
	if (proxy && proxy->kind == XETH_DEV_KIND_PORT)
		xeth_port_speed(proxy->nd, msg->mbps);
}

int xeth_sbrx_msg(struct net_device *mux, void *v, size_t n)
{
	struct xeth_msg_header *msg = v;
	if (n < sizeof(*msg) || !xeth_sbrx_is_msg(msg))
		return -EINVAL;
	if (msg->version != XETH_MSG_VERSION)
		return -EINVAL;
	switch (msg->kind) {
	case XETH_MSG_KIND_DUMP_IFINFO:
		xeth_mux_dump_all_ifinfo(mux);
		xeth_sbtx_break(mux);
		xeth_nd_prif_err(mux, xeth_nb_start_netdevice(mux));
		xeth_nd_prif_err(mux, xeth_nb_start_inetaddr(mux));
		break;
	case XETH_MSG_KIND_DUMP_FIBINFO:
		xeth_nb_start_all_fib(mux);
		xeth_sbtx_break(mux);
		xeth_nd_prif_err(mux, xeth_nb_start_netevent(mux));
		break;
	case XETH_MSG_KIND_CARRIER:
		xeth_sbrx_carrier(mux, v);
		break;
	case XETH_MSG_KIND_ETHTOOL_STAT:
		xeth_sbrx_et_stat(mux, v);
		break;
	case XETH_MSG_KIND_LINK_STAT:
		xeth_sbrx_link_stat(mux, v);
		break;
	case XETH_MSG_KIND_SPEED:
		xeth_sbrx_speed(mux, v);
		break;
	default:
		xeth_mux_inc_sbrx_invalid(mux);
		return -EINVAL;
	}
	return 0;
}
