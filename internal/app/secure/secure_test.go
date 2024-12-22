package secure

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/models"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecure(t *testing.T) {
	tests := []struct {
		name string
		key  string
		data string
	}{
		{
			name: "positive: test 1",
			key:  "supersecretkey",
			data: "1234567890",
		},
		{
			name: "positive: test 2",
			key:  "supersecretkey",
			data: "ABCabc123_#$",
		},
		{
			name: "positive: small key < 32 byte",
			key:  "key",
			data: "1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secure := New(tt.key, "", "")

			encrypted, nonce, err := secure.Encrypt([]byte(tt.data))
			assert.NoError(t, err)

			decrypted, err2 := secure.Decrypt(encrypted, nonce)
			assert.NoError(t, err2)

			assert.Equal(t, tt.data, string(decrypted))
		})
	}
}

func TestPacking(t *testing.T) {
	tests := []struct {
		name   string
		userID int64
	}{
		{
			name:   "positive: test 1",
			userID: 1,
		},
		{
			name:   "positive: test 2",
			userID: 100500,
		},
	}

	key := "supersecretkey"
	secure := New(key, "", "")
	nonce, e := generateRandom(secure.nonceSize)
	assert.NoError(t, e)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hexadecimal := secure.packTokenCookieData(tt.userID, nonce)
			userID, _, err := secure.unpackTokenCookieData(hexadecimal)

			assert.NoError(t, err)
			assert.Equal(t, tt.userID, userID)
		})
	}
}

func BenchmarkSecure_Encrypt(b *testing.B) {
	secure := New("supersecretkey", "", "")
	data := []byte("1234567890")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _ = secure.Encrypt(data)
	}
}

func BenchmarkSecure_Decrypt(b *testing.B) {
	secure := New("supersecretkey", "", "")
	data := []byte("1234567890")
	encrypted, nonce, _ := secure.Encrypt(data)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = secure.Decrypt(encrypted, nonce)
	}
}

func TestSecure_NewUser(t *testing.T) {
	type fields struct {
		key32               []byte
		aesbloc             cipher.Block
		aesgcm              cipher.AEAD
		nonceSize           int
		persist             bool
		filePath            string
		users               map[int64]*models.UserRow
		lastUserID          int64
		storageInstanceName string
		mutex               sync.RWMutex
	}
	type args struct {
		in0 context.Context
	}

	key := "supersecretkey"
	k32 := sha256.Sum256([]byte(key))
	aesbloc, _ := aes.NewCipher(k32[:])
	aesgcm, _ := cipher.NewGCM(aesbloc)

	tests := []struct {
		name       string
		args       args
		wantUserID int64
		wantErr    bool
	}{
		{
			name:       "nil",
			args:       args{in0: context.TODO()},
			wantUserID: 1,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Secure{
				key32:               []byte(key),
				aesbloc:             aesbloc,
				aesgcm:              aesgcm,
				nonceSize:           aesgcm.NonceSize(),
				persist:             false,
				filePath:            "",
				users:               make(map[int64]*models.UserRow),
				lastUserID:          0,
				storageInstanceName: "",
				mutex:               sync.RWMutex{},
			}
			gotUserID, err := s.NewUser(tt.args.in0)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("NewUser() gotUserID = %v, want %v", gotUserID, tt.wantUserID)
			}
		})
	}
}

func TestSecure_CheckUser(t *testing.T) {
	type fields struct {
		key32               []byte
		aesbloc             cipher.Block
		aesgcm              cipher.AEAD
		nonceSize           int
		persist             bool
		filePath            string
		users               map[int64]*models.UserRow
		lastUserID          int64
		storageInstanceName string
		mutex               sync.RWMutex
	}
	type args struct {
		in0      context.Context
		userID   int64
		userHash string
	}

	key := "supersecretkey"
	k32 := sha256.Sum256([]byte(key))
	aesbloc, _ := aes.NewCipher(k32[:])
	aesgcm, _ := cipher.NewGCM(aesbloc)
	config.SecureFilePath = "./secure_test.db"

	tests := []struct {
		name      string
		args      args
		wantFound bool
		wantErr   bool
	}{
		{
			name: "nil",
			args: args{
				in0:      context.TODO(),
				userID:   0,
				userHash: "",
			},
			wantErr: false,
		},
		{
			name: "err",
			args: args{
				in0:      context.TODO(),
				userID:   1,
				userHash: "q",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Secure{
				key32:               []byte(key),
				aesbloc:             aesbloc,
				aesgcm:              aesgcm,
				nonceSize:           aesgcm.NonceSize(),
				persist:             true,
				filePath:            config.SecureFilePath,
				users:               make(map[int64]*models.UserRow),
				lastUserID:          1,
				storageInstanceName: "",
				mutex:               sync.RWMutex{},
			}
			gotFound, err := s.CheckUser(tt.args.in0, tt.args.userID, tt.args.userHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFound != tt.wantFound {
				t.Errorf("CheckUser() gotFound = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

func TestSecure_loadFromFile(t *testing.T) {
	type fields struct {
		key32               []byte
		aesbloc             cipher.Block
		aesgcm              cipher.AEAD
		nonceSize           int
		persist             bool
		filePath            string
		users               map[int64]*models.UserRow
		lastUserID          int64
		storageInstanceName string
		mutex               sync.RWMutex
	}

	key := "supersecretkey"
	k32 := sha256.Sum256([]byte(key))
	aesbloc, _ := aes.NewCipher(k32[:])
	aesgcm, _ := cipher.NewGCM(aesbloc)
	config.SecureFilePath = "./secure_test.db"

	tests := []struct {
		name           string
		wantLastUserID int64
		wantErr        bool
	}{
		{
			name:           "nil",
			wantLastUserID: 0,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Secure{
				key32:               []byte(key),
				aesbloc:             aesbloc,
				aesgcm:              aesgcm,
				nonceSize:           aesgcm.NonceSize(),
				persist:             true,
				filePath:            config.SecureFilePath,
				users:               make(map[int64]*models.UserRow),
				lastUserID:          0,
				storageInstanceName: "",
				mutex:               sync.RWMutex{},
			}
			gotLastUserID, err := s.loadFromFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("loadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotLastUserID != tt.wantLastUserID {
				t.Errorf("loadFromFile() gotLastUserID = %v, want %v", gotLastUserID, tt.wantLastUserID)
			}
		})
	}
}
