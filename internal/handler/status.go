package handler

import "net/http"

func Status(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Сервер запущен :)`))
}
