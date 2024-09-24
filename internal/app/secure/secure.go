package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"go.uber.org/zap"
)

type Secure struct {
	key32   []byte
	aesbloc cipher.Block
	aesgcm  cipher.AEAD
}

func New(key string) *Secure {

	logger.Log.Debug("creating a 32 byte key AES for using AES-256")
	k32sl := sha256.Sum256([]byte(key)) // config.SecretKey
	key32 := k32sl[:]

	logger.Log.Debug("creating AES-256 cipher.Block")
	aesbloc, err := aes.NewCipher(key32)
	if err != nil {
		logger.Log.Fatal("creating AES bloc", zap.Error(err))
	}

	logger.Log.Debug("creating GCM for AES-256")
	aesgcm, err := cipher.NewGCM(aesbloc)
	if err != nil {
		logger.Log.Fatal("creating GCM for AES-256", zap.Error(err))
	}

	return &Secure{
		key32:   key32,
		aesbloc: aesbloc,
		aesgcm:  aesgcm,
	}
}

func (s *Secure) Encrypt(src []byte) (encrypted, nonce []byte, err error) {

	logger.Log.Debug("creating nonce before encryption")
	nonce, err = generateRandom(s.aesgcm.NonceSize())
	if err != nil {
		logger.Log.Error("creating nonce", zap.Error(err))
		// TODO: use advanced error handling (is, as ...)
		return nil, nil, err
	}

	logger.Log.Debug("encrypting src")
	encrypted = s.aesgcm.Seal(nil, nonce, src, nil)
	return encrypted, nonce, nil
}

func (s *Secure) Decrypt(encrypted, nonce []byte) (decrypted []byte, err error) {

	logger.Log.Debug("decrypting data")
	decrypted, err = s.aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		logger.Log.Error("decrypting data", zap.Error(err))
		return nil, err
	}

	return decrypted, nil
}

func generateRandom(size int) ([]byte, error) {
	// generating cryptographically strong random bytes in b
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
