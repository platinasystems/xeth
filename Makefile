INSTALL=/usr/bin/install

default:

distclean:
	$(MAKE) -C src clean
	rm -f debian/debhelper-build-stamp debian/files debian/*.substvars
	rm -f debian/*.debhelper debian/*.log
	rm -rf debian/.debhelper debian/xeth

bindeb-pkg:
	debuild -uc -us --lintian-opts --profile debian

.PHONY: default bindeb-pkg

