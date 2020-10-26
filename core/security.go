package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"io"
	"sale_ranking/pkg/util/crypto"

	"github.com/Luzifer/go-openssl"
)

func InitServerKey() error {
	var sKey string
	if err := redis.Get(serverKeyCache, &sKey); err != nil {
		if serverKey == "" {
			sKey = crypto.GenSecretString(32)
		} else {
			sKey = serverKey
		}
		if err := redis.Set(serverKey, sKey, 0); err != nil {
			return err
		}
	}
	serverKey = sKey
	return nil
}

func InitRSA() error {
	if err := redis.Get(signKeyCache, &signKey); err != nil {
		if redis.IsKeyNotFound(err) {
			signKey.Key, _ = rsa.GenerateKey(rand.Reader, 2048)
			if err := redis.Set(signKeyCache, signKey, 0); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func InitSecurityKey() error {
	// Shared Key - server key
	if err := InitServerKey(); err != nil {
		return err
	}
	// RSA
	if err := InitRSA(); err != nil {
		return err
	}
	return nil
}

func EncryptWithKey(data []byte, key string) (string, error) {
	gcmCipher, err := genGcmCipher(key)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcmCipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(gcmCipher.Seal(nonce, nonce, data, nil)), nil
}

func EncryptWithServerKey(data []byte) (string, error) {
	return EncryptWithKey(data, serverKey)
}

func DecryptWithKey(cipherText string, key string) ([]byte, error) {
	cipherByte, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}
	var data []byte
	gcmCipher, err := genGcmCipher(key)
	if err != nil {
		return data, err
	}
	nonceSize := gcmCipher.NonceSize()
	if len(cipherByte) < nonceSize {
		return nil, errors.New("invalid string size")
	}
	nonce, cipherByte := cipherByte[:nonceSize], cipherByte[nonceSize:]
	return gcmCipher.Open(nil, nonce, cipherByte, nil)
}

func DecryptWithServerKey(cipherText string) ([]byte, error) {
	return DecryptWithKey(cipherText, serverKey)
}

func AESDecryptCipherByKey(cipherText []byte, key string) ([]byte, error) {
	o := openssl.New()
	dec, err := o.DecryptBytes(key, cipherText)
	if err != nil {
		return nil, err
	}
	return dec, nil
}

func genGcmCipher(key string) (cipher.AEAD, error) {
	cipherBlock, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	gcmCipher, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return nil, err
	}
	return gcmCipher, nil
}
