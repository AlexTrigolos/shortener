package model

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"
)

type URL struct {
	mu              sync.Mutex
	data            map[string]string
	fileStoragePath string
	logger          *slog.Logger
}

func NewURL(fileStoragePath string) (*URL, error) {
	url := &URL{data: make(map[string]string), fileStoragePath: fileStoragePath, logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}
	err := url.load()
	if err != nil {
		url.logger.Error(err.Error())
		return nil, errors.New("не удалось загрузить сохраненные urls сокращения")
	}
	return url, nil
}

func (u *URL) Set(key, value string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	_, exists := u.data[key]
	if exists {
		return errors.New("такой ключ уже занят")
	}
	err := u.save(key, value)
	if err != nil {
		u.logger.Error(err.Error())
		return errors.New("не удалось записать новой сокращенный url")
	}
	u.data[key] = value
	return nil
}

func (u *URL) Get(key string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	value, exists := u.data[key]
	if !exists {
		return "", errors.New("такой ключ не найден")
	}
	return value, nil
}

func (u *URL) load() error {
	data, err := os.ReadFile(u.fileStoragePath)
	if errors.Is(err, os.ErrNotExist) {
		return u.createEmpty()
	} else if err != nil {
		return err
	}

	u.mu.Lock()
	defer u.mu.Unlock()
	jsonData := &[]map[string]string{}
	if err = json.Unmarshal(data, jsonData); err != nil {
		return err
	}
	for _, el := range *jsonData {
		u.data[el["short_url"]] = el["original_url"]
	}
	return nil
}

func (u *URL) createEmpty() error {
	file, err := os.OpenFile(u.fileStoragePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	data, err := json.Marshal([]string{})
	if err != nil {
		return err
	}
	file.Write(data)
	return nil
}

type addingSlice struct {
	key   string
	value string
	first bool
}

// необходим для добавления новой строки в красивом формате
func (s addingSlice) marshalJSON() ([]byte, error) {
	data, err := json.Marshal(map[string]string{"short_url": s.key, "original_url": s.value})
	if err != nil {
		return nil, err
	}
	start := []byte("\n\t")
	if !s.first {
		start = append([]byte("\n,"), start...)
	}
	withMarshal := append(start, data...)
	withEndFile := append(withMarshal, []byte("\n]")...)
	return withEndFile, nil
}

func (u *URL) save(key, value string) error {
	file, err := os.OpenFile(u.fileStoragePath, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	firstURL := len(u.data) == 0
	data, err := addingSlice{key: key, value: value, first: firstURL}.marshalJSON()
	if err != nil {
		return err
	}

	if firstURL {
		file.Seek(-1, io.SeekEnd)
	} else {
		file.Seek(-2, io.SeekEnd)
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
