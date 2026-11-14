package handler

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type DB interface {
	PingContext(ctx context.Context) error
}

type Ping struct {
	db     DB
	logger *zap.Logger
}

func NewPing(db DB, logger *zap.Logger) *Ping {
	return &Ping{db: db, logger: logger}
}

func (p *Ping) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		if err := p.db.PingContext(ctx); err != nil {
			p.logger.Error("бд не отвечает", zap.String("Error", err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`DB still not dead :D`))
	}
}
