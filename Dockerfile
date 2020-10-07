FROM golang:1
WORKDIR /go/src/github.com/ccatp/telescope-control-system
COPY . .
RUN ./build-deps
RUN go get -d -v
RUN go test -v
RUN CGO_ENABLED=0 GOOS=linux go install -a -v

FROM scratch
COPY --from=0 /go/bin/telescope-control-system /
CMD ["/telescope-control-system"]
