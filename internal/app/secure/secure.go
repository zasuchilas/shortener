package secure

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"github.com/zasuchilas/shortener/internal/app/utils/filefuncs"
	"github.com/zasuchilas/shortener/internal/app/utils/hashfuncs"
)

// Secure is the component structure.
type Secure struct {
	key32     []byte
	aesbloc   cipher.Block
	aesgcm    cipher.AEAD
	nonceSize int

	persist             bool
	filePath            string
	users               map[int64]*models.UserRow
	lastUserID          int64
	storageInstanceName string
	mutex               sync.RWMutex
}

// New creates an instance of the component.
func New(key, storageInstanceName string, filePath string) *Secure {

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

	var persist bool
	if filePath != "" {
		persist = true
	}

	sec := &Secure{
		key32:               key32,
		aesbloc:             aesbloc,
		aesgcm:              aesgcm,
		nonceSize:           aesgcm.NonceSize(),
		persist:             persist,
		filePath:            filePath,
		users:               make(map[int64]*models.UserRow),
		storageInstanceName: storageInstanceName,
		mutex:               sync.RWMutex{},
	}

	lastUserID, err := sec.loadFromFile()
	if err != nil {
		logger.Log.Fatal("loading user data from file", zap.Error(err))
	}
	sec.lastUserID = lastUserID

	return sec
}

// Encrypt encrypts the UserID value.
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

// EncryptCustom encrypts src value with nonce.
func (s *Secure) EncryptCustom(src, nonce []byte) (encrypted []byte) {
	encrypted = s.aesgcm.Seal(nil, nonce, src, nil)
	return encrypted
}

// Decrypt decrypts encrypted with nonce.
func (s *Secure) Decrypt(encrypted, nonce []byte) (decrypted []byte, err error) {

	// decrypting data
	decrypted, err = s.aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		logger.Log.Error("decrypting data", zap.Error(err))
		return nil, err
	}

	return decrypted, nil
}

// NewUser creates new user in secure storage.
func (s *Secure) NewUser(_ context.Context) (userID int64, err error) {
	// starting ~tx
	s.mutex.Lock()
	defer s.mutex.Unlock()

	nextUserID := s.lastUserID + 1
	userHash := hashfuncs.EncodeHeroHash(nextUserID)
	nextUser := &models.UserRow{
		UserID:   nextUserID,
		UserHash: userHash,
		UserDB:   s.storageInstanceName,
	}

	err = s.writeUserPersist(nextUser)
	if err != nil {
		return 0, err
	}

	s.users[nextUserID] = nextUser
	s.lastUserID = nextUserID
	logger.Log.Debug("inserted new user row",
		zap.Int64("userID", nextUserID),
		zap.String("userHash", userHash),
		zap.String("userDB", ""))

	return nextUserID, nil
}

// CheckUser checks user in the secure storage.
func (s *Secure) CheckUser(_ context.Context, userID int64, userHash string) (found bool, err error) {
	user, ok := s.users[userID]
	if !ok {
		return false, nil
	}

	if user.UserHash != userHash {
		return false,
			fmt.Errorf("checking user data: unexpected user hash (%d -> %s <> %s)",
				userID, user.UserHash, userHash)
	}

	return true, err
}

// packTokenCookieData packs a token data for cookie.
func (s *Secure) packTokenCookieData(userID int64, nonce []byte) (hexadecimal string) {

	hashUserID := hashfuncs.EncodeHeroHash(userID)
	logger.Log.Debug("encoding userID as hash", zap.String("hashUserID", hashUserID))

	encrypted := s.EncryptCustom([]byte(hashUserID), nonce)
	logger.Log.Debug("encrypting hash with userID by AES", zap.ByteString("encrypted", encrypted))

	tokenBytes := append(encrypted, nonce...)
	logger.Log.Debug("getting token bytes", zap.ByteString("tokenBytes", tokenBytes))

	return hex.EncodeToString(tokenBytes)
}

// unpackTokenCookieData unpacks a token data for cookie.
func (s *Secure) unpackTokenCookieData(hexadecimal string) (userID int64, userHash string, err error) {

	tokenBytes, err := hex.DecodeString(hexadecimal)
	if err != nil {
		logger.Log.Debug("decoding from hex", zap.Error(err))
		return 0, "", err
	}
	logger.Log.Debug("decoded from hex", zap.String("hexadecimal", hexadecimal))

	// getting nonce and payload
	tbLen := len(tokenBytes)
	if tbLen <= s.nonceSize {
		return 0, "", errors.New("token cookie value is too small")
	}
	nonce := tokenBytes[tbLen-s.nonceSize:]
	payload := tokenBytes[:tbLen-s.nonceSize]

	userHashBytes, err := s.Decrypt(payload, nonce)
	userHash = string(userHashBytes)
	if err != nil {
		logger.Log.Debug("decrypting token cookie (userID as hash)", zap.Error(err))
		return 0, "", err
	}
	logger.Log.Debug("decrypting token cookie (userID as hash)", zap.String("userHash", userHash))

	userID, err = hashfuncs.DecodeHeroHash(userHash)
	if err != nil {
		logger.Log.Debug("getting userID as int64", zap.Error(err))
		return 0, "", err
	}
	logger.Log.Debug("getting userID as int64", zap.Int64("userID", userID))

	return userID, userHash, nil
}

// loadFromFile load users from file storage into memory.
func (s *Secure) loadFromFile() (lastUserID int64, err error) {
	if !s.persist {
		return 0, nil
	}

	r, err := filefuncs.NewFileReader(s.filePath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	var lastUserHash string
	for {
		row, e := r.ReadUserRow()
		if e == io.EOF {
			break
		}
		if e != nil {
			err = e
			logger.Log.Debug("reading users from file", zap.Error(e))
			break
		}

		s.users[row.UserID] = row
		lastUserHash = row.UserHash
	}

	if lastUserHash == "" {
		return 0, nil
	}

	lastUserID, err = hashfuncs.DecodeHeroHash(lastUserHash)
	if err != nil {
		return 0, err
	}

	return lastUserID, nil
}

// writeUserPersist writes user data into secure storage file.
func (s *Secure) writeUserPersist(user *models.UserRow) error {
	if !s.persist {
		return nil
	}

	w, err := filefuncs.NewFileWriter(s.filePath)
	if err != nil {
		return err
	}
	defer w.Close()

	return w.WriteUserRow(user)
}

// generateRandom generates random bytes.
func generateRandom(size int) ([]byte, error) {
	// generating cryptographically strong random bytes in b
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
