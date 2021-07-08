FROM ubuntu:20.10

RUN apt-get update \
 && DEBIAN_FRONTEND="noninteractive" TZ=Europe/Zurich apt-get install -y \
    curl \
    dumb-init \
    htop \
    locales \
    man \
    nano \
    git \
    procps \
    ssh \
    sudo \
    vim \
    tini \
  && rm -rf /var/lib/apt/lists/*

ARG VERSION=3.10.2

RUN curl -SsL https://github.com/cdr/code-server/releases/download/v$VERSION/code-server-$VERSION-linux-amd64.tar.gz \
    | tar -C /tmp -xzf - && \
    mv /tmp/code-server* /usr/local/lib/code-server && \
    ln -s /usr/local/lib/code-server/bin/code-server /usr/local/bin/code-server

RUN sed -i "s/# en_US.UTF-8/en_US.UTF-8/" /etc/locale.gen && locale-gen
ENV LANG=en_US.UTF-8

RUN adduser --gecos '' --disabled-password cnde && \
  echo "cnde ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/nopasswd

COPY ./ttyd /usr/local/bin/ttyd

EXPOSE 8080
USER cnde
WORKDIR /home/cnde/project
ENTRYPOINT ["tini", "-s", "--", "code-server", "--auth", "none", "--bind-addr", "0.0.0.0:8080", "--disable-telemetry"]
