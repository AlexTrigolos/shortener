package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseFlags(t *testing.T) {
	tests := []struct {
		name        string
		env         [][2]string
		args        []string
		runAddr     string
		flagBaseURL string
	}{
		{
			name:        "не указаны флаги и env",
			env:         [][2]string{},
			args:        []string{},
			runAddr:     ":8080",
			flagBaseURL: "http://localhost:8080",
		},
		{
			name:        "флаги указаны но нет env",
			env:         [][2]string{},
			args:        []string{"-a=:9091", "-b=https://my.args.domain/specific_url"},
			runAddr:     ":9091",
			flagBaseURL: "https://my.args.domain",
		},
		{
			name:        "флаги не указаны но есть env",
			env:         [][2]string{{"SERVER_ADDRESS", ":9099"}, {"BASE_URL", "https://my.env.domain/specific_url"}},
			args:        []string{},
			runAddr:     ":9099",
			flagBaseURL: "https://my.env.domain",
		},
		{
			name:        "флаги и env указаны",
			env:         [][2]string{{"SERVER_ADDRESS", ":9099"}, {"BASE_URL", "https://my.env.domain/specific_url"}},
			args:        []string{"-a=:9091", "-b=https://my.args.domain/specific_url"},
			runAddr:     ":9099",
			flagBaseURL: "https://my.env.domain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedOSArgs := os.Args
			defer func() { os.Args = savedOSArgs }()
			os.Args = append(os.Args, tt.args...)

			for _, envData := range tt.env {
				os.Setenv(envData[0], envData[1])
				defer func() { os.Unsetenv(envData[0]) }()
			}

			config := ParseFlags()

			assert.Equal(t, tt.runAddr, config.GetServerAddr())
			assert.Equal(t, tt.flagBaseURL, config.GetBaseURL())
		})
	}
}
