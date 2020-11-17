# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.137.0/containers/go/.devcontainer/base.Dockerfile
ARG VARIANT="1"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

# [Optional] Install a version of Node.js using nvm for front end dev
ARG INSTALL_NODE="true"
ARG NODE_VERSION="lts/*"
RUN if [ "${INSTALL_NODE}" = "true" ]; then su vscode -c "source /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

# Install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && apt-get -y install --no-install-recommends ...

# Install goreleaser
RUN wget https://github.com/goreleaser/goreleaser/releases/download/v0.147.2/goreleaser_amd64.deb \
        -O /tmp/goreleaser_amd64.deb \
    && dpkg -i /tmp/goreleaser_amd64.deb \
    && rm -f /tmp/goreleaser_amd64.deb
