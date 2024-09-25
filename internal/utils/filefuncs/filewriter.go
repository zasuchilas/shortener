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
	logger.Log.Debug("opening file storage as file writer")
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("making FileWriter with json encoder")
	return &FileWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *FileWriter) WriteURLRow(uuid int64, shortURL, origURL string) error {
	row := models.URLRow{
		UUID:     uuid,
		ShortURL: shortURL,
		OrigURL:  origURL,
	}
	return p.encoder.Encode(&row)
}

func (p *FileWriter) Close() error {
	logger.Log.Debug("FileWriter closing")
	return p.file.Close()
}
