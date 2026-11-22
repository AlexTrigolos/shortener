package model

import (
	"context"
	"errors"
	"sync"
)

type URL struct {
	mu                sync.Mutex
	data              map[string]string
	values            map[string]string
	byUserIds         map[string][]UserURL
	transactionBuffer map[string]string
	transactionUserID string
}

func NewURL() (*URL, error) {
	return &URL{
		data:      make(map[string]string),
		values:    make(map[string]string),
		byUserIds: make(map[string][]UserURL),
	}, nil
}

func (u *URL) Set(_ context.Context, key, value, userID string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	_, exists := u.values[value]
	if exists {
		return ErrConflict
	}
	_, exists = u.data[key]
	if exists {
		return errors.New("такой ключ уже занят")
	}
	u.values[value] = key
	u.data[key] = value
	u.byUserIds[userID] = append(u.byUserIds[userID], UserURL{ShortURL: key, OriginalURL: value})
	return nil
}

func (u *URL) Get(_ context.Context, key string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	value, exists := u.data[key]
	if !exists {
		return "", errors.New("не найден ключ " + key)
	}
	return value, nil
}

func (u *URL) GetByOriginal(_ context.Context, original, baseURL string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	short, exists := u.values[original]
	if !exists {
		return "", errors.New("не нашли значение " + original)
	}
	return baseURL + "/" + short, nil
}

func (u *URL) GetByUserID(_ context.Context, userID, baseURL string) (*[]UserURL, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	userURLs := u.byUserIds[userID]

	for i := range userURLs {
		userURLs[i].ShortURL = baseURL + "/" + userURLs[i].ShortURL
	}

	return &userURLs, nil
}

func (u *URL) delete(key string, userID string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.remove(key, userID)
}

func (u *URL) remove(key string, userID string) {
	for i, v := range u.byUserIds[userID] {
		if v.ShortURL == key {
			u.byUserIds[userID] = append(u.byUserIds[userID][:i], u.byUserIds[userID][i+1:]...)
			break
		}
	}
	delete(u.values, u.data[key])
	delete(u.data, key)
}

func (u *URL) PingContext(ctx context.Context) error {
	return nil
}

// имитируем транзакцию
func (u *URL) Begin(count int) (*Transaction, error) {
	u.mu.Lock()
	u.transactionBuffer = make(map[string]string, count)
	transaction, err := NewTransaction(u, nil, u)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (u *URL) Rollback() error {
	defer u.mu.Unlock()
	u.transactionBuffer = nil
	u.transactionUserID = ""
	return nil
}

func (u *URL) Commit() error {
	defer u.mu.Unlock()
	defer func() { u.transactionBuffer = nil; u.transactionUserID = "" }()

	userID := u.transactionUserID
	savedKeys := []string{}
	for k, v := range u.transactionBuffer {
		_, exists := u.data[k]
		if exists {
			for _, delKey := range savedKeys {
				u.remove(delKey, userID)
			}
			return errors.New(k + " ключ уже занят")
		}
		_, exists = u.values[v]
		if exists {
			for _, delKey := range savedKeys {
				u.remove(delKey, userID)
			}
			return errors.New(k + " значение уже занято")
		}
		u.values[v] = k
		u.data[k] = v
		u.byUserIds[userID] = append(u.byUserIds[userID], UserURL{ShortURL: k, OriginalURL: v})
		savedKeys = append(savedKeys, k)
	}
	return nil
}

func (u *URL) SetTx(_ context.Context, key, value, userID string) error {
	_, exists := u.data[key]
	if exists {
		return errors.New(key + " ключ уже занят")
	}
	_, exists = u.transactionBuffer[key]
	if exists {
		return errors.New(key + " ключ уже занят в транзакции")
	}
	u.transactionBuffer[key] = value
	u.transactionUserID = userID
	return nil
}
