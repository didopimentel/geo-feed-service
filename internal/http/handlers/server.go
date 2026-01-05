package handlers

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type UseCases struct {
	IngestionAPIUseCases
	FeedAPIUseCases
	HealthAPIUseCases
}

func NewServer(ucs UseCases) *chi.Mux {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	healthAPI := NewHealthAPI(ucs.HealthAPIUseCases)
	ingestionAPI := NewIngestionAPI(ucs.IngestionAPIUseCases)
	feedAPI := NewFeedAPI(ucs.FeedAPIUseCases)

	r.Get("/health", healthAPI.GetHealth)
	r.Post("/ingestion", ingestionAPI.CreateContent)
	r.Get("/feed", feedAPI.GetFeed)

	return r
}
