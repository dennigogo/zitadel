package domain

import (
	"github.com/zitadel/logging"

	"github.com/dennigogo/zitadel/internal/crypto"
	"github.com/dennigogo/zitadel/internal/errors"
)

type authNKey interface {
	setPublicKey([]byte)
	setPrivateKey([]byte)
	expiration
}

type AuthNKeyType int32

const (
	AuthNKeyTypeNONE AuthNKeyType = iota
	AuthNKeyTypeJSON

	keyCount
)

func (k AuthNKeyType) Valid() bool {
	return k >= 0 && k < keyCount
}

func (key *MachineKey) GenerateNewMachineKeyPair(keySize int) error {
	privateKey, publicKey, err := crypto.GenerateKeyPair(keySize)
	if err != nil {
		return err
	}
	key.PublicKey, err = crypto.PublicKeyToBytes(publicKey)
	if err != nil {
		return err
	}
	key.PrivateKey = crypto.PrivateKeyToBytes(privateKey)
	return nil
}

func SetNewAuthNKeyPair(key authNKey, keySize int) error {
	privateKey, publicKey, err := NewAuthNKeyPair(keySize)
	if err != nil {
		return err
	}
	key.setPrivateKey(privateKey)
	key.setPublicKey(publicKey)
	return nil
}

func NewAuthNKeyPair(keySize int) (privateKey, publicKey []byte, err error) {
	private, public, err := crypto.GenerateKeyPair(keySize)
	if err != nil {
		logging.Log("AUTHN-Ud51I").WithError(err).Error("unable to create authn key pair")
		return nil, nil, errors.ThrowInternal(err, "AUTHN-gdg2l", "Errors.Project.CouldNotGenerateClientSecret")
	}
	publicKey, err = crypto.PublicKeyToBytes(public)
	if err != nil {
		logging.Log("AUTHN-Dbb35").WithError(err).Error("unable to convert public key")
		return nil, nil, errors.ThrowInternal(err, "AUTHN-Bne3f", "Errors.Project.CouldNotGenerateClientSecret")
	}
	privateKey = crypto.PrivateKeyToBytes(private)
	return privateKey, publicKey, nil
}
