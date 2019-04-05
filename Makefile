
erfa_version   := 1.4.0
erfa_tarball   := erfa-$(erfa_version).tar.gz
erfa_configure := erfa-$(erfa_version)/configure
prefix         := $(PWD)/deps
liberfa_a      := $(prefix)/lib/liberfa.a
cgo_vars       := CGO_CPPFLAGS="-I$(prefix)/include" \
                  CGO_LDFLAGS="$(prefix)/lib/liberfa.a -lm"

telescope-control-system: $(liberfa_a) *.go
	$(cgo_vars) go build -o $@

.PHONY: test
test:
	$(cgo_vars) go test

$(liberfa_a): $(erfa_configure)
	cd erfa-$(erfa_version) \
	    && ./configure --prefix="$(prefix)" --disable-shared \
	    && make install

$(erfa_configure): $(erfa_tarball)
	tar -xzf "$(erfa_tarball)"
	touch "$(erfa_configure)"

$(erfa_tarball):
	curl -LO "https://github.com/liberfa/erfa/releases/download/v$(erfa_version)/$(erfa_tarball)"

