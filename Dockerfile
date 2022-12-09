FROM golang:1
WORKDIR /go/src/github.com/ccatp/telescope-control-system

RUN mkdir /root/.ssh/
#COPY ~/.ssh/id_rsa /root/.ssh/id_ed25519
ARG SSH_PRIVATE_KEY
RUN echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id_ed25519
RUN chmod 700 /root/.ssh/id_ed25519
RUN echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config
RUN git config --global url.ssh://git@github.com/ccatp/.insteadOf https://github.com/ccatp/

COPY . .
RUN ./build-deps
RUN GOPRIVATE=github.com/ccatp go get -d -v
RUN go test -v
RUN go install -a -v -tags netgo -ldflags=-extldflags=-static

FROM scratch
COPY --from=0 /go/bin/telescope-control-system /
CMD ["/telescope-control-system"]
