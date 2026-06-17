

# ----------------------------------------------------
# STAGE 1: Build Go Manager (NATIVE)
# ----------------------------------------------------
# "--platform=$BUILDPLATFORM" is the magic fix.
# It tells Docker: "Use the Go image that matches the computer I am building on (ARM64)"
# This prevents QEMU crashes during 'go mod tidy'.
FROM golang:1.25 AS go-builder

# Pull in the target architecture from Docker buildx (e.g., linux and amd64)
ARG TARGETOS
ARG TARGETARCH
ARG GIT_HASH="unknown"
ARG BUILD_TIME=""

WORKDIR /app
COPY nvr_core/ nvr_core/
COPY Makefile .

# Now 'go mod tidy' runs natively on your Mac's CPU. fast and stable.
RUN [ -f go.mod ] || go mod init nvr-core
RUN go mod tidy

# Force cross-compilation to the target architecture, disable CGO, and build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build-go


# ----------------------------------------------------
# STAGE 2: Build C++ Worker (EMULATED)
# ----------------------------------------------------
# C++ must be compiled on the target architecture (AMD64).
# We have to use emulation here. It might be slow, but it won't crash like Go does.
FROM ubuntu:22.04 AS cpp-builder

ARG GIT_HASH="unknown"

# Install minimal build tools
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    build-essential \
    cmake \
    pkg-config \
    libavformat-dev \
    libavcodec-dev \
    libavutil-dev \
    libswscale-dev

WORKDIR /build
COPY cpp_engine/ cpp_engine/
COPY Makefile .

# Build the C++ binary
RUN make build-cpp



# ----------------------------------------------------
# STAGE 3: Final Runtime Image (AMD64)
# ----------------------------------------------------
FROM ubuntu:22.04

# Install runtime libs
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ffmpeg \
    libavformat58 \
    libavcodec58 \
    libavutil56 \
    libswscale5 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binaries
COPY --from=cpp-builder /build/nvr_worker ./nvr_worker
COPY --from=go-builder /app/nvr_core/nvr_service ./nvr_service


# --- Final Runtime Image (Optional but recommended for smaller size) ---
# For a bare-bones dev setup, we can stop here. 
# But let's verify it works in this environment.
CMD ["./nvr_service"]
