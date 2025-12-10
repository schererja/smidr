# syntax=docker/dockerfile:1
FROM ubuntu:22.04

# Multi-arch support: use TARGETARCH to conditionally install gcc-multilib on amd64 only
ARG TARGETARCH

RUN apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
  gawk wget git-core diffstat unzip texinfo \
  build-essential chrpath socat cpio python3 python3-pip python3-pexpect \
  xz-utils debianutils iputils-ping python3-git python3-jinja2 libegl1-mesa \
  libsdl1.2-dev pylint xterm locales sudo \
  file lz4 zstd && \
  if [ "$TARGETARCH" = "amd64" ]; then apt-get install -y gcc-multilib; fi && \
  rm -rf /var/lib/apt/lists/*

# Set up a UTF-8 locale
RUN locale-gen en_US.UTF-8
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8


# Create a non-root user for builds (optional, recommended for reproducibility)
RUN useradd -ms /bin/bash builder && \
  mkdir -p /home/builder/downloads /home/builder/sstate-cache /home/builder/work /home/builder/layers && \
  chown -R builder:builder /home/builder

# Set environment variables for Yocto integration
ENV YOCTO_DOWNLOADS=/home/builder/downloads \
  YOCTO_SSTATE=/home/builder/sstate-cache \
  YOCTO_WORK=/home/builder/work \
  YOCTO_LAYERS=/home/builder/layers

# Install debug helper and set up working directory
COPY --chown=builder:builder apps/tools/smidr-debug /usr/local/bin/smidr-debug
RUN chmod +x /usr/local/bin/smidr-debug || true

USER builder
WORKDIR /home/builder/work

# Entrypoint for interactive use
ENTRYPOINT ["/bin/bash"]

# Integration notes for smidr builder:
#   - Mount host directories to these paths for persistence:
#       -v $HOST_DOWNLOADS:/home/builder/downloads \
#       -v $HOST_SSTATE:/home/builder/sstate-cache \
#       -v $HOST_WORK:/home/builder/work
#   - Mount Yocto meta-layers for injection:
#       -v $HOST_LAYER1:/home/builder/layers/layer-0 \
#       -v $HOST_LAYER2:/home/builder/layers/layer-1 \
#       (add more as needed)
#   - The default working directory is /home/builder/work
#   - Environment variables YOCTO_DOWNLOADS, YOCTO_SSTATE, YOCTO_WORK, YOCTO_LAYERS are set for use in scripts/builds

# Build for a specific architecture:
#   docker buildx build --platform linux/amd64 -t smidr-yocto:amd64 .
#   docker buildx build --platform linux/arm64 -t smidr-yocto:arm64 .
#
# gcc-multilib is only installed on amd64 (x86_64) builds.
