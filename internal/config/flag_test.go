package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseFlags(t *testing.T) {
	tests := []struct {
		name            string
		env             [][2]string
		args            []string
		runAddr         string
		BaseURL         string
		FileStoragePath string
		DataBase        string
		wantErr         string
	}{
		{
			name:            "не указаны флаги и env",
			env:             [][2]string{},
			args:            []string{},
			runAddr:         ":8080",
			BaseURL:         "http://localhost:8080",
			FileStoragePath: "",
			DataBase:        "",
			wantErr:         "",
		},
		{
			name:            "флаги указаны но нет env",
			env:             [][2]string{},
			args:            []string{"-a=:9091", "-b=https://my.args.domain/specific_url", "-f=./flag_storage.json", "-d=flag_db_string"},
			runAddr:         ":9091",
			BaseURL:         "https://my.args.domain",
			FileStoragePath: "./flag_storage.json",
			DataBase:        "flag_db_string",
			wantErr:         "",
		},
		{
			name:            "флаги не указаны но есть env",
			env:             [][2]string{{"SERVER_ADDRESS", ":9099"}, {"BASE_URL", "https://my.env.domain/specific_url"}, {"FILE_STORAGE_PATH", "./env_storage.json"}, {"DATABASE_DSN", "env_db_string"}},
			args:            []string{},
			runAddr:         ":9099",
			BaseURL:         "https://my.env.domain",
			FileStoragePath: "./env_storage.json",
			DataBase:        "env_db_string",
			wantErr:         "",
		},
		{
			name:            "флаги и env указаны",
			env:             [][2]string{{"SERVER_ADDRESS", ":9099"}, {"BASE_URL", "https://my.env.domain/specific_url"}, {"FILE_STORAGE_PATH", "./env_storage.json"}, {"DATABASE_DSN", "env_db_string"}},
			args:            []string{"-a=:9091", "-b=https://my.args.domain/specific_url", "-f=./flag_storage.json", "-d=flag_db_string"},
			runAddr:         ":9099",
			BaseURL:         "https://my.env.domain",
			FileStoragePath: "./env_storage.json",
			DataBase:        "env_db_string",
			wantErr:         "",
		},
		{
			name:            "ошибочный флаг",
			env:             [][2]string{},
			args:            []string{"-b=\n"},
			runAddr:         "no_value",
			BaseURL:         "no_value",
			FileStoragePath: "no_value",
			DataBase:        "no_value",
			wantErr:         "ошибка парсинга флага",
		},
		{
			name:            "флаг не ссылка",
			env:             [][2]string{},
			args:            []string{"-b=123@mail"},
			runAddr:         "no_value",
			BaseURL:         "no_value",
			FileStoragePath: "no_value",
			DataBase:        "no_value",
			wantErr:         "флаг не является ссылкой",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedOSArgs := os.Args
			defer func() { os.Args = savedOSArgs }()
			os.Args = append([]string{os.Args[0]}, tt.args...)

			for _, envData := range tt.env {
				os.Setenv(envData[0], envData[1])
				defer func() { os.Unsetenv(envData[0]) }()
			}

			config, err := ParseFlags()
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.runAddr, config.Addr)
				assert.Equal(t, tt.BaseURL, config.BaseURL)
				assert.Equal(t, tt.FileStoragePath, config.FileStoragePath)
				assert.Equal(t, tt.DataBase, config.DataBase)
			}
		})
	}
}
