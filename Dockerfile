# ----------------------------------------------------
# STAGE 1: Build Go Manager (NATIVE)
# ----------------------------------------------------
FROM golang:1.25 AS go-builder

ARG TARGETOS
ARG TARGETARCH
ARG GIT_HASH="unknown"
ARG BUILD_TIME=""

WORKDIR /app
COPY nvr_core/ nvr_core/

# Initialize go.mod if needed
RUN cd nvr_core && ([ -f go.mod ] || go mod init nvr-core)

# Download and build
RUN cd nvr_core && go mod download && go mod tidy && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w -X 'nvr_core/buildinfo.Version=${GIT_HASH}' -X 'nvr_core/buildinfo.BuildTime=${BUILD_TIME}'" \
    -o /app/nvr_service


# ----------------------------------------------------
# STAGE 2: Build C++ Worker (EMULATED)
# -------------------------------------------------------
FROM ubuntu:22.04 AS cpp-builder

ARG GIT_HASH="unknown"

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    pkg-config \
    libavformat-dev \
    libavcodec-dev \
    libavutil-dev \
    libswscale-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY cpp_engine/ cpp_engine/

RUN mkdir -p cpp_engine/build && \
    cd cpp_engine/build && \
    cmake -DTHE_VERSION=0.1.0 -DGIT_HASH=${GIT_HASH} .. && \
    make && \
    cp nvr_worker /build/nvr_worker


# -------------------------------------------------------
# STAGE 3: Final Runtime Image
# -------------------------------------------------------
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    libavformat58 \
    libavcodec58 \
    libavutil56 \
    libswscale5 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=cpp-builder /build/nvr_worker ./nvr_worker
COPY --from=go-builder /app/nvr_service ./nvr_service

CMD ["./nvr_service"]
