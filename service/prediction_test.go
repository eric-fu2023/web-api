package service

import (
	"context"
	"testing"
	"web-api/repository"
)

func TestPredictionGetList(t *testing.T) {
	mockRepo := repository.NewMockPredictionRepo()

	_, err := mockRepo.GetList(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

}
