FROM golang as builder

WORKDIR /build

ADD go.mod .
ADD go.sum .

ADD . .

RUN go install github.com/bufbuild/buf/cmd/buf@latest
RUN go generate ./...
RUN go build ./cmd/moneyd

FROM debian:stable-slim

COPY --from=builder /build/moneyd /

ENTRYPOINT [ "./moneyd"]
