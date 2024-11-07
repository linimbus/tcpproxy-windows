package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"strings"
	"time"
)

func tlsCert(addr string) (tls.Certificate, error) {
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, max)
	subject := pkix.Name{
		Organization:       []string{"TcpProxy App co."},
		OrganizationalUnit: []string{"TcpProxy App"},
		CommonName:         "TcpProxy App",
	}

	if addr == "0.0.0.0" || addr == "::" {
		addr = "127.0.0.1"
	}

	ipAddress := make([]net.IP, 0)
	ipAddress = append(ipAddress, net.ParseIP(addr))

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  ipAddress,
	}
	pk, _ := rsa.GenerateKey(rand.Reader, 2048)

	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &pk.PublicKey, pk)

	certOut := bytes.NewBuffer(make([]byte, 0))
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut := bytes.NewBuffer(make([]byte, 0))
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)})

	return tls.X509KeyPair(certOut.Bytes(), keyOut.Bytes())
}

func TlsConfigClient(server string, client string, version string) (*tls.Config, error) {
	var certs tls.Certificate
	var err error

	certs, err = tlsCert(client)
	if err != nil {
		return nil, err
	}

	tls_version := tls.VersionTLS13

	if strings.Compare(version, "TLS1.2") == 0 {
		tls_version = tls.VersionTLS12
	}

	if strings.Compare(version, "TLS1.3") == 0 {
		tls_version = tls.VersionTLS13
	}

	return &tls.Config{
		MinVersion:         uint16(tls_version),
		MaxVersion:         tls.VersionTLS13,
		ServerName:         server,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{certs},
	}, nil
}

func TlsConfigServer(server string, version string) (*tls.Config, error) {
	var certs tls.Certificate
	var err error

	certs, err = tlsCert(server)
	if err != nil {
		return nil, err
	}

	tls_version := tls.VersionTLS13

	if strings.Compare(version, "TLS1.2") == 0 {
		tls_version = tls.VersionTLS12
	}

	if strings.Compare(version, "TLS1.3") == 0 {
		tls_version = tls.VersionTLS13
	}

	return &tls.Config{
		MinVersion:   uint16(tls_version),
		MaxVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{certs},
		ClientAuth:   tls.RequestClientCert,
	}, nil
}
