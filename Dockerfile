# syntax=docker/dockerfile:1

# Build with:
#   docker build --ssh default .
# Requires an ssh-agent holding a key with read access to
# github.com/ccatobs/antenna-control-unit (try `ssh-add -l`).

FROM golang:1.25-bookworm AS build
RUN apt-get update \
 && apt-get install -y --no-install-recommends autoconf automake libtool xz-utils \
 && rm -rf /var/lib/apt/lists/*
WORKDIR /src

# C deps: rebuilt only when build-deps changes
COPY build-deps .
RUN ./build-deps

# Go deps: re-downloaded only when go.mod/go.sum change
ENV GOPRIVATE=github.com/ccatobs GOFLAGS=-mod=readonly
COPY .github/known_hosts /root/.ssh/known_hosts
RUN git config --global url."ssh://git@github.com/ccatobs/".insteadOf https://github.com/ccatobs/
COPY go.mod go.sum ./
RUN --mount=type=ssh go mod download

COPY *.go ./
RUN go vet ./... && go test -v ./...
RUN go build -trimpath -tags netgo,osusergo \
      -ldflags='-s -w -extldflags=-static' -o /telescope-control-system

FROM scratch
COPY --from=build /telescope-control-system /
USER 65534:65534
ENTRYPOINT ["/telescope-control-system"]
