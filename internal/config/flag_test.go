package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		runAddr     string
		flagBaseURL string
		wantErr     string
	}{
		{
			name:        "не указаны флаги",
			args:        []string{},
			runAddr:     ":8080",
			flagBaseURL: "http://localhost:8080",
			wantErr:     "",
		},
		{
			name:        "флаги указаны",
			args:        []string{"-a=:9091", "-b=https://my.custom.domain/specific_url"},
			runAddr:     ":9091",
			flagBaseURL: "https://my.custom.domain",
			wantErr:     "",
		},
		{
			name:        "ошибочный флаг",
			args:        []string{"-b=\n"},
			runAddr:     "no_value",
			flagBaseURL: "no_value",
			wantErr:     "ошибка парсинга флага",
		},
		{
			name:        "флаг не ссылка",
			args:        []string{"-b=123@mail"},
			runAddr:     "no_value",
			flagBaseURL: "no_value",
			wantErr:     "флаг не является ссылкой",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedOSArgs := os.Args
			defer func() { os.Args = savedOSArgs }()
			os.Args = append([]string{os.Args[0]}, tt.args...)

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
