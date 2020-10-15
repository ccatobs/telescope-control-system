FROM golang:1
WORKDIR /go/src/github.com/ccatp/telescope-control-system

ADD github_ssh_key_ed25519 /root/.ssh/id_ed25519
RUN chmod 700 /root/.ssh/id_ed25519
RUN echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config
RUN git config --global url.ssh://git@github.com/ccatp/.insteadOf https://github.com/ccatp/

COPY . .
RUN ./build-deps
RUN go get -d -v
RUN go test -v
RUN go install -a -v -tags netgo -ldflags=-extldflags=-static

FROM scratch
COPY --from=0 /go/bin/telescope-control-system /
CMD ["/telescope-control-system"]