package storage

import (
	"encoding/json"
	"os"
)

type shortenUrl struct {
	Id   string `json:"uuid"`
	Hash string `json:"hash"`
	URL  string `json:"url"`
}

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) WriteItem(obj shortenUrl) error {
	return p.encoder.Encode(obj)
}

func (p *Producer) Close() error {
	return p.file.Close()
}

type Consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) ReadItem() (*shortenUrl, error) {
	obj := shortenUrl{}
	if err := c.decoder.Decode(&obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}
