package handlers

import (
	"net/http"
)

type FeedAPIUseCases interface{}

type FeedAPI struct {
	uc FeedAPIUseCases
}

func NewFeedAPI(uc FeedAPIUseCases) *FeedAPI {
	return &FeedAPI{uc: uc}
}

func (api *FeedAPI) GetFeed(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}
