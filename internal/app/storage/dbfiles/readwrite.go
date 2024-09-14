package dbfiles

import (
	"encoding/json"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"os"
)

type fileWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func newFileWriter(filename string) (*fileWriter, error) {
	logger.Log.Debug("opening file storage as file writer")
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("making fileWriter with json encoder")
	return &fileWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *fileWriter) writeURLRow(uuid int64, shortURL, origURL string) error {
	row := models.URLRow{
		Uuid:     uuid,
		ShortURL: shortURL,
		OrigURL:  origURL,
	}
	return p.encoder.Encode(&row)
}

func (p *fileWriter) close() error {
	logger.Log.Debug("fileWriter closing")
	return p.file.Close()
}

type fileReader struct {
	file    *os.File
	decoder *json.Decoder
}

func newFileReader(filename string) (*fileReader, error) {
	logger.Log.Debug("opening file storage as file reader")
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("making fileReader with json decoder")
	return &fileReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *fileReader) readURLRow() (*models.URLRow, error) {
	ur := &models.URLRow{}
	if err := c.decoder.Decode(ur); err != nil {
		return nil, err
	}
	return ur, nil
}

func (c *fileReader) close() error {
	logger.Log.Debug("fileReader closing")
	return c.file.Close()
}
