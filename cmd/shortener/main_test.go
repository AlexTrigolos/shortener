package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)


func mockSuccessRun() error {
	return nil
}

func mockFailRun() error {
	return errors.New("не побежал")
}

func TestMain(t *testing.T) {
	tests := []struct {
		name      string
		mock      func() error
		wantPanic bool
	}{
		{
			name:      "удачный запуск",
			mock:      mockSuccessRun,
			wantPanic: false,
		},
		{
			name:      "неудачный запуск",
			mock:      mockFailRun,
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run = tt.mock
			if tt.wantPanic {
				assert.PanicsWithError(t, "не побежал", main)
			} else {
				assert.NotPanics(t, main)
			}
		})
	}
}
