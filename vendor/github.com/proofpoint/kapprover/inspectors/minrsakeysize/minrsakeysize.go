package minrsakeysize

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/proofpoint/kapprover/csr"
	"github.com/proofpoint/kapprover/inspectors"
	certificates "k8s.io/api/certificates/v1beta1"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

func init() {
	inspectors.Register("minrsakeysize", &minrsakeysize{3072})
}

// Minkeysize is an Inspector that verifies that the CSR either has a non-RSA public key or has an
// RSA public key of at least a configured minimum size. If you want to restrict public keys, use
// the signaturealgorithm Inspector.
type minrsakeysize struct {
	minSize int
}

func (m *minrsakeysize) Configure(config string) (inspectors.Inspector, error) {
	if config != "" {
		minsize, err := strconv.ParseUint(config, 10, 0)
		if err != nil {
			return nil, err
		}
		return &minrsakeysize{minSize: int(minsize)}, nil
	}
	return m, nil
}

func (m *minrsakeysize) Inspect(client kubernetes.Interface, request *certificates.CertificateSigningRequest) (string, error) {
	certificateRequest, msg := csr.Extract(request.Spec.Request)
	if msg != "" {
		return msg, nil
	}

	if certificateRequest.PublicKeyAlgorithm != x509.RSA {
		return "", nil
	}

	bitsize := certificateRequest.PublicKey.(*rsa.PublicKey).N.BitLen()
	if bitsize < m.minSize {
		return fmt.Sprintf("Public key too small: %d < %d", bitsize, m.minSize), nil
	}

	return "", nil
}
