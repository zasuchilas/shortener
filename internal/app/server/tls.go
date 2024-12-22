package server

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

// makeTLS generates cert.pem and key.pem
// Another way is use generate_cert.go in crypto/tls to generate cert.pem and key.pem
// https://pkg.go.dev/net/http#ListenAndServeTLS
func makePemFiles() error {
	// creating a certificate template
	cert := &x509.Certificate{
		// specify the unique certificate number
		SerialNumber: big.NewInt(1658),
		// filling in the basic information about the certificate holder
		Subject: pkix.Name{
			Organization: []string{"YP_GO"},
			Country:      []string{"RU"},
		},
		// allow the use of the certificate for 127.0.0.1 and ::1
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		// the certificate is valid starting from the time of creation
		NotBefore: time.Now(),
		// the certificate's lifetime is 10 years
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// setting the use of the key for digital signature,
		// as well as client and server authorization
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// creating a new private RSA key with a length of 4096 bits
	// note that rand.Reader is used as a random data source
	// to generate the key and certificate
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	// creating an x.509 certificate
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// encoding the certificate and key in the PEM format,
	// which is used for storing and exchanging cryptographic keys
	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return err
	}

	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return err
	}

	// creating cert and key files
	if err = os.WriteFile("key.pem", privateKeyPEM.Bytes(), 0600); err != nil {
		return err
	}
	if err = os.WriteFile("cert.pem", certPEM.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}
