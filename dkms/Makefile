KERNELRELEASE ?= $(shell uname -r)
MODLIB = $(INSTALL_MOD_PATH)/lib/modules/$(KERNELRELEASE)
KDIR ?= $(MODLIB)/build

modules modules_install clean:
	$(if $(wildcard $(KDIR)),$(MAKE) -C $(KDIR) M=$(CURDIR) $@,true)

.PHONY: clean modules modules_install
