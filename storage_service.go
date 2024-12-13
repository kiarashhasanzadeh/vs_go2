package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type StorageService interface {
	Upload(file *os.File) (*Video, error)
}

type cloudinaryStorageService struct{}

func NewCloudinaryStorageService() *cloudinaryStorageService {
	return &cloudinaryStorageService{}
}

func (s *cloudinaryStorageService) Upload(file *os.File) (*Video, error) {
	ctx := context.Background()

	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	uploadResult, err := cld.Upload.Upload(ctx, file.Name(), uploader.UploadParams{
		Folder:       "videos",
		ResourceType: "video",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload video: %w", err)
	}

	return NewVideo(file.Name(), uploadResult.SecureURL), nil
}
