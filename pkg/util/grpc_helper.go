package util

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc/credentials"
)

func NewServerTLSFromFile(certFile, keyFile string) (credentials.TransportCredentials, error) {
	cert, err := loadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}}), nil
}

func loadX509KeyPair(certFile, keyFile string) (tls.Certificate, error) {
	certPEMBlock, err := ioutil.ReadFile(certFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	cc1, err := base64.StdEncoding.DecodeString(string(certPEMBlock))
	if err != nil {
		return tls.Certificate{}, err
	}

	keyPEMBlock, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	cc2, err := base64.StdEncoding.DecodeString(string(keyPEMBlock))
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.X509KeyPair(cc1, cc2)
}

func NewClientTLSFromFile(certFile, serverNameOverride string) (credentials.TransportCredentials, error) {
	b, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	cc, err := base64.StdEncoding.DecodeString(string(b))
	if err != nil {
		return nil, err
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(cc) {
		return nil, fmt.Errorf("credentials: failed to append certificates")
	}
	return credentials.NewTLS(&tls.Config{ServerName: serverNameOverride, RootCAs: cp}), nil
}
