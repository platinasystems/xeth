INSTALL=/usr/bin/install

default:

distclean:
	rm -f debian/debhelper-build-stamp debian/files debian/*.substvars debian/*.debhelper debian/*.log
	rm -rf debian/.debhelper debian/xeth

bindeb-pkg:
	debuild -uc -us --lintian-opts --profile debian

.PHONY: default bindeb-pkg

src/xeth.ko:
	$(MAKE) -C src
