package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/stretchr/testify/assert"
)

var (
	compactCalled bool
	statusCalled  bool
)

// Мок для CompactHandler
func mockCompactHandler(w http.ResponseWriter, r *http.Request, urls *model.URL) {
	compactCalled = true
}

// Мок для StatusHandler
func mockStatusHandler(w http.ResponseWriter, r *http.Request) {
	statusCalled = true
}

func TestRootHandler(t *testing.T) {
	// ИИ написал, что оих надо бы возвращать, но я не вижу смысла это же правки в отдельной gorutine который не должен бы влиять ни на что

	// // Сохраняем оригинальные функции, чтобы вернуть их после теста
	// origCompact := CompactHndlr
	// origStatus := StatusHndlr

	// // Обязательно восстанавливаем в конце!
	// defer func() {
	// 	CompactHndlr = origCompact
	// 	StatusHndlr = origStatus
	// }()

	// Подменяем реализации на наши моки
	CompactHndlr = mockCompactHandler
	StatusHndlr = mockStatusHandler

	type want struct {
		Code  int
		Compact bool
		Status  bool
	}

	tests := []struct {
		name   string
		method string
		want   want
	}{
		{
			name: "POST запрос долже вызывать CompactHandler",
			method: http.MethodPost,
			want: want{
				Code: http.StatusOK,
				Compact: true,
				Status: false,
			},
		},
		{
			name: "GET запрос должен вызывать StatusHandler",
			method: http.MethodGet,
			want: want{
				Code: http.StatusOK,
				Compact: false,
				Status: true,
			},
		},
		{
			name: "PUT запрос должен возвращать 405",
			method: http.MethodPut,
			want: want{
				Code: http.StatusMethodNotAllowed,
				Compact: false,
				Status: false,
			},
		},
		{
			name: "DELETE запрос должен возвращать 405",
			method: http.MethodDelete,
			want: want{
				Code: http.StatusMethodNotAllowed,
				Compact: false,
				Status: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compactCalled = false
			statusCalled = false

			req := httptest.NewRequest(tt.method, "/", nil)
			rr := httptest.NewRecorder()

			RootHandler(nil)(rr, req)

			assert.Equal(t, tt.want.Code, rr.Code)

			assert.Equal(t, tt.want.Compact, compactCalled)
			assert.Equal(t, tt.want.Status, statusCalled)
		})
	}
}
