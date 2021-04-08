ifneq ($(DH_VERBOSE),)
	V ?= 1
endif

ifneq ($(V),)
	MAKEFLAGS += V=$(V)
else
	Q = @
	MAKEFLAGS += --no-print-directory
endif

ifneq ($(KDIR),)
	MAKEFLAGS += KDIR=$(KDIR)
endif

# subject, verb, object(s)
svo = $(if $(Q),$(info $1 $3))$(Q)$2 $3

define clean
$(if $(files),$(call svo, CLEAN, rm, $(files)))
$(if $(dirs),$(call svo, CLEAN, rm -r, $(dirs)))
endef

define debuild
$(call svo, DEBUILD, debuild -uc -us --lintian-opts --profile debian)
endef

define mk-dkms
$(call svo, DKMS, $(MAKE) -C dkms, $@)
endef

default:

bindeb-pkg: ; $(debuild)

clean: files = $(wildcard README.html)
clean:
	$(mk-dkms)
	$(clean)

docs: README.html

modules modules_install: ; $(mk-dkms)

.PHONY: default bindeb-pkg clean docs modules modules_install

%.html: %.md ; pandoc --from gfm --to html -o $@ $<
