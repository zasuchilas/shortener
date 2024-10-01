package filefuncs

import (
	"encoding/json"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/models"
	"os"
)

type FileReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewFileReader(filename string) (*FileReader, error) {
	logger.Log.Debug("opening file storage as file reader")
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("making FileReader with json decoder")
	return &FileReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *FileReader) Close() error {
	logger.Log.Debug("FileReader closing")
	return c.file.Close()
}

func (c *FileReader) ReadURLRow() (*models.URLRow, error) {
	ur := &models.URLRow{}
	if err := c.decoder.Decode(ur); err != nil {
		return nil, err
	}
	return ur, nil
}

func (c *FileReader) ReadUserRow() (*models.UserRow, error) {
	ur := &models.UserRow{}
	if err := c.decoder.Decode(ur); err != nil {
		return nil, err
	}
	return ur, nil
}
