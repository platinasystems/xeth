obj-m += xeth.o

xeth-y := xeth_mod.o
xeth-y += xeth_mux.o
xeth-y += xeth_proxy.o
xeth-y += xeth_bridge.o
xeth-y += xeth_lb.o
xeth-y += xeth_lag.o
xeth-y += xeth_vlan.o
xeth-y += xeth_port.o
xeth-y += xeth_nb.o
xeth-y += xeth_sbrx.o
xeth-y += xeth_sbtx.o
xeth-y += xeth_qsfp.o

xeth-srctree-uapi := $(srctree)/include/uapi/linux/xeth.h
xeth-uapi := $(or $(wildcard $(src)/xeth_uapi.h),$(xeth-srctree-uapi))
ccflags-y += --include=$(xeth-uapi)

KERNELRELEASE ?= $(shell uname -r)
MODLIB = $(INSTALL_MOD_PATH)/lib/modules/$(KERNELRELEASE)
KDIR ?= $(MODLIB)/build

modules modules_install clean:
	$(if $(wildcard $(KDIR)),$(MAKE) -C $(KDIR) M=$(CURDIR) $@,true)

.PHONY: clean modules modules_install
