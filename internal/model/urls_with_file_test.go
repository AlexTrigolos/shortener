package model

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewURLWithFile(t *testing.T) {
	urls, err := NewURL()
	require.NoError(t, err)

	dirname := t.TempDir()
	defer os.RemoveAll(dirname)
	file, err := os.CreateTemp(dirname, "*_file.json")
	require.NoError(t, err)
	defer file.Close()

	fileData, err := json.MarshalIndent([]map[string]string{{"uuid": "1", "short_url": "4rSPg8ap", "original_url": "http://yandex.ru"}}, "", "	")
	require.NoError(t, err)
	_, err = file.Write(fileData)
	require.NoError(t, err)

	filename := file.Name()

	newFile, err := os.CreateTemp(dirname, "*_new_file.json")
	require.NoError(t, err)
	defer newFile.Close()
	emptyData, err := json.Marshal([]string{})
	require.NoError(t, err)

	_, err = newFile.Write(emptyData)
	require.NoError(t, err)

	newFilename := newFile.Name()

	tests := []struct {
		name         string
		filename     string
		wantFileData []byte
		want         *URLWithFile
	}{
		{
			name:         "есть файл с сохранениями",
			filename:     filename,
			wantFileData: fileData,
			want: &URLWithFile{
				URL:             urls,
				fileStoragePath: filename,
			},
		},
		{
			name:         "не было сохранений",
			filename:     newFilename,
			wantFileData: emptyData,
			want: &URLWithFile{
				URL:             urls,
				fileStoragePath: newFilename,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newURLWithFile, err := NewURLWithFile(tt.filename)
			require.NoError(t, err)
			assert.Equal(t, reflect.TypeOf(tt.want), reflect.TypeOf(newURLWithFile))

			data, err := os.ReadFile(tt.filename)
			require.NoError(t, err)
			assert.JSONEq(t, string(tt.wantFileData), string(data))
		})
	}
}

func TestURLWithFile_Set(t *testing.T) {
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
			defer file.Close()
			filename := file.Name()

			wantFileData, err := json.Marshal([]string{})
			require.NoError(t, err)

			urlsCount := 1
			urlsWithFile := &URLWithFile{URL: &URL{data: map[string]string{}}, fileStoragePath: filename}
			if tt.name != "добавление нового ключа в пустой" {
				urlsWithFile.URL.data["key1"] = "value1"

				newData, err := addingSlice{key: "key1", value: "value1", count: urlsCount}.marshalJSON()
				require.NoError(t, err)
				urlsCount++
				wantFileData = append(wantFileData[:len(wantFileData)-1], newData...)
			}

			_, err = file.Write(wantFileData)
			require.NoError(t, err)

			if tt.wantErr == "" {
				newData, err := addingSlice{key: "key0", value: "value0", count: urlsCount}.marshalJSON()
				require.NoError(t, err)
				lenWantFileData := len(wantFileData) - 1
				if urlsCount == 2 {
					lenWantFileData -= 1
				}
				wantFileData = append(wantFileData[:lenWantFileData], newData...)
			}

			err = urlsWithFile.Set(t.Context(), tt.key, tt.value)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Contains(t, urlsWithFile.URL.data, tt.key)
				assert.Equal(t, urlsWithFile.URL.data[tt.key], tt.value)
			}

			data, err := os.ReadFile(filename)
			require.NoError(t, err)
			assert.JSONEq(t, string(wantFileData), string(data))
		})
	}
}

func TestURLWithFile_Get(t *testing.T) {
	tests := []struct {
		name    string
		urls    *URLWithFile
		key     string
		wantVal string
		wantErr string
	}{
		{
			name:    "получает значение по ключу",
			urls:    &URLWithFile{URL: &URL{data: map[string]string{"key1": "value1"}}},
			key:     "key1",
			wantVal: "value1",
			wantErr: "",
		},
		{
			name:    "не находит значение по ключу",
			urls:    &URLWithFile{URL: &URL{data: map[string]string{}}},
			key:     "key1",
			wantVal: "",
			wantErr: "такой ключ не найден",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := tt.urls
			value, err := urls.Get(t.Context(), tt.key)
			assert.Equal(t, tt.wantVal, value)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestURLWithFile_PingContext(t *testing.T) {
	t.Run("нет доступа к бд", func(t *testing.T) {
		uwf := &URLWithFile{}
		gotErr := uwf.PingContext(t.Context())
		require.ErrorContains(t, gotErr, "нет доступа к бд")
	})
}
