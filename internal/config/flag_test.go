package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseFlags(t *testing.T) {
	tests := []struct {
		name                string
		env                 [][2]string
		args                []string
		runAddr             string
		flagBaseURL         string
		flagFileStoragePath string
		wantErr             string
	}{
		{
			name:                "не указаны флаги и env",
			env:                 [][2]string{},
			args:                []string{},
			runAddr:             ":8080",
			flagBaseURL:         "http://localhost:8080",
			flagFileStoragePath: "./urls_storage.json",
			wantErr:             "",
		},
		{
			name:                "флаги указаны но нет env",
			env:                 [][2]string{},
			args:                []string{"-a=:9091", "-b=https://my.args.domain/specific_url", "-f=./flag_storage.json"},
			runAddr:             ":9091",
			flagBaseURL:         "https://my.args.domain",
			flagFileStoragePath: "./flag_storage.json",
			wantErr:             "",
		},
		{
			name:                "флаги не указаны но есть env",
			env:                 [][2]string{{"SERVER_ADDRESS", ":9099"}, {"BASE_URL", "https://my.env.domain/specific_url"}, {"FILE_STORAGE_PATH", "./env_storage.json"}},
			args:                []string{},
			runAddr:             ":9099",
			flagBaseURL:         "https://my.env.domain",
			flagFileStoragePath: "./env_storage.json",
			wantErr:             "",
		},
		{
			name:                "флаги и env указаны",
			env:                 [][2]string{{"SERVER_ADDRESS", ":9099"}, {"BASE_URL", "https://my.env.domain/specific_url"}, {"FILE_STORAGE_PATH", "./env_storage.json"}},
			args:                []string{"-a=:9091", "-b=https://my.args.domain/specific_url", "-f=./flag_storage.json"},
			runAddr:             ":9099",
			flagBaseURL:         "https://my.env.domain",
			flagFileStoragePath: "./env_storage.json",
			wantErr:             "",
		},
		{
			name:                "ошибочный флаг",
			env:                 [][2]string{},
			args:                []string{"-b=\n"},
			runAddr:             "no_value",
			flagBaseURL:         "no_value",
			flagFileStoragePath: "no_value",
			wantErr:             "ошибка парсинга флага",
		},
		{
			name:                "флаг не ссылка",
			env:                 [][2]string{},
			args:                []string{"-b=123@mail"},
			runAddr:             "no_value",
			flagBaseURL:         "no_value",
			flagFileStoragePath: "no_value",
			wantErr:             "флаг не является ссылкой",
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
				assert.Equal(t, tt.runAddr, config.GetServerAddr())
				assert.Equal(t, tt.flagBaseURL, config.GetBaseURL())
			}
		})
	}
}
