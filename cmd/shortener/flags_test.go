package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseFlags(t *testing.T) {
	parseFlags()
	tests := []struct {
		name           string
		args           []string
		runAddr        string
		flagDefaultURL string
	}{
		{
			name:           "не указаны флаги",
			args:           []string{},
			runAddr:        ":8080",
			flagDefaultURL: "http://localhost:8080/unknown_short_url",
		},
		{
			name:           "флаги указаны",
			args:           []string{"-a=:9091", "-b=https://my.custom.domain/specific_url"},
			runAddr:        ":9091",
			flagDefaultURL: "https://my.custom.domain/specific_url",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = append(os.Args, tt.args...)
			flag.Parse()
			assert.Equal(t, tt.runAddr, flagRunAddr)
			assert.Equal(t, tt.flagDefaultURL, flagDefaultURL)
		})
	}
}
