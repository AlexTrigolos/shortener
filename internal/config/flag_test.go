package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		runAddr     string
		flagBaseURL string
	}{
		{
			name:        "не указаны флаги",
			args:        []string{},
			runAddr:     ":8080",
			flagBaseURL: "http://localhost:8080",
		},
		{
			name:        "флаги указаны",
			args:        []string{"-a=:9091", "-b=https://my.custom.domain/specific_url"},
			runAddr:     ":9091",
			flagBaseURL: "https://my.custom.domain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			savedOSArgs := os.Args
			defer func() { os.Args = savedOSArgs }()
			os.Args = append([]string{os.Args[0]}, tt.args...)

			config := ParseFlags()

			assert.Equal(t, tt.runAddr, config.GetServerAddr())
			assert.Equal(t, tt.flagBaseURL, config.GetBaseURL())
		})
	}
}
