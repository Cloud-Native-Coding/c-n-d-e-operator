FROM bash:4

ARG K8S_VERSION=v1.14.3
RUN set -x                  && \
    apk --update upgrade    && \
    apk add ca-certificates openssl && \
    rm -rf /var/cache/apk/* && \
    wget -O /kubectl https://storage.googleapis.com/kubernetes-release/release/$K8S_VERSION/bin/linux/amd64/kubectl && \
    chmod +x /kubectl

COPY *.sh /

CMD /create_kubeconfig.sh