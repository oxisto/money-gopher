FROM golang as builder

WORKDIR /build

RUN apt update && apt install -y nodejs npm

ADD go.mod .
ADD go.sum .

RUN go install github.com/bufbuild/buf/cmd/buf@latest

ADD . .

RUN go generate ./...
RUN go build ./cmd/moneyd

FROM debian:stable-slim

COPY --from=builder /build/moneyd /

EXPOSE 8080
ENTRYPOINT [ "./moneyd"]
