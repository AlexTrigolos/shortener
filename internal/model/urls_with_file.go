package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
)

type URLWithFile struct {
	URL               *URL
	fileStoragePath   string
	transactionBuffer map[string]string
	transactionUserID string
}

func NewURLWithFile(fileStoragePath string) (*URLWithFile, error) {
	url, err := NewURL()
	if err != nil {
		return nil, errors.New("не удалось создать urls хранилище")
	}
	urlWithFile := &URLWithFile{URL: url, fileStoragePath: fileStoragePath}
	err = urlWithFile.load(context.Background()) // контекст не важен, ведь доступа к бд нету
	if err != nil {
		return nil, errors.New("не удалось загрузить сохраненные urls сокращения")
	}
	return urlWithFile, nil
}

func (uwf *URLWithFile) Set(ctx context.Context, key, value, userID string) error {
	err := uwf.URL.Set(ctx, key, value, userID)
	if err != nil {
		return err
	}
	err = uwf.save(key, value, userID)
	if err != nil {
		uwf.URL.delete(key, userID)
		return errors.New("не удалось записать новой сокращенный url в файл")
	}
	return nil
}

func (uwf *URLWithFile) Get(ctx context.Context, key string) (string, error) {
	return uwf.URL.Get(ctx, key)
}

func (uwf *URLWithFile) GetByOriginal(ctx context.Context, original, baseURL string) (string, error) {
	return uwf.URL.GetByOriginal(ctx, original, baseURL)
}

func (uwf *URLWithFile) GetByUserID(ctx context.Context, userID, baseURL string) (*[]UserURL, error) {
	return uwf.URL.GetByUserID(ctx, userID, baseURL)
}

func (uwf *URLWithFile) load(ctx context.Context) error {
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
		err = uwf.URL.Set(ctx, el["short_url"], el["original_url"], el["user_id"])
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
	key     string
	value   string
	count   int
	userID string
}

// необходим для добавления новой строки в красивом формате
func (s addingSlice) marshalJSON() ([]byte, error) {
	data, err := json.Marshal(map[string]string{"uuid": strconv.Itoa(s.count), "short_url": s.key, "original_url": s.value, "user_id": s.userID})
	if err != nil {
		return nil, err
	}
	start := []byte("\n\t")
	if s.count != 1 {
		start = append([]byte(","), start...)
	}
	withMarshal := append(start, data...)
	return withMarshal, nil
}

func (uwf *URLWithFile) save(key, value, userID string) error {
	file, err := os.OpenFile(uwf.fileStoragePath, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	urlsCount := len(uwf.URL.data)
	data, err := addingSlice{key: key, value: value, count: urlsCount, userID: userID}.marshalJSON()
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

	_, err = file.Write(append(data, []byte("\n]")...))
	if err != nil {
		return err
	}

	return nil
}

func (uwf *URLWithFile) PingContext(ctx context.Context) error {
	return nil
}

// имитируем транзакцию
func (uwf *URLWithFile) Begin(count int) (*Transaction, error) {
	uwf.URL.mu.Lock()
	uwf.transactionBuffer = make(map[string]string, count)
	transaction, err := NewTransaction(uwf, nil, uwf)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (uwf *URLWithFile) Rollback() error {
	defer uwf.URL.mu.Unlock()
	uwf.transactionBuffer = nil
	uwf.transactionUserID = ""
	return nil
}

func (uwf *URLWithFile) Commit() error {
	defer uwf.URL.mu.Unlock()
	defer func() { uwf.transactionBuffer = nil; uwf.transactionUserID = "" }()

	userID := uwf.transactionUserID
	savedKeys := make([]string, 0, len(uwf.transactionBuffer))
	buffedFileBytes := make([][]byte, 0, len(uwf.transactionBuffer))
	for k, v := range uwf.transactionBuffer {
		_, exists := uwf.URL.data[k]
		if exists {
			return uwf.rollbackURLCommit(savedKeys, userID, k+" ключ уже занят")
		}
		_, exists = uwf.URL.values[v]
		if exists {
			return uwf.rollbackURLCommit(savedKeys, userID, v+" значение уже занято")
		}

		uwf.URL.values[v] = k
		uwf.URL.data[k] = v
		uwf.URL.byUserIds[userID] = append(uwf.URL.byUserIds[userID], UserURL{ShortURL: k, OriginalURL: v})
		savedKeys = append(savedKeys, k)

		buff, err := uwf.buffCommit(k, v, userID)
		if err != nil {
			return uwf.rollbackURLCommit(savedKeys, userID, "проблема с генерацией строки для сохранения: "+err.Error())
		}
		buffedFileBytes = append(buffedFileBytes, buff)
	}
	if len(uwf.transactionBuffer) > 0 {
		err := uwf.saveCommit(buffedFileBytes)
		if err != nil {
			return uwf.rollbackURLCommit(savedKeys, userID, "проблема с сохранением файла транзакции: "+err.Error())
		}
	}
	return nil
}

func (uwf *URLWithFile) rollbackURLCommit(savedKeys []string, userID string, errStr string) error {
	for _, delKey := range savedKeys {
		uwf.URL.remove(delKey, userID)
	}
	return errors.New(errStr)
}

func (uwf *URLWithFile) buffCommit(key, value, userID string) ([]byte, error) {
	urlsCount := len(uwf.URL.data)
	data, err := addingSlice{key: key, value: value, count: urlsCount, userID: userID}.marshalJSON()
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (uwf *URLWithFile) saveCommit(saveBytes [][]byte) error {
	file, err := os.OpenFile(uwf.fileStoragePath, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	data := bytes.Join(saveBytes, []byte(""))

	seekNum := int64(-1)
	if len(uwf.URL.data) != len(saveBytes) {
		seekNum -= 1
	}
	_, err = file.Seek(seekNum, io.SeekEnd)
	if err != nil {
		return err
	}

	_, err = file.Write(append(data, []byte("\n]")...))
	if err != nil {
		return err
	}

	return nil
}

func (uwf *URLWithFile) SetTx(ctx context.Context, key, value, userID string) error {
	_, exists := uwf.URL.data[key]
	if exists {
		return errors.New(key + " ключ уже занят")
	}
	_, exists = uwf.transactionBuffer[key]
	if exists {
		return errors.New(key + " ключ уже занят в транзакции")
	}
	uwf.transactionBuffer[key] = value
	uwf.transactionUserID = userID
	return nil
}
