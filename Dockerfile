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

FROM python:3.8.7-slim


RUN apt update
RUN apt install -y sshpass

RUN pip install ansible \
    && pip install netaddr \
    && pip install pywinrm \
    && pip install grpcio-tools \
    && pip install grpcio

COPY --from=stage-build /build/kobe/dist/etc /etc/
COPY --from=stage-build /build/kobe/dist/usr /usr/

COPY  --from=stage-build /build/kobe/plugin /tmp/plugin

WORKDIR /tmp/plugin
RUN python setup.py bdist_wheel
RUN pip install dist/*
RUN rm -fr /tmp/plugin

WORKDIR /root
RUN echo 'kobe-server' >> /root/entrypoint.sh

VOLUME ["/var/kobe/data"]

EXPOSE 8080

CMD ["sh","/root/entrypoint.sh"]
