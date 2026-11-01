package model

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURL(t *testing.T) {
	tests := []struct {
		name string
		want *URL
	}{
		{
			name: "удачное получение",
			want: &URL{
				mu:   sync.Mutex{},
				data: map[string]string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newURL, err := NewURL()
			require.NoError(t, err)
			assert.Equal(t, tt.want, newURL)
		})
	}
}

func TestURL_Set(t *testing.T) {
	urls := &URL{data: map[string]string{"key1": "value1"}}
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
			value:   "value1",
			wantErr: "такой ключ уже занят",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := urls.Set(tt.key, tt.value)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Contains(t, urls.data, tt.key)
				assert.Equal(t, urls.data[tt.key], tt.value)
			}
		})
	}
}

func TestURL_Get(t *testing.T) {
	urls := &URL{data: map[string]string{"key1": "value1"}}
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
			wantErr: "такой ключ не найден",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := urls.Get(tt.key)
			assert.Equal(t, tt.wantVal, value)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_delete(t *testing.T) {
	urls := &URL{data: map[string]string{"key1": "value1"}}
	tests := []struct {
		name string
		key string
	}{
		{
			name: "удачное удаление имеющегося",
			key: "key1",
		},
		{
			name: "удачное удаление не имеющегося",
			key: "key0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := urls.delete(tt.key)
			require.NoError(t, err)
			assert.NotContains(t, urls.data, tt.key)
		})
	}
}