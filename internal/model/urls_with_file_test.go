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
				URL:               urls,
				fileStoragePath:   filename,
				transactionBuffer: nil,
			},
		},
		{
			name:         "не было сохранений",
			filename:     newFilename,
			wantFileData: emptyData,
			want: &URLWithFile{
				URL:               urls,
				fileStoragePath:   newFilename,
				transactionBuffer: nil,
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
			value:   "value2",
			wantErr: "такой ключ уже занят",
		},
		{
			name:    "конфликт значений",
			key:     "key1",
			value:   "value1",
			wantErr: "conflict",
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
			urlsWithFile := &URLWithFile{URL: &URL{data: map[string]string{}, values: map[string]string{}}, fileStoragePath: filename}
			if tt.name != "добавление нового ключа в пустой" {
				urlsWithFile.URL.Set(t.Context(), "key1", "value1")

				newData, err := addingSlice{key: "key1", value: "value1", count: urlsCount}.marshalJSON()
				require.NoError(t, err)
				urlsCount++
				wantFileData = append(wantFileData[:len(wantFileData)-1], newData...)
				wantFileData = append(wantFileData, []byte("\n]")...)
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
				wantFileData = append(wantFileData, []byte("\n]")...)
			}

			err = urlsWithFile.Set(t.Context(), tt.key, tt.value)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Contains(t, urlsWithFile.URL.data, tt.key)
				assert.Equal(t, urlsWithFile.URL.data[tt.key], tt.value)
				assert.Contains(t, urlsWithFile.URL.values, tt.value)
				assert.Equal(t, urlsWithFile.URL.values[tt.value], tt.key)
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
			urls:    &URLWithFile{URL: &URL{data: map[string]string{"key1": "value1"}, values: map[string]string{"value1": "key1"}}},
			key:     "key1",
			wantVal: "value1",
			wantErr: "",
		},
		{
			name:    "не находит значение по ключу",
			urls:    &URLWithFile{URL: &URL{data: map[string]string{}, values: map[string]string{}}},
			key:     "key1",
			wantVal: "",
			wantErr: "не найден ключ key1",
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

func TestURLWithFile_GetByOriginal(t *testing.T) {
	tests := []struct {
		name    string
		urls    *URLWithFile
		value   string
		wantKey string
		wantErr string
	}{
		{
			name:    "получает ключ по значению",
			urls:    &URLWithFile{URL: &URL{data: map[string]string{"key1": "value1"}, values: map[string]string{"value1": "key1"}}},
			value:   "value1",
			wantKey: "key1",
			wantErr: "",
		},
		{
			name:    "не находит ключ по значению",
			urls:    &URLWithFile{URL: &URL{data: map[string]string{}, values: map[string]string{}}},
			value:   "value1",
			wantKey: "",
			wantErr: "не нашли значение value1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls := tt.urls
			key, err := urls.GetByOriginal(t.Context(), tt.value)
			assert.Equal(t, tt.wantKey, key)

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

func TestURLWithFile_Begin(t *testing.T) {
	dirname := t.TempDir()
	defer os.RemoveAll(dirname)

	file, err := os.CreateTemp(dirname, "*_file.json")
	require.NoError(t, err)
	defer file.Close()
	emptyData, err := json.Marshal([]string{})
	require.NoError(t, err)

	_, err = file.Write(emptyData)
	require.NoError(t, err)

	t.Run("удачное создание транзакции", func(t *testing.T) {
		uwf, err := NewURLWithFile(file.Name())
		require.NoError(t, err)

		tr, err := uwf.Begin(5)
		require.NoError(t, err)
		assert.Equal(t, &Transaction{tx: uwf, selfSetTx: uwf, commited: false}, tr)
		assert.NotNil(t, uwf.transactionBuffer)
		assert.False(t, uwf.URL.mu.TryLock())
	})
}

func TestURLWithFile_Rollback(t *testing.T) {
	dirname := t.TempDir()
	defer os.RemoveAll(dirname)

	file, err := os.CreateTemp(dirname, "*_file.json")
	require.NoError(t, err)
	defer file.Close()
	emptyData, err := json.Marshal([]string{})
	require.NoError(t, err)

	_, err = file.Write(emptyData)
	require.NoError(t, err)

	t.Run("удачная отмена транзакции", func(t *testing.T) {
		uwf, err := NewURLWithFile(file.Name())
		require.NoError(t, err)

		_, err = uwf.Begin(5)
		require.NoError(t, err)

		err = uwf.Rollback()
		require.NoError(t, err)
		assert.Nil(t, uwf.transactionBuffer)
		assert.True(t, uwf.URL.mu.TryLock())
	})
}

func TestURLWithFile_Commit(t *testing.T) {
	dirname := t.TempDir()
	defer os.RemoveAll(dirname)

	file, err := os.CreateTemp(dirname, "*_file.json")
	require.NoError(t, err)
	defer file.Close()
	emptyData, err := json.Marshal([]string{})
	require.NoError(t, err)

	_, err = file.Write(emptyData)
	require.NoError(t, err)

	t.Run("удачное завершение транзакции", func(t *testing.T) {
		uwf, err := NewURLWithFile(file.Name())
		require.NoError(t, err)

		_, err = uwf.Begin(5)
		require.NoError(t, err)

		uwf.transactionBuffer["short_1"] = "long_1"
		uwf.transactionBuffer["short_2"] = "long_2"
		want1 := "[\n\t{\"original_url\":\"long_1\",\"short_url\":\"short_1\",\"uuid\":\"1\"},\n\t{\"original_url\":\"long_2\",\"short_url\":\"short_2\",\"uuid\":\"2\"}\n]"
		want2 := "[\n\t{\"original_url\":\"long_2\",\"short_url\":\"short_2\",\"uuid\":\"1\"},\n\t{\"original_url\":\"long_1\",\"short_url\":\"short_1\",\"uuid\":\"2\"}\n]"

		err = uwf.Commit()
		require.NoError(t, err)
		assert.Nil(t, uwf.transactionBuffer)
		assert.True(t, uwf.URL.mu.TryLock())

		assert.Len(t, uwf.URL.data, 2)
		assert.Contains(t, uwf.URL.data, "short_1")
		assert.Equal(t, uwf.URL.data["short_1"], "long_1")
		assert.Contains(t, uwf.URL.data, "short_2")
		assert.Equal(t, uwf.URL.data["short_2"], "long_2")

		assert.Len(t, uwf.URL.values, 2)
		assert.Contains(t, uwf.URL.values, "long_1")
		assert.Equal(t, uwf.URL.values["long_1"], "short_1")
		assert.Contains(t, uwf.URL.values, "long_2")
		assert.Equal(t, uwf.URL.values["long_2"], "short_2")

		data, err := os.ReadFile(file.Name())
		require.NoError(t, err)
		require.True(t, want1 == string(data) || want2 == string(data))
	})
}

func TestURLWithFile_SetTx(t *testing.T) {
	dirname := t.TempDir()
	defer os.RemoveAll(dirname)

	file, err := os.CreateTemp(dirname, "*_file.json")
	require.NoError(t, err)
	defer file.Close()
	data, err := json.MarshalIndent([]map[string]string{{"uuid": "1", "short_url": "key_0", "original_url": "value_0"}}, "", "	")
	require.NoError(t, err)

	_, err = file.Write(data)
	require.NoError(t, err)

	t.Run("удачное добавление внутри транзакции", func(t *testing.T) {
		uwf, err := NewURLWithFile(file.Name())
		require.NoError(t, err)

		_, err = uwf.Begin(5)
		require.NoError(t, err)

		err = uwf.SetTx(t.Context(), "key_1", "value_1")
		require.NoError(t, err)
		err = uwf.SetTx(t.Context(), "key_2", "value_2")
		require.NoError(t, err)
		assert.NotNil(t, uwf.transactionBuffer)
		assert.False(t, uwf.URL.mu.TryLock())

		assert.Len(t, uwf.URL.data, 1)

		assert.Len(t, uwf.transactionBuffer, 2)
		assert.Contains(t, uwf.transactionBuffer, "key_1")
		assert.Equal(t, uwf.transactionBuffer["key_1"], "value_1")
		assert.Contains(t, uwf.transactionBuffer, "key_2")
		assert.Equal(t, uwf.transactionBuffer["key_2"], "value_2")
	})
	t.Run("неудачное добавление внутри транзакции", func(t *testing.T) {
		uwf, err := NewURLWithFile(file.Name())
		require.NoError(t, err)

		_, err = uwf.Begin(5)
		require.NoError(t, err)

		err = uwf.SetTx(t.Context(), "key_0", "value_0")
		require.ErrorContains(t, err, "key_0 ключ уже занят")
		assert.NotNil(t, uwf.transactionBuffer)
		assert.False(t, uwf.URL.mu.TryLock())

		assert.Len(t, uwf.URL.data, 1)
		assert.Len(t, uwf.transactionBuffer, 0)
	})
	t.Run("неудачное добавление внутри транзакции из-за совпадения с временным хранилищем", func(t *testing.T) {
		uwf, err := NewURLWithFile(file.Name())
		require.NoError(t, err)

		_, err = uwf.Begin(5)
		require.NoError(t, err)

		err = uwf.SetTx(t.Context(), "key_1", "value_1")
		require.NoError(t, err)
		err = uwf.SetTx(t.Context(), "key_1", "value_1")
		require.ErrorContains(t, err, "key_1 ключ уже занят")
		assert.NotNil(t, uwf.transactionBuffer)
		assert.False(t, uwf.URL.mu.TryLock())

		assert.Len(t, uwf.URL.data, 1)
		assert.Len(t, uwf.transactionBuffer, 1)
	})
}
