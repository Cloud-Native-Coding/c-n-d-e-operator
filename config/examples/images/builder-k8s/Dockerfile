FROM fakod/code-server-base

FROM ubuntu:19.10

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