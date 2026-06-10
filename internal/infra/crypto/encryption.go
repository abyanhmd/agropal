package crypto

import (
	"agropal/config"
	"agropal/internal/domain/shared"
	"agropal/pkg"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
)

type Encryption struct {
	logger *pkg.Logger
	config *config.Config
}

func NewEncryption(
	logger *pkg.Logger,
	config *config.Config,
) *Encryption {
	return &Encryption{
		logger: logger,
		config: config,
	}
}

type EncryptionType int

const (
	EncRefreshToken EncryptionType = iota
)

func deriveAESKey(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

func (e *Encryption) EncryptAES(ctx context.Context, encType EncryptionType, text string) (string, error) {
	var encTypeMap = map[EncryptionType]string{
		EncRefreshToken: e.config.Secret.GcmToken,
	}
	key := deriveAESKey(encTypeMap[encType])

	block, errAes := aes.NewCipher(key)
	if errAes != nil {
		e.logger.Error(ctx, shared.System, errAes.Error())
		return "", errAes
	}

	nonce := make([]byte, 12)
	if _, errRead := io.ReadFull(rand.Reader, nonce); errRead != nil {
		e.logger.Error(ctx, shared.System, errRead.Error())
		return "", errRead
	}

	aesGCM, errGcm := cipher.NewGCM(block)
	if errGcm != nil {
		e.logger.Error(ctx, shared.System, errGcm.Error())
		return "", errGcm
	}

	cipherText := aesGCM.Seal(nil, nonce, []byte(text), nil)

	finalCipherText := append(nonce, cipherText...)
	return base64.StdEncoding.EncodeToString(finalCipherText), nil
}

func (e *Encryption) DecryptAES(ctx context.Context, encType EncryptionType, text string) (string, error) {
	var encTypeMap = map[EncryptionType]string{
		EncRefreshToken: e.config.Secret.GcmToken,
	}
	key := deriveAESKey(encTypeMap[encType])

	data, errBase64 := base64.StdEncoding.DecodeString(text)
	if errBase64 != nil {
		e.logger.Error(ctx, shared.System, errBase64.Error())
		return "", errBase64
	}

	block, errAes := aes.NewCipher(key)
	if errAes != nil {
		e.logger.Error(ctx, shared.System, errAes.Error())
		return "", errAes
	}

	nonce, cipherText := data[:12], data[12:]

	aesGCM, errGcm := cipher.NewGCM(block)
	if errGcm != nil {
		e.logger.Error(ctx, shared.System, errGcm.Error())
		return "", errGcm
	}

	plainText, errDecrypt := aesGCM.Open(nil, nonce, cipherText, nil)
	if errDecrypt != nil {
		e.logger.Error(ctx, shared.System, errDecrypt.Error())
		return "", errDecrypt
	}

	return string(plainText), nil
}
