package storage

import (
	"encoding/json"
	"errors"
	"os"
)

type ShortenURL struct {
	ID   string `json:"uuid"`
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

func (p *Producer) WriteItem(obj ShortenURL) error {
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

func (c *Consumer) ReadItem() (*ShortenURL, error) {
	obj := ShortenURL{}
	if err := c.decoder.Decode(&obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

type FileStorage struct {
	producer *Producer
	consumer *Consumer
	isActive bool
}

func NewFileStorage(filePath string) (*FileStorage, error) {
	p, err := NewProducer(filePath)
	if err != nil {
		return &FileStorage{
			producer: nil,
			consumer: nil,
			isActive: false,
		}, err
	}

	c, err := NewConsumer(filePath)
	if err != nil {
		return &FileStorage{
			producer: nil,
			consumer: nil,
			isActive: false,
		}, err
	}

	return &FileStorage{
		producer: p,
		consumer: c,
		isActive: true,
	}, err
}

func (fileStorage *FileStorage) ReadAllData(tmp map[string]string) error {
	for {
		readItem, err := fileStorage.consumer.ReadItem()
		if err != nil {
			break
		}
		tmp[readItem.Hash] = readItem.URL
	}
	return nil
}

func (fileStorage *FileStorage) Close() error {
	err1 := fileStorage.consumer.Close()
	err2 := fileStorage.producer.Close()

	err := errors.Join(err1, err2)

	return err
}

func (fileStorage *FileStorage) SaveURL(obj ShortenURL) error {
	err := fileStorage.producer.WriteItem(obj)
	if err != nil {
		return err
	}

	return nil
}
