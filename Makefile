INSTALL=/usr/bin/install

default:

doc: README.html

distclean:
	$(MAKE) -C src clean
	rm -f README.html
	rm -f debian/debhelper-build-stamp debian/files debian/*.substvars
	rm -f debian/*.debhelper debian/*.log
	rm -rf debian/.debhelper debian/xeth

bindeb-pkg:
	debuild -uc -us --lintian-opts --profile debian

.PHONY: default doc bindeb-pkg

%.html: %.md
	pandoc --from gfm --to html --standalone -T README $< --output $@
