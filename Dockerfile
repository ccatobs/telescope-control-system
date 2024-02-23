FROM golang:1
WORKDIR /go/src/github.com/ccatobs/telescope-control-system

RUN mkdir /root/.ssh/
ARG SSH_PRIVATE_KEY
RUN echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id_ed25519
RUN chmod 700 /root/.ssh/id_ed25519
RUN echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config
RUN git config --global url.ssh://git@github.com/ccatobs/.insteadOf https://github.com/ccatobs/

COPY . .
RUN ./build-deps
RUN GOPRIVATE=github.com/ccatobs go get -d -v
RUN go test -v
RUN go install -a -v -tags netgo -ldflags=-extldflags=-static

FROM scratch
COPY --from=0 /go/bin/telescope-control-system /
CMD ["/telescope-control-system"]
