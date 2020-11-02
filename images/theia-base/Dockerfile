ARG NODE_VERSION=10
FROM node:${NODE_VERSION}-alpine
RUN apk add --no-cache make gcc g++ python
ARG version=latest
WORKDIR /home/theia
ADD $version.package.json ./package.json
ARG GITHUB_TOKEN
RUN yarn --pure-lockfile && \
    NODE_OPTIONS="--max_old_space_size=4096" yarn theia build && \
    yarn theia download:plugins && \
    yarn --production && \
    yarn autoclean --init && \
    echo *.ts >> .yarnclean && \
    echo *.ts.map >> .yarnclean && \
    echo *.spec.* >> .yarnclean && \
    yarn autoclean --force && \
    yarn cache clean

FROM ubuntu:20.04

RUN apt-get update && apt-get install curl sudo musl-dev -y && curl -sL https://deb.nodesource.com/setup_10.x | bash && \
    apt-get install tini git nodejs -y && rm -rf /var/lib/apt/lists/* && ln -s /usr/lib/x86_64-linux-musl/libc.so /lib/libc.musl-x86_64.so.1

RUN adduser --gecos '' --disabled-password cnde && \
  echo "cnde ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/nopasswd

RUN chmod g+rw /home && \
    mkdir -p /home/cnde/project && \
    chown -R cnde:cnde /home/cnde && \
    chown -R cnde:cnde /home/cnde/project;

ENV HOME /home/cnde
WORKDIR /home/cnde
COPY --from=0 --chown=cnde:cnde /home/theia /home/cnde
COPY ./ttyd /usr/local/bin/ttyd
EXPOSE 3000
ENV SHELL=/bin/bash \
    THEIA_DEFAULT_PLUGINS=local-dir:/home/cnde/plugins
ENV USE_LOCAL_GIT true
USER cnde
ENTRYPOINT [ "tini", "-s", "--", "node", "/home/cnde/src-gen/backend/main.js", "/home/cnde/project", "--hostname=0.0.0.0", "--port=8080" ]
