# First stage - Debian with document processing dependencies and curl for downloading.
# Kept on debian:bookworm-slim (not ubuntu:noble) so the builder's glibc matches the
# distroless final stage (both Debian 12 / glibc 2.36); Noble's glibc 2.39 libraries
# fail to load on the base with "version `GLIBC_2.38' not found". See mattermost/mattermost#36182.
FROM debian:bookworm-slim@sha256:7b140f374b289a7c2befc338f42ebe6441b7ea838a042bbd5acbfca6ec875818 AS builder
# Setting bash as our shell, and enabling pipefail option
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Build Arguments
ARG PUID=2000
ARG PGID=2000
# MM_PACKAGE build arguments controls which version of mattermost to install, defaults to latest stable enterprise
# i.e. https://releases.mattermost.com/9.7.1/mattermost-9.7.1-linux-amd64.tar.gz
ARG MM_PACKAGE="https://latest.mattermost.com/mattermost-enterprise-linux"

# Note: In the original implementation, it downloads the uploaded artifacts. In the customized version, it uses local artifacts instead.
ARG MM_CUSTOM_PACKAGE="mattermost-team-linux-amd64.tar.gz"
COPY ${MM_CUSTOM_PACKAGE} /tmp/app_package.tar.gz

# Install needed packages and indirect dependencies
RUN apt-get update \
  && DEBIAN_FRONTEND=noninteractive apt-get install --no-install-recommends -y \
  ca-certificates \
  curl \
  media-types \
  mailcap \
  unrtf \
  wv \
  poppler-utils \
  tidy \
  tzdata \
  && rm -rf /var/lib/apt/lists/*

# Resolve and stage the full runtime library closure of the document-extraction
# utilities. These libraries live under the Debian multiarch path
# /usr/lib/x86_64-linux-gnu/ (not /usr/lib/), and each binary also pulls in
# transitive dependencies (libpoppler -> liblcms2, libopenjp2, libtiff, libjpeg,
# libnss3, ...) that a hand-maintained COPY list misses. Resolve them with ldd so
# nothing is left out. glibc and the dynamic loader are deliberately excluded so the
# distroless base's own copies (identical Debian 12 versions) are left untouched for
# the mattermost/mmctl Go binaries. ldd failures are ignored because some tools
# (e.g. wvText) are shell-script wrappers, not ELF binaries; their real dependencies
# come in via the binary they exec (wvWare). See mattermost/mattermost#36182.
RUN set -eu \
  && mkdir -p /deps \
  && for bin in /usr/bin/pdftotext /usr/bin/wvText /usr/bin/wvWare /usr/bin/unrtf /usr/bin/tidy; do \
       ldd "$bin" 2>/dev/null || true; \
     done \
     | awk '{ for (i = 1; i <= NF; i++) if ($i ~ /^\//) print $i }' \
     | sort -u \
     | grep -Ev '/(ld-linux[^/]*|libc|libm|libmvec|libdl|libpthread|librt|libresolv|libanl|libutil|libnsl|libBrokenLocale)\.so' \
     | while read -r lib; do \
         mkdir -p "/deps$(dirname "$lib")"; \
         cp -aL "$lib" "/deps$lib"; \
       done

# Set mattermost group/user and download Mattermost
RUN mkdir -p /mattermost/data /mattermost/plugins /mattermost/client/plugins \
  && groupadd --gid ${PGID} mattermost \
  && useradd --uid ${PUID} --gid ${PGID} --comment "" --home-dir /mattermost mattermost \
  # && curl -L $MM_PACKAGE | tar -xvz \
  && tar -xzf /tmp/app_package.tar.gz -C /mattermost --strip-components=1 \
  && rm /tmp/app_package.tar.gz \
  && chown -R mattermost:mattermost /mattermost /mattermost/data /mattermost/plugins /mattermost/client/plugins

# Create PostgreSQL client SSL directory structure for ssl_mode=require
RUN mkdir -p /mattermost/.postgresql \
  && chmod 700 /mattermost/.postgresql

# Final stage using distroless for minimal attack surface
FROM gcr.io/distroless/base-debian12

# Some ENV variables
ENV PATH="/mattermost/bin:${PATH}"
ENV MM_SERVICESETTINGS_ENABLELOCALMODE="true"

# Copy over metadata files needed by runtime
COPY --from=builder /etc/mime.types /etc

# Copy CA certificates for SSL/TLS validation with proper ownership
COPY --from=builder --chown=2000:2000 /etc/ssl/certs /etc/ssl/certs

# Copy document processing utilities and necessary support files
COPY --from=builder /usr/bin/pdftotext /usr/bin/pdftotext
COPY --from=builder /usr/bin/wvText /usr/bin/wvText
COPY --from=builder /usr/bin/wvWare /usr/bin/wvWare
COPY --from=builder /usr/bin/unrtf /usr/bin/unrtf
COPY --from=builder /usr/bin/tidy /usr/bin/tidy
COPY --from=builder /usr/share/wv /usr/share/wv

# Copy the resolved runtime library closure for the document-processing
# utilities (staged via ldd in the builder stage). This replaces the previous
# hand-listed COPY lines, which used the non-multiarch path /usr/lib/ and so
# silently copied nothing on Ubuntu Noble. See mattermost/mattermost#36182.
COPY --from=builder /deps/ /

# Copy mattermost from builder stage
COPY --from=builder --chown=2000:2000 /mattermost /mattermost

# Copy passwd including mattermost user
COPY  passwd /etc/passwd

# We should refrain from running as privileged user
USER mattermost

# Healthcheck to make sure container is ready - using mmctl instead of curl for distroless compatibility
HEALTHCHECK --interval=30s --timeout=10s \
  CMD ["/mattermost/bin/mmctl", "system", "status", "--local"]

# Configure entrypoint and command with proper permissions
WORKDIR /mattermost
CMD ["/mattermost/bin/mattermost"]

EXPOSE 8065 8067 8074 8075

# Declare volumes for mount point directories
VOLUME ["/mattermost/data", "/mattermost/logs", "/mattermost/config", "/mattermost/plugins", "/mattermost/client/plugins"]
