ifeq ($(or $(V), $(DH_VERBOSE)),)
Q = @
MAKEFLAGS += --no-print-directory
endif

# subject, verb, object(s)
svo = $(if $(Q),$(info $1 $3))$(Q)$2 $3

install: docs := README.md
install: docs-dest := $(DESTDIR)/usr/share/doc/xeth
install: figures := $(wildcard figures/*.svg)
install: figures-dest := $(DESTDIR)/usr/share/doc/xeth/figures
install: src := $(filter-out %.mod.c,$(wildcard src/Makefile src/*.[ch]))
install: src-dest := $(DESTDIR)/usr/src/xeth$(if $(VERSION),-)$(VERSION)

INSTALL=/usr/bin/install

define install
$(call svo, INSTALL, $(INSTALL) -m 0644 -D -t $(src-dest), $(src))
$(call svo, INSTALL, $(INSTALL) -m 0644 -D -t $(docs-dest), $(docs))
$(call svo, INSTALL, $(INSTALL) -m 0644 -D -t $(figures-dest), $(figures))
endef

clean: files := $(wildcard README.html)
distclean: dirs := $(wildcard debian/.debhelper debian/xeth)
distclean: files := $(wildcard debian/debhelper-build-stamp debian/files\
	debian/*.substvars debian/*.debhelper debian/*.log)

define clean
$(if $(dirs),$(call svo, CLEAN, rm -r, $(dirs)))
$(if $(files),$(call svo, CLEAN, rm, $(files)))
endef

default:

bindeb-pkg:
	$(call svo, DEBUILD, debuild -uc -us --lintian-opts --profile debian)

clean:
	$(Q)$(MAKE) -C src Q=$(Q) $@
	$(clean)

distclean:
	$(clean)

docs: README.html

install:
	$(install)

modules:
	$(Q)$(MAKE) -C src Q=$(Q) $@

.PHONY: default bindeb-pkg clean distclean docs install modules

%.html: %.md
	pandoc --from gfm --to html -o $@ $<
