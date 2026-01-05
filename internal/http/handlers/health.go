package handlers

import (
	"net/http"
)

type HealthAPIUseCases interface{}

type HealthAPI struct {
	uc HealthAPIUseCases
}

func NewHealthAPI(uc HealthAPIUseCases) *HealthAPI {
	return &HealthAPI{uc: uc}
}

func (api *HealthAPI) GetHealth(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("ok"))
}
