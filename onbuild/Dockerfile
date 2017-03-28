# vim: ft=dockerfile
FROM alpine:3.5
MAINTAINER Jimmy Zelinskie <jimmy.zelinskie@coreos.com>

ENV GOLANG_VERSION 1.8
ENV GOLANG_SRC_URL https://golang.org/dl/go$GOLANG_VERSION.src.tar.gz
ENV GOLANG_SRC_SHA256 406865f587b44be7092f206d73fc1de252600b79b3cacc587b74b5ef5c623596
ENV GOLANG_NO_PIC_PATCH_URL https://raw.githubusercontent.com/docker-library/golang/master/1.8/alpine/no-pic.patch
ENV GOPATH /go
ENV GOBIN /usr/local/bin
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

COPY go-wrapper /usr/local/bin/

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

CMD ["go-wrapper", "run"]

ONBUILD COPY . /go/src/app
ONBUILD RUN set -ex \
	&& apk add --no-cache --virtual .build-deps \
		bash \
		curl \
		gcc \
		go \
		musl-dev \
		git \
		openssl \
	&& apk add --no-cache --virtual .runtime-deps \
		ca-certificates \
	&& export GOROOT_BOOTSTRAP="$(go env GOROOT)" \
	&& wget -q "$GOLANG_SRC_URL" -O golang.tar.gz \
	&& echo "$GOLANG_SRC_SHA256  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz \
	&& cd /usr/local/go/src \
	&& curl ${GOLANG_NO_PIC_PATCH_URL} | patch -p2 \
	&& ./make.bash \
	&& cd /go/src/app \
	&& go-wrapper download \
	&& go-wrapper install \
	&& update-ca-certificates \
	&& apk del .build-deps \
	&& cd / \
	&& rm -rf /tmp/* \
	&& rm -rf /go \
	&& rm -rf /usr/local/go*
