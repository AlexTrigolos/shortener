package model

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
)

type URLWithFile struct {
	URL             *URL
	fileStoragePath string
}

func NewURLWithFile(fileStoragePath string) (*URLWithFile, error) {
	url, err := NewURL()
	if err != nil {
		return nil, errors.New("не удалось создать urls хранилище")
	}
	urlWithFile := &URLWithFile{URL: url, fileStoragePath: fileStoragePath}
	err = urlWithFile.load()
	if err != nil {
		return nil, errors.New("не удалось загрузить сохраненные urls сокращения")
	}
	return urlWithFile, nil
}

func (uwf *URLWithFile) Set(key, value string) error {
	err := uwf.URL.Set(key, value)
	if err != nil {
		return err
	}
	err = uwf.save(key, value)
	if err != nil {
		err := uwf.URL.delete(key)
		if err != nil {
			return errors.New("критичное расхождение " + key + ": " + value)
		}
		return errors.New("не удалось записать новой сокращенный url в файл")
	}
	return nil
}

func (uwf *URLWithFile) Get(key string) (string, error) {
	return uwf.URL.Get(key)
}

func (uwf *URLWithFile) load() error {
	data, err := os.ReadFile(uwf.fileStoragePath)
	if errors.Is(err, os.ErrNotExist) {
		return uwf.createEmpty()
	} else if err != nil {
		return err
	}

	jsonData := &[]map[string]string{}
	if err = json.Unmarshal(data, jsonData); err != nil {
		return err
	}
	for _, el := range *jsonData {
		err = uwf.URL.Set(el["short_url"], el["original_url"])
		if err != nil {
			return err
		}
	}
	return nil
}

func (uwf *URLWithFile) createEmpty() error {
	file, err := os.OpenFile(uwf.fileStoragePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := json.Marshal([]string{})
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

type addingSlice struct {
	key   string
	value string
	count int
}

// необходим для добавления новой строки в красивом формате
func (s addingSlice) marshalJSON() ([]byte, error) {
	data, err := json.Marshal(map[string]string{"uuid": strconv.Itoa(s.count), "short_url": s.key, "original_url": s.value})
	if err != nil {
		return nil, err
	}
	start := []byte("\n\t")
	if s.count != 1 {
		start = append([]byte(","), start...)
	}
	withMarshal := append(start, data...)
	withEndFile := append(withMarshal, []byte("\n]")...)
	return withEndFile, nil
}

func (uwf *URLWithFile) save(key, value string) error {
	file, err := os.OpenFile(uwf.fileStoragePath, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	urlsCount := len(uwf.URL.data)
	data, err := addingSlice{key: key, value: value, count: urlsCount}.marshalJSON()
	if err != nil {
		return err
	}

	seekNum := int64(-1)
	if urlsCount != 1 {
		seekNum -= 1
	}
	_, err = file.Seek(seekNum, io.SeekEnd)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
