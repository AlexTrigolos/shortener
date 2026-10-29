package model

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURL(t *testing.T) {
	dirname := t.TempDir()
	defer os.RemoveAll(dirname)
	file, err := os.CreateTemp(dirname, "*_file.json")
	require.NoError(t, err)

	fileData, err := json.MarshalIndent([]map[string]string{{"short_url": "4rSPg8ap", "original_url": "http://yandex.ru"}}, "", "	")
	require.NoError(t, err)
	_, err = file.Write(fileData)
	require.NoError(t, err)

	filename := file.Name()

	emptyData, err := json.Marshal([]string{})
	require.NoError(t, err)

	tests := []struct {
		name         string
		filename     string
		wantFileData []byte
		want         *URL
	}{
		{
			name:         "есть файл с сохранениями",
			filename:     filename,
			wantFileData: fileData,
			want: &URL{
				mu:              sync.Mutex{},
				data:            map[string]string{"4rSPg8ap": "http://yandex.ru"},
				fileStoragePath: filename,
				logger:          slog.New(slog.NewTextHandler(os.Stdout, nil)),
			},
		},
		{
			name:         "не было сохранений",
			filename:     "./new_file.json",
			wantFileData: emptyData,
			want: &URL{
				mu:              sync.Mutex{},
				data:            map[string]string{},
				fileStoragePath: "./new_file.json",
				logger:          slog.New(slog.NewTextHandler(os.Stdout, nil))},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newURL, err := NewURL(tt.filename)
			require.NoError(t, err)
			assert.Equal(t, tt.want, newURL)

			data, err := os.ReadFile(tt.filename)
			require.NoError(t, err)
			assert.JSONEq(t, string(tt.wantFileData), string(data))

			err = os.Remove(tt.filename)
			require.NoError(t, err)
		})
	}
}

func TestURL_Set(t *testing.T) {
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
		{
			name:    "добавление нового ключа в пустой",
			key:     "key0",
			value:   "value0",
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirname := t.TempDir()
			defer os.RemoveAll(dirname)
			file, err := os.CreateTemp(dirname, "*_file.json")
			require.NoError(t, err)
			filename := file.Name()

			wantFileData, err := json.Marshal([]string{})
			require.NoError(t, err)

			firstURL := true
			urls := &URL{data: map[string]string{}, fileStoragePath: filename}
			if tt.name != "добавление нового ключа в пустой" {
				urls.data["key1"] = "value1"

				newData, err := addingSlice{key: "key1", value: "value1", first: firstURL}.marshalJSON()
				require.NoError(t, err)
				firstURL = false
				wantFileData = append(wantFileData[:len(wantFileData)-1], newData...)
			}

			_, err = file.Write(wantFileData)
			require.NoError(t, err)

			if tt.wantErr == "" {
				newData, err := addingSlice{key: "key0", value: "value0", first: firstURL}.marshalJSON()
				require.NoError(t, err)
				if firstURL {
					wantFileData = append(wantFileData[:len(wantFileData)-1], newData...)
				} else {
					wantFileData = append(wantFileData[:len(wantFileData)-2], newData...)
				}
			}

			err = urls.Set(tt.key, tt.value)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Contains(t, urls.data, tt.key)
				assert.Equal(t, urls.data[tt.key], tt.value)
			}

			data, err := os.ReadFile(filename)
			require.NoError(t, err)
			assert.JSONEq(t, string(wantFileData), string(data))
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
