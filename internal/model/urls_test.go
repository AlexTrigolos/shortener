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
			name: "создается url",
			want: &URL{mu: sync.RWMutex{}, data: make(map[string]string)},
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

func TestURL_IsExists(t *testing.T) {
	tests := []struct {
		name string
		key  string
		urls *URL
		want bool
	}{
		{
			name: "нет такого ключа",
			key:  "key1",
			urls: &URL{data: map[string]string{}},
			want: false,
		},
		{
			name: "имеется такой ключ",
			key:  "key1",
			urls: &URL{data: map[string]string{"key1": "value1"}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := tt.urls
			assert.Equal(t, tt.want, urls.IsExists(tt.key))
		})
	}
}

func TestURL_Set(t *testing.T) {
	tests := []struct {
		name    string
		urls    *URL
		key     string
		value   string
		wantErr string
	}{
		{
			name:    "добавление нового ключа",
			urls:    &URL{data: map[string]string{}},
			key:     "key1",
			value:   "value1",
			wantErr: "",
		},
		{
			name:    "установка по имеющемуся ключу",
			urls:    &URL{data: map[string]string{"key1": "value0"}},
			key:     "key1",
			value:   "value1",
			wantErr: "такой ключ уже занят",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := tt.urls
			err := urls.Set(tt.key, tt.value)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Contains(t, urls.data, tt.key)
			assert.Equal(t, urls.data[tt.key], tt.value)
		})
	}
}

func TestURL_Get(t *testing.T) {
	tests := []struct {
		name    string
		urls    *URL
		key     string
		wantVal string
		wantErr string
	}{
		{
			name:    "получает значение по ключу",
			urls:    &URL{data: map[string]string{"key1": "value1"}},
			key:     "key1",
			wantVal: "value1",
			wantErr: "",
		},
		{
			name:    "не находит значение по ключу",
			urls:    &URL{data: map[string]string{}},
			key:     "key1",
			wantVal: "",
			wantErr: "такой ключ не найден",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := tt.urls
			value, err := urls.Get(tt.key)
			assert.Equal(t, tt.wantVal, value)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
