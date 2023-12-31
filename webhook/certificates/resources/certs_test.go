/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resources

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	. "github.com/Yangfisher1/knative-common-pkg/logging/testing"
)

func TestCreateCerts(t *testing.T) {
	sKey, serverCertPEM, caCertBytes, err := CreateCerts(TestContextWithLogger(t), "got-the-hook", "knative-webhook", time.Now().AddDate(0, 0, 7))
	if err != nil {
		t.Fatal("Failed to create certs", err)
	}

	// Test server private key
	p, _ := pem.Decode(sKey)
	if p.Type != "PRIVATE KEY" {
		t.Fatal("Expected the key to be Private key type")
	}
	key, err := x509.ParsePKCS8PrivateKey(p.Bytes)
	if err != nil {
		t.Fatal("Failed to parse private key", err)
	}
	if _, ok := key.(*ecdsa.PrivateKey); !ok {
		t.Fatalf("Key is not ecdsa format, actually %t", key)
	}

	// Test Server Cert
	sCert, err := validCertificate(serverCertPEM, t)
	if err != nil {
		t.Fatal(err)
	}

	// Test CA Cert
	caParsedCert, err := validCertificate(caCertBytes, t)
	if err != nil {
		t.Fatal(err)
	}

	// Verify common name
	const expectedCommonName = "got-the-hook.knative-webhook.svc"

	if caParsedCert.Subject.CommonName != expectedCommonName {
		t.Fatalf("Unexpected Cert Common Name %q, wanted %q", caParsedCert.Subject.CommonName, expectedCommonName)
	}

	// Verify domain names
	expectedDNSNames := []string{
		"got-the-hook",
		"got-the-hook.knative-webhook",
		"got-the-hook.knative-webhook.svc",
		"got-the-hook.knative-webhook.svc.cluster.local",
	}
	if diff := cmp.Diff(caParsedCert.DNSNames, expectedDNSNames); diff != "" {
		t.Fatal("Unexpected CA Cert DNS Name (-want +got) :", diff)
	}

	if diff := cmp.Diff(caParsedCert.DNSNames, expectedDNSNames); diff != "" {
		t.Fatal("Unexpected CA Cert DNS Name (-want +got):", diff)
	}

	// Verify Server Cert is Signed by CA Cert
	if err = sCert.CheckSignatureFrom(caParsedCert); err != nil {
		t.Fatal("Failed to verify that the signature on server certificate is from parent CA cert", err)
	}
}

func validCertificate(cert []byte, t *testing.T) (*x509.Certificate, error) {
	t.Helper()
	const certificate = "CERTIFICATE"
	caCert, _ := pem.Decode(cert)
	if caCert.Type != certificate {
		return nil, fmt.Errorf("cert.Type = %s, want: %s", caCert.Type, certificate)
	}
	parsedCert, err := x509.ParseCertificate(caCert.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse cert %w", err)
	}
	if parsedCert.SignatureAlgorithm != x509.ECDSAWithSHA256 {
		return nil, fmt.Errorf("Failed to match signature. Got: %s, want: %s", parsedCert.SignatureAlgorithm, x509.SHA256WithRSA)
	}
	return parsedCert, nil
}
