FROM golang:1.20-bullseye

ENV VERSION_ID Debian_11
RUN apt-get update \
    && apt-get -y install curl gnupg2 git \
    && echo "deb https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/${VERSION_ID}/ /" | tee /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list \
    && curl -L "https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/${VERSION_ID}/Release.key" | apt-key add - \
    && apt-get update && apt-get -y upgrade && apt-get -y install podman dumb-init \
    && rm -rf /var/lib/apt/lists/*

RUN useradd --create-home --shell /bin/bash rootless
RUN mkdir -p /home/rootless/src
WORKDIR /home/rootless/src
USER rootless

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["bash"]
