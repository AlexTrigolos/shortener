package model

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	uUserID  = "user_id"
	uBaseURL = "base_url"
)

func TestNewURL(t *testing.T) {
	t.Run("удачное получение", func(t *testing.T) {
		newURL, err := NewURL()
		require.NoError(t, err)
		assert.Equal(t, &URL{
			mu:                sync.Mutex{},
			data:              map[string]string{},
			values:            map[string]string{},
			transactionBuffer: nil,
			byUserIds:         map[string][]UserURL{},
			deleteFlag: 	   map[string]struct{}{},
			transactionUserID: "",
		}, newURL)
	})
}

func TestURL_Set(t *testing.T) {
	urls := &URL{
		data:      map[string]string{"key1": "value1"},
		values:    map[string]string{"value1": "key1"},
		byUserIds: map[string][]UserURL{uUserID: {{ShortURL: "key1", OriginalURL: "value1"}}},
	}
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr string
	}{
		{
			name:    "добавление нового ключа",
			key:     "key0",
			value:   "value0",
			wantErr: "",
		},
		{
			name:    "установка по имеющемуся ключу",
			key:     "key1",
			value:   "value2",
			wantErr: "такой ключ уже занят",
		},
		{
			name:    "конфликт значений",
			key:     "key1",
			value:   "value1",
			wantErr: "conflict",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := urls.Set(t.Context(), tt.key, tt.value, uUserID)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Contains(t, urls.data, tt.key)
				assert.Equal(t, tt.value, urls.data[tt.key])
				assert.Contains(t, urls.values, tt.value)
				assert.Equal(t, tt.key, urls.values[tt.value])
				assert.Contains(t, urls.byUserIds, uUserID)
				assert.Equal(t, []UserURL{{ShortURL: "key1", OriginalURL: "value1"}, {ShortURL: tt.key, OriginalURL: tt.value}}, urls.byUserIds[uUserID])
			}
		})
	}
}

func TestURL_Get(t *testing.T) {
	urls := &URL{data: map[string]string{"key1": "value1"}, deleteFlag: map[string]struct{}{"key_del": {}}}
	tests := []struct {
		name    string
		key     string
		wantVal string
		wantErr string
	}{
		{
			name:    "получает значение по ключу",
			key:     "key1",
			wantVal: "value1",
			wantErr: "",
		},
		{
			name:    "не находит значение по ключу",
			key:     "key0",
			wantVal: "",
			wantErr: "не найден ключ key0",
		},
		{
			name:    "ключ удален",
			key:     "key_del",
			wantVal: "",
			wantErr: "deleted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := urls.Get(t.Context(), tt.key)
			assert.Equal(t, tt.wantVal, value)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestURL_GetByOriginal(t *testing.T) {
	urls := &URL{values: map[string]string{"value1": "key1"}}
	tests := []struct {
		name    string
		value   string
		wantKey string
		wantErr string
	}{
		{
			name:    "получает ключа по значению",
			value:   "value1",
			wantKey: uBaseURL + "/key1",
			wantErr: "",
		},
		{
			name:    "не находит ключ по значению",
			value:   "value0",
			wantKey: "",
			wantErr: "не нашли значение value0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := urls.GetByOriginal(t.Context(), tt.value, uBaseURL)
			assert.Equal(t, tt.wantKey, key)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestURL_GetByUserID(t *testing.T) {
	urls := &URL{byUserIds: map[string][]UserURL{uUserID: {{ShortURL: "key1", OriginalURL: "value1"}}}}
	tests := []struct {
		name     string
		userID  string
		wantURLs []UserURL
	}{
		{
			name:     "получает ссылок по пользователю",
			userID:  uUserID,
			wantURLs: []UserURL{{ShortURL: uBaseURL + "/key1", OriginalURL: "value1"}},
		},
		{
			name:     "не находит ключ по значению",
			userID:  "unknown_user_id",
			wantURLs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userURls, err := urls.GetByUserID(t.Context(), tt.userID, uBaseURL)
			require.NoError(t, err)
			assert.Equal(t, tt.wantURLs, *userURls)
		})
	}
}

func Test_delete(t *testing.T) {
	urls := &URL{
		data:      map[string]string{"key0": "value0", "key1": "value1", "key2": "value2"},
		values:    map[string]string{"value0": "key0", "value1": "key1", "value2": "key2"},
		byUserIds: map[string][]UserURL{uUserID: {{ShortURL: "key0", OriginalURL: "value0"}, {ShortURL: "key1", OriginalURL: "value1"}, {ShortURL: "key2", OriginalURL: "value2"}}},
	}
	tests := []struct {
		name string
		key  string
	}{
		{
			name: "удачное удаление имеющегося",
			key:  "key1",
		},
		{
			name: "удачное удаление не имеющегося",
			key:  "key999",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := urls.data[tt.key]
			urls.delete(tt.key, uUserID)
			assert.NotContains(t, urls.data, tt.key)
			assert.NotContains(t, urls.values, value)
			assert.Equal(t, []UserURL{{ShortURL: "key0", OriginalURL: "value0"}, {ShortURL: "key2", OriginalURL: "value2"}}, urls.byUserIds[uUserID])
		})
	}
}

func TestURL_PingContext(t *testing.T) {
	t.Run("нет доступа к бд", func(t *testing.T) {
		u, err := NewURL()
		require.NoError(t, err)
		err = u.PingContext(t.Context())
		require.NoError(t, err)
	})
}

func TestURL_Begin(t *testing.T) {
	t.Run("удачное создание транзакции", func(t *testing.T) {
		urls, err := NewURL()
		require.NoError(t, err)

		tr, err := urls.Begin(5)
		require.NoError(t, err)
		assert.Equal(t, &Transaction{tx: urls, selfSetTx: urls, commited: false}, tr)
		assert.Equal(t, make(map[string]string, 5), urls.transactionBuffer)
		assert.Empty(t, urls.transactionUserID)
		assert.False(t, urls.mu.TryLock())
	})
}

func TestURL_Rollback(t *testing.T) {
	t.Run("удачная отмена транзакции", func(t *testing.T) {
		urls, err := NewURL()
		require.NoError(t, err)

		_, err = urls.Begin(5)
		require.NoError(t, err)

		err = urls.Rollback()
		require.NoError(t, err)
		assert.Nil(t, urls.transactionBuffer)
		assert.Empty(t, urls.transactionUserID)
		assert.True(t, urls.mu.TryLock())
	})
}

func TestURL_Commit(t *testing.T) {
	t.Run("удачное завершение транзакции", func(t *testing.T) {
		urls, err := NewURL()
		require.NoError(t, err)

		_, err = urls.Begin(5,)
		require.NoError(t, err)

		urls.transactionUserID = uUserID
		urls.transactionBuffer["short_1"] = "long_1"
		urls.transactionBuffer["short_2"] = "long_2"

		err = urls.Commit()
		require.NoError(t, err)
		assert.Nil(t, urls.transactionBuffer)
		assert.Empty(t, urls.transactionUserID)
		assert.True(t, urls.mu.TryLock())

		assert.Len(t, urls.data, 2)
		assert.Contains(t, urls.data, "short_1")
		assert.Equal(t, urls.data["short_1"], "long_1")
		assert.Contains(t, urls.data, "short_2")
		assert.Equal(t, urls.data["short_2"], "long_2")

		assert.Len(t, urls.values, 2)
		assert.Contains(t, urls.values, "long_1")
		assert.Equal(t, urls.values["long_1"], "short_1")
		assert.Contains(t, urls.values, "long_2")
		assert.Equal(t, urls.values["long_2"], "short_2")

		assert.Len(t, urls.byUserIds, 1)
		assert.Contains(t, urls.byUserIds, uUserID)
		assert.Len(t, urls.byUserIds[uUserID], 2)
		assert.Contains(t, urls.byUserIds[uUserID], UserURL{ShortURL: "short_1", OriginalURL: "long_1"})
		assert.Contains(t, urls.byUserIds[uUserID], UserURL{ShortURL: "short_2", OriginalURL: "long_2"})
	})
	t.Run("совпадение при завершении транзакции", func(t *testing.T) {
		urls := &URL{
			data:      map[string]string{"short_2": "long_2"},
			values:    map[string]string{"long_2": "short_2"},
			byUserIds: map[string][]UserURL{uUserID: {{ShortURL: "short_2", OriginalURL: "long_2"}}},
		}

		_, err := urls.Begin(5)
		require.NoError(t, err)

		urls.transactionUserID = uUserID
		urls.transactionBuffer["short_1"] = "long_1"
		urls.transactionBuffer["short_2"] = "long_2"

		err = urls.Commit()
		require.ErrorContains(t, err, "short_2 ключ уже занят")
		assert.Nil(t, urls.transactionBuffer)
		assert.Empty(t, urls.transactionUserID)
		assert.True(t, urls.mu.TryLock())

		assert.Len(t, urls.data, 1)
		assert.Contains(t, urls.data, "short_2")
		assert.Equal(t, urls.data["short_2"], "long_2")

		assert.Len(t, urls.values, 1)
		assert.Contains(t, urls.values, "long_2")
		assert.Equal(t, urls.values["long_2"], "short_2")

		assert.Len(t, urls.byUserIds, 1)
		assert.Contains(t, urls.byUserIds, uUserID)
		assert.Len(t, urls.byUserIds[uUserID], 1)
		assert.Contains(t, urls.byUserIds, uUserID)
		assert.Contains(t, urls.byUserIds[uUserID], UserURL{ShortURL: "short_2", OriginalURL: "long_2"})
	})
}

func TestURL_SetTx(t *testing.T) {
	t.Run("удачное добавление внутри транзакции", func(t *testing.T) {
		urls, err := NewURL()
		require.NoError(t, err)

		_, err = urls.Begin(5)
		require.NoError(t, err)

		err = urls.SetTx(t.Context(), "key_1", "value_1", uUserID)
		require.NoError(t, err)
		err = urls.SetTx(t.Context(), "key_2", "value_2", uUserID)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"key_1": "value_1", "key_2": "value_2"}, urls.transactionBuffer)
		assert.Equal(t, uUserID, urls.transactionUserID)
		assert.False(t, urls.mu.TryLock())

		assert.Len(t, urls.data, 0)

		assert.Len(t, urls.transactionBuffer, 2)
		assert.Contains(t, urls.transactionBuffer, "key_1")
		assert.Equal(t, urls.transactionBuffer["key_1"], "value_1")
		assert.Contains(t, urls.transactionBuffer, "key_2")
		assert.Equal(t, urls.transactionBuffer["key_2"], "value_2")
	})
	t.Run("неудачное добавление внутри транзакции", func(t *testing.T) {
		urls, err := NewURL()
		require.NoError(t, err)
		err = urls.Set(t.Context(), "key_1", "value_1", uUserID)
		require.NoError(t, err)

		_, err = urls.Begin(5)
		require.NoError(t, err)

		err = urls.SetTx(t.Context(), "key_1", "value_1", uUserID)
		require.ErrorContains(t, err, "key_1 ключ уже занят")
		assert.Equal(t, make(map[string]string, 5), urls.transactionBuffer)
		assert.Empty(t, urls.transactionUserID)
		assert.False(t, urls.mu.TryLock())

		assert.Len(t, urls.data, 1)
		assert.Len(t, urls.transactionBuffer, 0)
	})
	t.Run("неудачное добавление внутри транзакции из-за совпадения с временным хранилищем", func(t *testing.T) {
		urls, err := NewURL()
		require.NoError(t, err)

		_, err = urls.Begin(5)
		require.NoError(t, err)

		err = urls.SetTx(t.Context(), "key_1", "value_1", uUserID)
		require.NoError(t, err)
		err = urls.SetTx(t.Context(), "key_1", "value_1", uUserID)
		require.ErrorContains(t, err, "key_1 ключ уже занят")
		assert.Equal(t, map[string]string{"key_1": "value_1"}, urls.transactionBuffer)
		assert.Equal(t, uUserID, urls.transactionUserID)
		assert.False(t, urls.mu.TryLock())

		assert.Len(t, urls.data, 0)
		assert.Len(t, urls.transactionBuffer, 1)
	})
}
