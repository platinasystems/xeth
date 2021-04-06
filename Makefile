ifeq ($(or $(V), $(DH_VERBOSE)),)
	Q = @
	MAKEFLAGS += --no-print-directory
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

define mk-src
$(call svo, SRC, $(MAKE) -C src $(if $(KDIR),KDIR=$(KDIR)), $@)
endef

default:

bindeb-pkg: ; $(debuild)

clean: files = $(wildcard README.html)
clean:
	$(mk-src)
	$(clean)

docs: README.html

modules modules_install: ; $(mk-src)

.PHONY: default bindeb-pkg clean docs modules modules_install

%.html: %.md ; pandoc --from gfm --to html -o $@ $<
