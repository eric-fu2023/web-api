package service

import (
	"context"
	"testing"
	"web-api/repository"
)

func TestAnalystGetList(t *testing.T) {
	mockRepo := repository.NewMockAnalystRepo()

	_, err := mockRepo.GetList(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

}
