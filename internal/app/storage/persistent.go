package storage

import (
	"encoding/json"
	"github.com/zasuchilas/shortener/internal/app/logger"
	"os"
)

type URLRow struct {
	UUID        uint   `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(filename string) (*Producer, error) {
	logger.Log.Debug("opening file storage with producer (for write)")
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("making Producer with json encoder")
	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) WriteURLRow(shortURL, origURL string) error {
	row := URLRow{
		UUID:        1,
		ShortURL:    shortURL,
		OriginalURL: origURL,
	}
	return p.encoder.Encode(&row)
}

func (p *Producer) Close() error {
	logger.Log.Debug("producer closing")
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(filename string) (*Consumer, error) {
	logger.Log.Debug("opening file storage with consumer (for read)")
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("making Consumer with json decoder")
	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) ReadURLRow() (*URLRow, error) {
	ur := &URLRow{}
	if err := c.decoder.Decode(ur); err != nil {
		return nil, err
	}
	return ur, nil
}

func (c *Consumer) Close() error {
	logger.Log.Debug("consumer closing")
	return c.file.Close()
}
