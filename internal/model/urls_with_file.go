package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
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
		return uwf.createEmpty(uwf.fileStoragePath)
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

	dir := filepath.Dir(uwf.fileStoragePath)
	ext := filepath.Ext(uwf.fileStoragePath)
	removedPath := filepath.Join(dir, "removed"+ext)
	data, err = os.ReadFile(removedPath)
	if errors.Is(err, os.ErrNotExist) {
		return uwf.createEmpty(removedPath)
	} else if err != nil {
		return err
	}

	jsonRemovedData := &[]string{}
	if err = json.Unmarshal(data, jsonRemovedData); err != nil {
		return err
	}
	for _, el := range *jsonRemovedData {
		uwf.URL.deleteFlag[el] = struct{}{}
	}

	return nil
}

func (uwf *URLWithFile) createEmpty(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
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
	key    string
	value  string
	count  int
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

func (uwf *URLWithFile) CheckUserIDURLs(ctx context.Context, urls []string, userID string) error {
	return uwf.URL.CheckUserIDURLs(ctx, urls, userID)
}

func (uwf *URLWithFile) DeleteURLs(ctx context.Context, urlsData []URLsData) error {
	uwf.URL.mu.Lock()
	data := uwf.URL.data
	values := uwf.URL.values
	byUserIds := uwf.URL.byUserIds
	deleteFlag := uwf.URL.deleteFlag
	uwf.URL.mu.Unlock()

	err := uwf.URL.DeleteURLs(ctx, urlsData)
	if err != nil {
		return err
	}

	backupPath, err := uwf.tempFile()
	defer os.Remove(backupPath)
	if err != nil {
		uwf.URL.data = data
		uwf.URL.values = values
		uwf.URL.byUserIds = byUserIds
		uwf.URL.deleteFlag = deleteFlag
		return err
	}

	if err := uwf.applyEdits(); err != nil {
		uwf.URL.data = data
		uwf.URL.values = values
		uwf.URL.byUserIds = byUserIds
		uwf.URL.deleteFlag = deleteFlag

		if renameErr := os.Rename(backupPath, uwf.fileStoragePath); renameErr != nil {
			log.Printf("откат не атомарен: %v; копия: %s", renameErr, backupPath)
		}
		return err
	}
	
	err = uwf.saveRemoved(uwf.URL.deleteFlag)
	if err != nil {
		uwf.URL.data = data
		uwf.URL.values = values
		uwf.URL.byUserIds = byUserIds
		uwf.URL.deleteFlag = deleteFlag

		if renameErr := os.Rename(backupPath, uwf.fileStoragePath); renameErr != nil {
			log.Printf("откат не атомарен: %v; копия: %s", renameErr, backupPath)
		}
		return err
	}

	return nil
}

func (uwf *URLWithFile) tempFile() (string, error) {
	dir := filepath.Dir(uwf.fileStoragePath)
	backup, err := os.CreateTemp(dir, filepath.Base(uwf.fileStoragePath)+".bak_*")
	if err != nil {
		return "", err
	}
	backupPath := backup.Name()

	src, err := os.Open(uwf.fileStoragePath)
	if err != nil {
		backup.Close()
		os.Remove(backupPath)
		return "", err
	}
	if _, err := io.Copy(backup, src); err != nil {
		src.Close()
		os.Remove(backupPath)
		return "", err
	}
	src.Close()
	backup.Close()

	return backupPath, nil
}

func (uwf *URLWithFile) applyEdits() error {
	data, err := os.ReadFile(uwf.fileStoragePath)
	if err != nil {
		return err
	}

	jsonData := &[]map[string]string{}
	if err = json.Unmarshal(data, jsonData); err != nil {
		return err
	}

	valid := []map[string]string{}
	for _, el := range *jsonData {
		if _, ex := uwf.URL.deleteFlag[el["short_url"]]; !ex {
			valid = append(valid, el)
		}
	}

	file, err := os.OpenFile(uwf.fileStoragePath, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	resData := []byte{}
	for i, v := range valid {
		data, err := addingSlice{key: v["short_url"], value: v["original_url"], count: i + 1, userID: v["user_id"]}.marshalJSON()
		if err != nil {
			return err
		}
		resData = append(resData, data...)
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = file.Write(
		append(
			append([]byte("[\n"), resData...),
			[]byte("\n]")...
		),
	)
	if err != nil {
		return err
	}

	return nil
}

func (uwf *URLWithFile) saveRemoved(removed map[string]struct{}) error {
	dir := filepath.Dir(uwf.fileStoragePath)
	ext := filepath.Ext(uwf.fileStoragePath)
	file, err := os.OpenFile(filepath.Join(dir, "removed"+ext), os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	removedArray := []string{}
	for removedKey := range removed {
		removedArray = append(removedArray, removedKey)
	}

	data, err := json.Marshal(removedArray)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (uwf *URLWithFile) PingContext(_ context.Context) error {
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
