
erfa_version   := 1.4.0
erfa_tarball   := erfa-$(erfa_version).tar.gz
erfa_configure := erfa-$(erfa_version)/configure
prefix         := $(PWD)/deps
liberfa_a      := $(prefix)/lib/liberfa.a

telescope-control-system: $(liberfa_a) *.go
	CGO_CPPFLAGS="-I$(prefix)/include" \
	CGO_LDFLAGS="$(prefix)/lib/liberfa.a -lm" \
	    go build -o $@

$(liberfa_a): $(erfa_configure)
	cd erfa-$(erfa_version) \
	    && ./configure --prefix="$(prefix)" --disable-shared \
	    && make install

$(erfa_configure): $(erfa_tarball)
	tar -xaf "$(erfa_tarball)"
	touch "$(erfa_configure)"

$(erfa_tarball):
	wget "https://github.com/liberfa/erfa/releases/download/v$(erfa_version)/$(erfa_tarball)"

