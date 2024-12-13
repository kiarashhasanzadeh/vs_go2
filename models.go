package main

import (
	"time"

	"github.com/google/uuid"
)

type Video struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

func NewVideo(title, url string) *Video {
	return &Video{
		ID:        uuid.New().String(),
		Title:     title,
		URL:       url,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
}
