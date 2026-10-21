# The UI build is platform-independent, so it runs on the build host even
# when the image targets another architecture.
FROM --platform=$BUILDPLATFORM node:22-bookworm-slim AS ui
WORKDIR /src/ui
COPY ui/package.json ui/package-lock.json ./
RUN npm ci
COPY api/schema.graphql /src/api/schema.graphql
COPY ui/ .
RUN npm run build

# Go cross-compiles for the target, so this stage runs natively too — no
# emulation for the arm64 image.
FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui /src/ui/build/site ui/build/site
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /moneyd ./cmd/moneyd

# PDF import shells out to poppler's pdftotext, so the runtime needs it.
FROM debian:bookworm-slim
RUN apt-get update \
	&& apt-get install -y --no-install-recommends poppler-utils ca-certificates \
	&& rm -rf /var/lib/apt/lists/* \
	&& useradd --system --uid 1000 --create-home moneyd
COPY --from=build /moneyd /usr/local/bin/moneyd
# moneyd keeps its data in ./moneyd-data; with /data as the working
# directory, mounting /data persists it.
WORKDIR /data
RUN chown moneyd /data
USER moneyd
VOLUME /data
EXPOSE 8080
ENTRYPOINT ["moneyd"]
