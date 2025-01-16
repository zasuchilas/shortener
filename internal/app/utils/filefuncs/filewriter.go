package filefuncs

import (
	"encoding/json"
	"os"

	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/model"
)

// FileWriter is the structure for writing json strings in a storage file.
type FileWriter struct {
	file    *os.File
	encoder *json.Encoder
}

// NewFileWriter is the FileWriter constructor.
func NewFileWriter(filename string) (*FileWriter, error) {
	return newFileWriter(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
}

// NewFileReWriter is the FileWriter constructor.
func NewFileReWriter(filename string) (*FileWriter, error) {
	return newFileWriter(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
}

// Close _
func (p *FileWriter) Close() error {
	logger.Log.Debug("FileWriter closing")
	return p.file.Close()
}

// WriteURLRow writes the URL string in the storage file.
func (p *FileWriter) WriteURLRow(url *model.URLRow) error {
	return p.encoder.Encode(url)
}

// WriteUserRow writes the user string in the storage file.
func (p *FileWriter) WriteUserRow(user *model.UserRow) error {
	return p.encoder.Encode(user)
}

func newFileWriter(filename string, flag int, perm os.FileMode) (*FileWriter, error) {
	logger.Log.Debug("opening file storage as file writer")
	file, err := os.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("making FileWriter with json encoder")
	return &FileWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}
