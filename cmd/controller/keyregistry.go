package main

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"log"

	"k8s.io/client-go/kubernetes"
	certUtil "k8s.io/client-go/util/cert"
)

type KeyRegistry struct {
	client         kubernetes.Interface
	namespace      string
	keyPrefix      string
	keyLabel       string
	keysize        int
	currentKeyName string
	keys           map[string]*rsa.PrivateKey
	cert           *x509.Certificate
}

func NewKeyRegistry(client kubernetes.Interface, namespace, keyPrefix, keyLabel string, keysize int) *KeyRegistry {
	return &KeyRegistry{
		client:    client,
		namespace: namespace,
		keyPrefix: keyPrefix,
		keysize:   keysize,
		keyLabel:  keyLabel,
		keys:      map[string]*rsa.PrivateKey{},
	}
}

func (kr *KeyRegistry) generateKey() (string, error) {
	key, cert, err := generatePrivateKeyAndCert(kr.keysize)
	if err != nil {
		return "", err
	}
	certs := []*x509.Certificate{cert}
	generatedName, err := writeKey(kr.client, key, certs, kr.namespace, kr.keyLabel, kr.keyPrefix)
	if err != nil {
		return "", err
	}
	// Only store key to local store if write to k8s workedk
	kr.registerNewKey(generatedName, key, cert)
	log.Printf("New key written to %s/%s\n", kr.namespace, generatedName)
	log.Printf("Certificate is \n%s\n", certUtil.EncodeCertPEM(cert))
	return generatedName, nil
}

func (kr *KeyRegistry) registerNewKey(keyName string, privKey *rsa.PrivateKey, cert *x509.Certificate) {
	kr.keys[keyName] = privKey
	kr.cert = cert
	kr.currentKeyName = keyName
}

func (kr *KeyRegistry) latestKeyName() string {
	return kr.currentKeyName
}

func (kr *KeyRegistry) getPrivateKey(keyname string) (*rsa.PrivateKey, error) {
	key, ok := kr.keys[keyname]
	if !ok {
		return nil, fmt.Errorf("No key exists with name %s", keyname)
	}
	return key, nil
}

func (kr *KeyRegistry) getCert(keyname string) (*x509.Certificate, error) {
	return kr.cert, nil
}
