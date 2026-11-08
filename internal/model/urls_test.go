package model

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURL(t *testing.T) {
	t.Run("удачное получение", func(t *testing.T) {
		newURL, err := NewURL()
		require.NoError(t, err)
		assert.Equal(t, &URL{mu: sync.Mutex{}, data: map[string]string{}, values: map[string]string{}, transactionBuffer: nil}, newURL)
	})
}

func TestURL_Set(t *testing.T) {
	urls := &URL{data: map[string]string{"key1": "value1"}, values: map[string]string{"value1": "key1"}}
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
			err := urls.Set(t.Context(), tt.key, tt.value)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Contains(t, urls.data, tt.key)
				assert.Equal(t, urls.data[tt.key], tt.value)
				assert.Contains(t, urls.values, tt.value)
				assert.Equal(t, urls.values[tt.value], tt.key)
			}
		})
	}
}

func TestURL_Get(t *testing.T) {
	urls := &URL{data: map[string]string{"key1": "value1"}, values: map[string]string{"value1": "key1"}}
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
	urls := &URL{data: map[string]string{"key1": "value1"}, values: map[string]string{"value1": "key1"}}
	tests := []struct {
		name    string
		value   string
		wantKey string
		wantErr string
	}{
		{
			name:    "получает ключа по значению",
			value:   "value1",
			wantKey: "key1",
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
			key, err := urls.GetByOriginal(t.Context(), tt.value)
			assert.Equal(t, tt.wantKey, key)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_delete(t *testing.T) {
	urls := &URL{data: map[string]string{"key1": "value1"}, values: map[string]string{"value1": "key1"}}
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
			key:  "key0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := urls.data[tt.key]
			err := urls.delete(tt.key)
			require.NoError(t, err)
			assert.NotContains(t, urls.data, tt.key)
			assert.NotContains(t, urls.values, value)
		})
	}
}

func TestURL_PingContext(t *testing.T) {
	t.Run("нет доступа к бд", func(t *testing.T) {
		u, err := NewURL()
		require.NoError(t, err)
		gotErr := u.PingContext(t.Context())
		require.ErrorContains(t, gotErr, "нет доступа к бд")
	})
}

func TestURL_Begin(t *testing.T) {
	t.Run("удачное создание транзакции", func(t *testing.T) {
		urls, err := NewURL()
		require.NoError(t, err)

		tr, err := urls.Begin(5)
		require.NoError(t, err)
		assert.Equal(t, &Transaction{tx: urls, commited: false}, tr)
		assert.NotNil(t, urls.transactionBuffer)
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
		assert.True(t, urls.mu.TryLock())
	})
}

func TestURL_Commit(t *testing.T) {
	t.Run("удачное завершение транзакции", func(t *testing.T) {
		urls, err := NewURL()
		require.NoError(t, err)

		_, err = urls.Begin(5)
		require.NoError(t, err)

		urls.transactionBuffer["short_1"] = "long_1"
		urls.transactionBuffer["short_2"] = "long_2"

		err = urls.Commit()
		require.NoError(t, err)
		assert.Nil(t, urls.transactionBuffer)
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
	})
}

func TestURL_SetTx(t *testing.T) {
	t.Run("удачное добавление внутри транзакции", func(t *testing.T) {
		urls, err := NewURL()
		require.NoError(t, err)

		_, err = urls.Begin(5)
		require.NoError(t, err)

		err = urls.SetTx(t.Context(), "key_1", "value_1")
		require.NoError(t, err)
		err = urls.SetTx(t.Context(), "key_2", "value_2")
		require.NoError(t, err)
		assert.NotNil(t, urls.transactionBuffer)
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
		err = urls.Set(t.Context(), "key_1", "value_1")
		require.NoError(t, err)

		_, err = urls.Begin(5)
		require.NoError(t, err)

		err = urls.SetTx(t.Context(), "key_1", "value_1")
		require.ErrorContains(t, err, "key_1 ключ уже занят")
		assert.NotNil(t, urls.transactionBuffer)
		assert.False(t, urls.mu.TryLock())

		assert.Len(t, urls.data, 1)
		assert.Len(t, urls.transactionBuffer, 0)
	})
}
