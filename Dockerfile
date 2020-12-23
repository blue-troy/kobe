FROM golang:1.14-alpine as stage-build
LABEL stage=stage-build
WORKDIR /build/kobe
ARG GOARCH

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=$GOARCH
ENV CGO_ENABLED=0


RUN  apk update \
  && apk add git \
  && apk add make
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build_server_linux GOARCH=$GOARCH

FROM alpinelinux/ansible:latest

RUN apk add sshpass \
    && apk add git \
    && apk add python3-dev \
    && apk add g++ \
    && pip3 install netaddr \
    && pip3 install pywinrm \
    && pip3 install grpcio-tools \
    && pip3 install grpcio


WORKDIR /tmp
RUN git clone https://github.com/KubeOperator/KobeAnsiblePlugin.git

WORKDIR /tmp/KobeAnsiblePlugin
RUN mkdir -r /var/kobe/lib/ansible
RUN cp -rf plugins /var/kobe/lib/ansible
RUN python setup.py install


WORKDIR /root
RUN mkdir /root/.ssh  \
    && touch /root/.ssh/config \
    && echo -e "Host *\n\tStrictHostKeyChecking no\n\tUserKnownHostsFile /dev/null" > /root/.ssh/config

COPY --from=stage-build /build/kobe/dist/etc /etc/
COPY --from=stage-build /build/kobe/dist/usr /usr/
COPY --from=stage-build /build/kobe/dist/var /var/

RUN echo 'kobe-server' >> /root/entrypoint.sh

VOLUME ["/var/kobe/data"]

EXPOSE 8080

CMD ["sh","/root/entrypoint.sh"]
