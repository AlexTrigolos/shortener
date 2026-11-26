package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	Secret = "abc"
	Signed = "new_user_id.c257a6475d17f21450bbf1302b873c7122496716da1dbb0b87da8aa5aaa6d8a0"
	UserID = "new_user_id"
)

func TestNewSigner(t *testing.T) {
	t.Run("удачно создал подпись", func(t *testing.T) {
		got := NewSigner(Secret)
		expectedSigner := &Signer{key: []byte(Secret)}
		assert.Equal(t, expectedSigner, got)
	})
}

func TestSigner_Sign(t *testing.T) {
	t.Run("удачно подписали", func(t *testing.T) {
		s := NewSigner(Secret)
		got := s.Sign(UserID)
		assert.Equal(t, Signed, got)
	})
}

func TestSigner_Verify(t *testing.T) {
	s := NewSigner(Secret)
	tests := []struct {
		name   string
		signed string
		want   string
		wantOK bool
	}{
		{
			name:   "удачно проверили подпись",
			signed: Signed,
			want:   UserID,
			wantOK: true,
		},
		{
			name:   "нет значения",
			signed: ".e2636077506729a8f61aff2441332e40e844a8ad44489efd80210ea6d1f51088",
			want:   "",
			wantOK: false,
		},
		{
			name:   "ошибка в подписи",
			signed: "new_user_id.e2636077506729a8f61aff2441332e40e844a8ad44489efd80210ea6d1f51088",
			want:   "",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := s.Verify(tt.signed)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestSigner_Value(t *testing.T) {
	s := NewSigner(Secret)
	tests := []struct {
		name   string
		signed string
		want   string
	}{
		{
			name:   "удачно получили значение",
			signed: Signed,
			want:   UserID,
		},
		{
			name:   "нет значения",
			signed: ".e2636077506729a8f61aff2441332e40e844a8ad44489efd80210ea6d1f51088",
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Value(tt.signed)
			assert.Equal(t, tt.want, got)
		})
	}
}
