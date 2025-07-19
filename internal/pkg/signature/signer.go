package signature

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"

)

// CertificateInfo contiene la información del certificado digital
type CertificateInfo struct {
	CertPath    string
	KeyPath     string
	Password    string
	Certificate *x509.Certificate
	PrivateKey  *rsa.PrivateKey
}

// LoadCertificate carga el certificado y la clave privada desde los archivos
func LoadCertificate() (*CertificateInfo, error) {
	// Obtener ruta absoluta de la carpeta actual del proyecto
	baseDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("no se pudo obtener el directorio de trabajo: %v", err)
	}
	certDir := filepath.Join(baseDir, "certificados")
	pemFile := filepath.Join(certDir, "C23022479065.pem")

	// Cargar certificado y clave privada desde PEM
	cert, key, err := LoadKeyPairFromPEM(pemFile)
	if err != nil {
		return nil, err
	}

	return &CertificateInfo{
		CertPath:    pemFile,
		Password:    "", // No se requiere contraseña para PEM
		Certificate: cert,
		PrivateKey:  key,
	}, nil
}

// LoadKeyPairFromPFX devuelve un error porque el proyecto ahora utiliza PEM en lugar de PFX
// LoadKeyPairFromPFX ahora se deja como stub porque el proyecto usa PEM en lugar de PFX
func LoadKeyPairFromPFX(pfxPath, password string) (*x509.Certificate, *rsa.PrivateKey, error) {
	return nil, nil, fmt.Errorf("carga de PFX no soportada; usa certificado PEM")
}

// LoadKeyPairFromPEM carga certificado y clave privada desde un archivo PEM (CERTIFICATE + PRIVATE KEY)
func LoadKeyPairFromPEM(pemPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	pemData, err := os.ReadFile(pemPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error leyendo archivo PEM: %v", err)
	}

	var cert *x509.Certificate
	var key *rsa.PrivateKey

	for len(pemData) > 0 {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}

		switch block.Type {
		case "CERTIFICATE":
			cert, err = x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, nil, fmt.Errorf("error parseando certificado: %v", err)
			}

		// Soportar tanto claves RSA PKCS#1 ("RSA PRIVATE KEY") como PKCS#8 ("PRIVATE KEY")
		case "PRIVATE KEY", "RSA PRIVATE KEY":
			// Intentar primero PKCS#1
			var rsaKey *rsa.PrivateKey
			rsaKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				// Si falla, intentar PKCS#8
				var pkIfc interface{}
				pkIfc, err = x509.ParsePKCS8PrivateKey(block.Bytes)
				if err != nil {
					return nil, nil, fmt.Errorf("error parseando clave privada: %v", err)
				}
				var ok bool
				rsaKey, ok = pkIfc.(*rsa.PrivateKey)
				if !ok {
					return nil, nil, fmt.Errorf("clave privada no es del tipo RSA")
				}
			}
			key = rsaKey
		}
	}

	if cert == nil || key == nil {
		return nil, nil, fmt.Errorf("no se encontró certificado o clave privada en el PEM")
	}

	return cert, key, nil
}

// SignXMLAsElement firma el XML y retorna el elemento de firma como string
func SignXMLAsElement(xmlString string, certInfo *CertificateInfo) (string, error) {
	// Parsear el XML
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlString); err != nil {
		return "", fmt.Errorf("error parseando XML: %v", err)
	}

	// Crear elemento de firma
	signature := etree.NewElement("ds:Signature")
	signature.CreateAttr("xmlns:ds", "http://www.w3.org/2000/09/xmldsig#")

	// SignedInfo
	signedInfo := signature.CreateElement("ds:SignedInfo")
	signedInfo.CreateElement("ds:CanonicalizationMethod").
		CreateAttr("Algorithm", "http://www.w3.org/TR/2001/REC-xml-c14n-20010315")
	signedInfo.CreateElement("ds:SignatureMethod").
		CreateAttr("Algorithm", "http://www.w3.org/2000/09/xmldsig#rsa-sha1")

	// Reference
	reference := signedInfo.CreateElement("ds:Reference")
	reference.CreateAttr("URI", "")
	transforms := reference.CreateElement("ds:Transforms")
	transforms.CreateElement("ds:Transform").
		CreateAttr("Algorithm", "http://www.w3.org/2000/09/xmldsig#enveloped-signature")
	transforms.CreateElement("ds:Transform").
		CreateAttr("Algorithm", "http://www.w3.org/TR/2001/REC-xml-c14n-20010315")
	reference.CreateElement("ds:DigestMethod").
		CreateAttr("Algorithm", "http://www.w3.org/2000/09/xmldsig#sha1")

	// Calcular DigestValue
	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		return "", fmt.Errorf("error canonicalizando XML: %v", err)
	}
	canonXML := buf.String()

	digestValue := calculateSHA1(canonXML)
	reference.CreateElement("ds:DigestValue").SetText(base64.StdEncoding.EncodeToString(digestValue))

	// SignatureValue
	signatureValue := signature.CreateElement("ds:SignatureValue")
	buf.Reset()
	doc = etree.NewDocument()
	doc.SetRoot(signedInfo.Copy())
	if _, err := doc.WriteTo(&buf); err != nil {
		return "", fmt.Errorf("error serializando SignedInfo: %v", err)
	}
	signedInfoXML := buf.String()

	signedBytes, err := signWithKey([]byte(signedInfoXML), certInfo.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("error firmando XML: %v", err)
	}
	signatureValue.SetText(base64.StdEncoding.EncodeToString(signedBytes))

	// KeyInfo
	keyInfo := signature.CreateElement("ds:KeyInfo")
	x509Data := keyInfo.CreateElement("ds:X509Data")
	x509Data.CreateElement("ds:X509Certificate").SetText(formatCertificate(certInfo.Certificate))

	// Serializar firma
	buf.Reset()
	doc = etree.NewDocument()
	doc.SetRoot(signature.Copy())
	if _, err := doc.WriteTo(&buf); err != nil {
		return "", fmt.Errorf("error serializando firma: %v", err)
	}

	return buf.String(), nil
}

func calculateSHA1(data string) []byte {
	h := sha1.New()
	h.Write([]byte(data))
	return h.Sum(nil)
}

func signWithKey(data []byte, key *rsa.PrivateKey) ([]byte, error) {
	h := sha1.New()
	h.Write(data)
	digest := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA1, digest)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func formatCertificate(cert *x509.Certificate) string {
	pemCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	certStr := string(pemCert)
	certStr = strings.ReplaceAll(certStr, "-----BEGIN CERTIFICATE-----\n", "")
	certStr = strings.ReplaceAll(certStr, "\n-----END CERTIFICATE-----\n", "")
	certStr = strings.ReplaceAll(certStr, "\n", "")
	return certStr
}
