package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/storage"
	"github.com/zasuchilas/shortener/internal/app/storage/hashfuncs"
	"go.uber.org/zap"
)

type Secure struct {
	key32     []byte
	aesbloc   cipher.Block
	aesgcm    cipher.AEAD
	nonceSize int

	store storage.Storage
}

func New(key string, store storage.Storage) *Secure {

	// creating a 32 byte key AES for using AES-256
	k32sl := sha256.Sum256([]byte(key)) // config.SecretKey
	key32 := k32sl[:]

	// creating AES-256 cipher.Block
	aesbloc, err := aes.NewCipher(key32)
	if err != nil {
		logger.Log.Fatal("creating AES bloc", zap.Error(err))
	}

	// creating GCM for AES-256
	aesgcm, err := cipher.NewGCM(aesbloc)
	if err != nil {
		logger.Log.Fatal("creating GCM for AES-256", zap.Error(err))
	}

	return &Secure{
		key32:     key32,
		aesbloc:   aesbloc,
		aesgcm:    aesgcm,
		nonceSize: aesgcm.NonceSize(),
		store:     store,
	}
}

func (s *Secure) Encrypt(src []byte) (encrypted, nonce []byte, err error) {

	nonce, err = generateRandom(s.aesgcm.NonceSize())
	if err != nil {
		logger.Log.Error("creating nonce", zap.Error(err))
		// TODO: use advanced error handling (is, as ...)
		return nil, nil, err
	}
	logger.Log.Debug("creating nonce before encryption", zap.ByteString("nonce", nonce))

	encrypted = s.EncryptCustom(src, nonce)
	return encrypted, nonce, nil
}

func (s *Secure) EncryptCustom(src, nonce []byte) (encrypted []byte) {
	encrypted = s.aesgcm.Seal(nil, nonce, src, nil)
	return encrypted
}

func (s *Secure) Decrypt(encrypted, nonce []byte) (decrypted []byte, err error) {

	// decrypting data
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

func (s *Secure) packTokenCookieData(userID int64, nonce []byte) (hexadecimal string) {

	hashUserID := hashfuncs.EncodeHeroHash(userID)
	logger.Log.Debug("encoding userID as hash", zap.String("hashUserID", hashUserID))

	encrypted := s.EncryptCustom([]byte(hashUserID), nonce)
	logger.Log.Debug("encrypting hash with userID by AES", zap.ByteString("encrypted", encrypted))

	tokenBytes := append(encrypted, nonce...)
	logger.Log.Debug("getting token bytes", zap.ByteString("tokenBytes", tokenBytes))

	return hex.EncodeToString(tokenBytes)
}

func (s *Secure) unpackTokenCookieData(hexadecimal string) (userID int64, err error) {

	tokenBytes, err := hex.DecodeString(hexadecimal)
	if err != nil {
		logger.Log.Debug("decoding from hex", zap.Error(err))
		return 0, err
	}
	logger.Log.Debug("decoded from hex", zap.String("hexadecimal", hexadecimal))

	// getting nonce and payload
	tbLen := len(tokenBytes)
	if tbLen <= s.nonceSize {
		return 0, errors.New("token cookie value is too small")
	}
	nonce := tokenBytes[tbLen-s.nonceSize:]
	payload := tokenBytes[:tbLen-s.nonceSize]

	hashUserID, err := s.Decrypt(payload, nonce)
	if err != nil {
		logger.Log.Debug("decrypting token cookie (userID as hash)", zap.Error(err))
		return 0, err
	}
	logger.Log.Debug("decrypting token cookie (userID as hash)", zap.ByteString("hashUserID", hashUserID))

	userID, err = hashfuncs.DecodeHeroHash(string(hashUserID))
	if err != nil {
		logger.Log.Debug("getting userID as int64", zap.Error(err))
	}
	logger.Log.Debug("getting userID as int64", zap.Int64("userID", userID))

	return userID, nil
}
