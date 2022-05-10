FROM golang:1.14 as stage-build
LABEL stage=stage-build
WORKDIR /build/kobe
ARG GOARCH

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=$GOARCH
ENV CGO_ENABLED=1

RUN apt-get update
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build_server_linux GOARCH=$GOARCH

FROM kubeoperator/kobe:euler2sp10-20220111

RUN mkdir /root/.ssh  \
    && touch /root/.ssh/config \
    && echo -e "Host *\n\tStrictHostKeyChecking no\n\tUserKnownHostsFile /dev/null" > /root/.ssh/config

COPY --from=stage-build /build/kobe/dist/etc /etc/
COPY --from=stage-build /build/kobe/dist/usr /usr/
COPY --from=stage-build /build/kobe/dist/var /var/

RUN echo 'kobe-server' >> /root/entrypoint.sh

RUN chmod 550 /root/entrypoint.sh /usr/local/bin/kobe-server /usr/local/bin/kobe-inventory

VOLUME ["/var/kobe/data"]

RUN mkdir /var/kobe/conf -p
COPY conf/server.k /var/kobe/conf
COPY conf/server.p /var/kobe/conf

EXPOSE 8080

CMD ["sh","/root/entrypoint.sh"]
