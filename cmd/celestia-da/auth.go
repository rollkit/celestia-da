package main

import (
	"crypto/rand"
	"errors"
	"io"
	"path/filepath"

	"github.com/celestiaorg/celestia-node/api/rpc/perms"
	"github.com/celestiaorg/celestia-node/libs/authtoken"
	"github.com/celestiaorg/celestia-node/libs/keystore"
	nodemod "github.com/celestiaorg/celestia-node/nodebuilder/node"

	"github.com/cristalhq/jwt"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/mitchellh/go-homedir"
)

func buildJWTToken(body []byte, permissions []auth.Permission) (string, error) {
	signer, err := jwt.NewHS256(body)
	if err != nil {
		return "", err
	}
	return authtoken.NewSignedJWT(signer, permissions)
}

func generateNewKey(ks keystore.Keystore) (keystore.PrivKey, error) {
	sk, err := io.ReadAll(io.LimitReader(rand.Reader, 32))
	if err != nil {
		return keystore.PrivKey{}, err
	}
	// save key
	key := keystore.PrivKey{Body: sk}
	err = ks.Put(nodemod.SecretName, key)
	if err != nil {
		return keystore.PrivKey{}, err
	}
	return key, nil
}

func newKeystore(path string) (keystore.Keystore, error) {
	expanded, err := homedir.Expand(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	return keystore.NewFSKeystore(filepath.Join(expanded, "keys"), nil)
}

func authToken(path string) (string, error) {
	ks, err := newKeystore(path)
	if err != nil {
		return "", err
	}

	key, err := ks.Get(nodemod.SecretName)
	if err != nil {
		if !errors.Is(err, keystore.ErrNotFound) {
			return "", err
		}
		key, err = generateNewKey(ks)
		if err != nil {
			return "", err
		}
	}

	token, err := buildJWTToken(key.Body, perms.ReadWritePerms)
	if err != nil {
		return "", err
	}
	return token, nil
}
