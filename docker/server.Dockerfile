FROM golang:1.18 AS builder

RUN apt-get -qq update && apt-get -yqq install upx

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /powwy

COPY . .

RUN go build \
  -trimpath \
  -ldflags "-s -w -extldflags '-static'" \
  -installsuffix cgo \
  -o /bin/powwy \
  ./cmd/powwy-srv

RUN strip /bin/powwy
RUN upx -q -9 /bin/powwy


FROM scratch
COPY --from=builder /bin/powwy /bin/powwy
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/bin/powwy"]