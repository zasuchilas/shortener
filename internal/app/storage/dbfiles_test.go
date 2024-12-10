package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/zasuchilas/shortener/internal/app/config"
	"github.com/zasuchilas/shortener/internal/app/models"
	"os"
	"testing"
)

func TestDBFiles_InstanceName(t *testing.T) {
	config.FileStoragePath = "./storage_test.db"
	s := NewDBFile()
	defer func() {
		_ = os.Remove(config.FileStoragePath)
	}()
	assert.Equal(t, InstanceFile, s.InstanceName())
	_ = os.Remove(config.FileStoragePath)
}

func TestDBFiles_Ping(t *testing.T) {
	config.FileStoragePath = "./storage_test.db"
	s := NewDBFile()
	defer func() {
		_ = os.Remove(config.FileStoragePath)
	}()
	err := s.Ping(context.TODO())
	assert.Error(t, err)
	assert.Equal(t, "not allowed", err.Error())
	_ = os.Remove(config.FileStoragePath)
}

func TestDBFiles_WriteURL(t *testing.T) {
	config.FileStoragePath = "./storage_test.db"
	s := NewDBFile()
	defer func() {
		_ = os.Remove(config.FileStoragePath)
	}()

	tests := []struct {
		name     string
		origURL  string
		userID   int64
		shortURL string
		conflict bool
		err      error
	}{
		{
			name:     "valid write",
			origURL:  "https://ya.ru",
			shortURL: "19xtf1ts",
			userID:   1,
			conflict: false,
			err:      nil,
		},
		{
			name:     "wrong repeated write",
			origURL:  "https://ya.ru",
			shortURL: "19xtf1ts",
			userID:   1,
			conflict: true,
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortURL, conflict, err := s.WriteURL(context.TODO(), tt.origURL, tt.userID)

			assert.Equal(t, tt.shortURL, shortURL)
			assert.Equal(t, tt.conflict, conflict)
			assert.NoError(t, err)
		})
	}
}

func TestDBFiles_ReadURL(t *testing.T) {
	config.FileStoragePath = "./storage_test.db"
	s := NewDBFile()
	defer func() {
		_ = os.Remove(config.FileStoragePath)
	}()

	// select from empty db
	_, err := s.ReadURL(context.TODO(), "19xtf1ts")
	assert.Error(t, ErrNotFound, err)

	// creating URL row
	_, _, _ = s.WriteURL(
		context.TODO(),
		"https://ya.ru",
		1,
	)

	//
	origURL, err := s.ReadURL(context.TODO(), "19xtf1ts")
	assert.NoError(t, err)
	assert.Equal(t, "https://ya.ru", origURL)
}

func TestDBFiles_UserURLs(t *testing.T) {
	config.FileStoragePath = "./storage_test.db"
	s := NewDBFile()
	defer func() {
		_ = os.Remove(config.FileStoragePath)
	}()

	tests := []struct {
		name         string
		userID       int64
		expectedURLs []*models.URLRow
		expectedErr  error
	}{
		{
			name:         "nil",
			userID:       0,
			expectedURLs: nil,
			expectedErr:  ErrNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actual, err := s.UserURLs(context.TODO(), tt.userID)

			assert.Equal(t, tt.expectedURLs, actual)
			assert.Error(t, tt.expectedErr, err)
		})
	}
}

func TestDBFiles_CheckDeletedURLs(t *testing.T) {
	config.FileStoragePath = "./storage_test.db"
	s := NewDBFile()
	defer func() {
		_ = os.Remove(config.FileStoragePath)
	}()

	tests := []struct {
		name        string
		userID      int64
		shortURLs   []string
		expectedErr error
	}{
		{
			name:        "nil",
			userID:      0,
			shortURLs:   nil,
			expectedErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := s.CheckDeletedURLs(context.TODO(), tt.userID, tt.shortURLs)

			assert.Error(t, tt.expectedErr, err)
		})
	}
}

func TestDBFiles_DeleteURLs(t *testing.T) {
	config.FileStoragePath = "./storage_test.db"
	s := NewDBFile()
	defer func() {
		_ = os.Remove(config.FileStoragePath)
	}()

	tests := []struct {
		name        string
		shortURLs   []string
		expectedErr error
	}{
		{
			name:        "nil",
			shortURLs:   nil,
			expectedErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := s.DeleteURLs(context.TODO(), tt.shortURLs...)

			assert.Error(t, tt.expectedErr, err)
		})
	}
}
