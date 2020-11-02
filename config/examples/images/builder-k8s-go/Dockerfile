FROM fakod/code-server-base

FROM ubuntu:18.04

RUN apt-get update && apt-get install -y \
	openssl \
	net-tools \
	git \
	locales \
	sudo \
	dumb-init \
	vim \
	curl \
	wget

RUN locale-gen en_US.UTF-8
# We cannot use update-locale because docker will not use the env variables
# configured in /etc/default/locale so we need to set it manually.
ENV LC_ALL=en_US.UTF-8 \
	SHELL=/bin/bash

RUN adduser --gecos '' --disabled-password cnde && \
	echo "cnde ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/nopasswd

USER cnde
# We create first instead of just using WORKDIR as when WORKDIR creates, the
# user is root.
RUN mkdir -p /home/cnde/project

WORKDIR /home/cnde/project

# This ensures we have a volume mounted even if the user forgot to do bind
# mount. So that they do not lose their data if they delete the container.
#VOLUME [ "/home/coder/project" ]

COPY --from=0 /code-server /usr/local/bin/code-server

#############################################################################################################

ENV HOMEPATH /home/cnde

USER root

RUN apt-get update; \
    apt-get install -y apt-transport-https ca-certificates gnupg-agent software-properties-common; \
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -; \
    add-apt-repository \
    "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
    $(lsb_release -cs) \
    stable"; \
    apt-get update && apt-get install -y docker-ce-cli; \
    rm -rf /var/lib/apt/lists/*

RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -; \
    echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list; \
    apt-get update; apt-get install -y kubectl; \
    rm -rf /var/lib/apt/lists/*

RUN curl -L https://github.com/wercker/stern/releases/download/1.10.0/stern_linux_amd64 -o /usr/local/bin/stern; \
    chmod +x /usr/local/bin/stern

RUN apt-get update; apt-get install -y zsh; rm -rf /var/lib/apt/lists/*;

RUN curl https://raw.githubusercontent.com/blendle/kns/master/bin/kns -o /usr/local/bin/kns; \ 
    chmod +x /usr/local/bin/kns

RUN apt-get update && apt-get install -y --no-install-recommends \
		g++ \
		gcc \
		libc6-dev \
		make \
		pkg-config \
	&& rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.12.6

RUN set -eux; \
	\
# this "case" statement is generated via "update.sh"
	dpkgArch="$(dpkg --print-architecture)"; \
	case "${dpkgArch##*-}" in \
		amd64) goRelArch='linux-amd64'; goRelSha256='dbcf71a3c1ea53b8d54ef1b48c85a39a6c9a935d01fc8291ff2b92028e59913c' ;; \
		armhf) goRelArch='linux-armv6l'; goRelSha256='0708fbc125e7b782b44d450f3a3864859419b3691121ad401f1b9f00e488bddb' ;; \
		arm64) goRelArch='linux-arm64'; goRelSha256='8f4e3909c74b4f3f3956715f32419b28d32a4ad57dbd79f74b7a8a920b21a1a3' ;; \
		i386) goRelArch='linux-386'; goRelSha256='7aaf25164a9ab5e1005c15535ed16ee122df50ac192c2d79b7940315c2b74f2c' ;; \
		ppc64el) goRelArch='linux-ppc64le'; goRelSha256='67eacb68c1e251c1428e588776c5a02e287a508e3d44f940d31d8ff5d57f0eef' ;; \
		s390x) goRelArch='linux-s390x'; goRelSha256='c14baa10b87a38e56f28a176fae8a839e9052b0e691bdc0461677d4bcedea9aa' ;; \
		*) goRelArch='src'; goRelSha256='c96c5ccc7455638ae1a8b7498a030fe653731c8391c5f8e79590bce72f92b4ca'; \
			echo >&2; echo >&2 "warning: current architecture ($dpkgArch) does not have a corresponding Go binary release; will be building from source"; echo >&2 ;; \
	esac; \
	\
	url="https://golang.org/dl/go${GOLANG_VERSION}.${goRelArch}.tar.gz"; \
	wget -O go.tgz "$url"; \
	echo "${goRelSha256} *go.tgz" | sha256sum -c -; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	\
	if [ "$goRelArch" = 'src' ]; then \
		echo >&2; \
		echo >&2 'error: UNIMPLEMENTED'; \
		echo >&2 'TODO install golang-any from jessie-backports for GOROOT_BOOTSTRAP (and uninstall after build)'; \
		echo >&2; \
		exit 1; \
	fi; \
	\
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

ENV GOPATH /$HOMEPATH/project/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

ENV DOCKER_HOST tcp://0.0.0.0:2375

USER cnde
WORKDIR /home/cnde

RUN sudo chown -R cnde .

RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh)"
COPY --chown=cnde:cnde .zshrc .
RUN sudo chown -R cnde .
RUN git clone --depth 1 https://github.com/junegunn/fzf.git ~/.fzf ; ~/.fzf/install

WORKDIR /home/cnde/project
ENTRYPOINT ["dumb-init", "code-server"]