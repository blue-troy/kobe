package main

import (
	"log"
	"net"

	"github.com/KubeOperator/kobe/api"
	"github.com/KubeOperator/kobe/pkg/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func newTcpListener(address string) (*net.Listener, error) {
	s, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
func newServer() *grpc.Server {
	c, err := credentials.NewServerTLSFromFile("/var/kobe/conf/server.pem", "/var/kobe/conf/server.key")
	if err != nil {
		log.Fatalf("credentials.NewServerTLSFromFile err: %v", err)
	}

	options := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(100 * 1024 * 1024 * 1024),
		grpc.MaxSendMsgSize(100 * 1024 * 1024 * 1024),
		grpc.Creds(c),
	}
	gs := grpc.NewServer(options...)
	kobe := server.NewKobe()
	api.RegisterKobeApiServer(gs, kobe)
	return gs
}
