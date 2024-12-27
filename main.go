package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	engine := html.New("./templates", ".tpl")
	app := fiber.New(fiber.Config{
		AppName: "Kiach",
		Views:   engine,
	})

	storageService := NewCloudinaryStorageService()

	var videoURL string

	app.Post("/upload", func(c *fiber.Ctx) error {
		filePath := os.Getenv("VIDEO_PATH")
		localFile, err := os.Open(filePath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open local file"})
		}
		defer localFile.Close()

		video, err := storageService.Upload(localFile)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// Store the video URL
		videoURL = video.URL

		return c.JSON(fiber.Map{"message": "Video uploaded successfully"})
	})

	// Home route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", nil)
	})

	app.Get("/stream", func(c *fiber.Ctx) error {
		// Use the stored video URL
		return streamVideo(c, videoURL)
	})

	// Start the Fiber server
	port := os.Getenv("PORT")
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}

func streamVideo(c *fiber.Ctx, url string) error {
	// Check if the URL parameter is empty
	if url == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request: URL parameter missing")
	}

	fmt.Printf("Streaming video from URL: %s\n", url)

	// Fetch the video from the given URL
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching video: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	defer resp.Body.Close()

	// Set necessary headers for the response
	c.Set("Content-Type", resp.Header.Get("Content-Type"))
	c.Set("Content-Disposition", "inline")

	// Set Content-Length header if available
	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		c.Set(fiber.HeaderContentLength, contentLength)
	}

	// Handle range requests
	rangeHeader := c.GetReqHeaders()["Range"]
	if len(rangeHeader) > 0 {
		parts := strings.SplitN(rangeHeader[0], "=", 2)
		if len(parts) != 2 {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid Range Header format")
		}

		start := parts[1]
		if !strings.ContainsAny(start, "0-9") {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid start byte position")
		}

		startInt, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		endInt := startInt + (resp.ContentLength - 1)

		c.Set(fiber.HeaderContentLength, strconv.FormatInt(endInt-startInt+1, 10))
		c.Set(fiber.HeaderAcceptRanges, "bytes")
		c.Status(fiber.StatusPartialContent)

		_, err = resp.Body.(io.Seeker).Seek(startInt, io.SeekStart)
		if err != nil {
			log.Printf("Error seeking to start position: %v", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}
	}

	// Copy the response body to the client
	_, copyErr := io.Copy(c.Response().BodyWriter(), resp.Body)
	if copyErr != nil {
		log.Printf("Error copying entire file to response: %v", copyErr)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	fmt.Println("streaming was successful !!!")

	return nil
}
