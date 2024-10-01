package filefuncs

import (
	"encoding/json"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"os"
)

type FileWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func NewFileWriter(filename string) (*FileWriter, error) {
	return newFileWriter(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
}

func NewFileReWriter(filename string) (*FileWriter, error) {
	return newFileWriter(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_TRUNC, 0666)
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

func (p *FileWriter) Close() error {
	logger.Log.Debug("FileWriter closing")
	return p.file.Close()
}

func (p *FileWriter) WriteURLRow(url *models.URLRow) error {
	return p.encoder.Encode(url)
}

func (p *FileWriter) WriteUserRow(user *models.UserRow) error {
	return p.encoder.Encode(user)
}
