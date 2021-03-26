ifneq ($(KERNELRELEASE),)

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

else

KVERSION ?= `uname -r`
KDIR ?= /lib/modules/$(KVERSION)/build

xeth-kmod:
	$(MAKE) -C $(KDIR) M=$$PWD

clean:
	rm -f *.o *.a *.cmd *.ko .xeth*.cmd Module.symvers modules.order xeth.mod xeth.mod.c

.PHONY: xeth-kmod clean

endif
