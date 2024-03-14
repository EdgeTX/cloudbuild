FROM golang:1.20-bullseye as builder

RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN make build

FROM node:slim as ui-builder

RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN npm ci --prefix ui/
RUN npm run build --prefix ui/

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get -y install curl gnupg2 git podman dumb-init \
    && rm -rf /var/lib/apt/lists/*

RUN useradd --create-home --shell /bin/bash rootless
RUN mkdir -p /home/rootless/src
WORKDIR /home/rootless/src

USER rootless
COPY --from=builder /build/bin/ebuild ./
COPY --from=builder /build/targets.json ./
COPY --from=ui-builder /build/static ./static

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["bash"]
