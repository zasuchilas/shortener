package filefuncs

import (
	"encoding/json"
	"os"

	"github.com/zasuchilas/shortener/internal/app/model"

	"github.com/zasuchilas/shortener/internal/app/logger"
)

// FileReader is the structure for reading json strings from a storage file.
type FileReader struct {
	file    *os.File
	decoder *json.Decoder
}

// NewFileReader is the FileReader constructor.
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

// Close _
func (c *FileReader) Close() error {
	logger.Log.Debug("FileReader closing")
	return c.file.Close()
}

// ReadURLRow reads the URL string from the storage file.
func (c *FileReader) ReadURLRow() (*model.URLRow, error) {
	ur := &model.URLRow{}
	if err := c.decoder.Decode(ur); err != nil {
		return nil, err
	}
	return ur, nil
}

// ReadUserRow reads the user string from the storage file.
func (c *FileReader) ReadUserRow() (*model.UserRow, error) {
	ur := &model.UserRow{}
	if err := c.decoder.Decode(ur); err != nil {
		return nil, err
	}
	return ur, nil
}
