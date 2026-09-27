package handler

import "net/http"

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Сервер запущен :)`))
}
