package handlers

import (
	"net/http"
)

type IngestionAPIUseCases interface{}

type IngestionAPI struct {
	uc IngestionAPIUseCases
}

func NewIngestionAPI(uc IngestionAPIUseCases) *IngestionAPI {
	return &IngestionAPI{uc: uc}
}

func (api *IngestionAPI) CreateContent(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}
