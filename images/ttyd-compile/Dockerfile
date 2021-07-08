FROM tsl0922/musl-cross
LABEL maintainer "Shuanglei Tao - tsl0922@gmail.com"

RUN git clone --single-branch --branch 1.6.3 --depth=1 https://github.com/tsl0922/ttyd.git /ttyd \
    && cd /ttyd \
    && ./scripts/cross-build.sh x86_64