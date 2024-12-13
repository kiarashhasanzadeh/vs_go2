package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Defines the interface for uploading videos
type StorageService interface {
	// Uploads a video file and returns the uploaded Video object along with any error
	Upload(file *os.File) (*Video, error)
}

// Implements the StorageService interface
type cloudinaryStorageService struct{}

// Creates a new instance of cloudinaryStorageService
func NewCloudinaryStorageService() *cloudinaryStorageService {
	return &cloudinaryStorageService{}
}

// Implements the StorageService interface
func (s *cloudinaryStorageService) Upload(file *os.File) (*Video, error) {
	ctx := context.Background()

	// Initialize Cloudinary client with environment variables
	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	// Upload the video using Cloudinary
	uploadResult, err := cld.Upload.Upload(ctx, file.Name(), uploader.UploadParams{
		Folder:       "videos",
		ResourceType: "video",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %w", err)
	}

	// Return a new Video object with the uploaded video information
	return NewVideo(file.Name(), uploadResult.SecureURL), nil
}
