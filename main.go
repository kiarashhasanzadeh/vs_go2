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

	// Video upload route
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

		return c.JSON(video)
	})

	// Home route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", nil)
	})

	// Video streaming route
	app.Get("/stream", streamVideo)

	// Start the Fiber server
	port := os.Getenv("PORT")
	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}

func streamVideo(c *fiber.Ctx) error {

	url := "https://res.cloudinary.com/dh0erjteo/video/upload/v1734107266/videos/ubcrjofobp3pgubrkzpg.mp4"

	// Check if the URL is provided
	if url == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Bad Request: URL parameter missing")
	}

	fmt.Printf("Streaming video from URL: %s\n", url)

	// Make an HTTP GET request to fetch the video
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching video: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}
	defer resp.Body.Close()

	// Get the Content-Length header
	contentLength := resp.Header.Get("Content-Length")
	var fileSize int64
	if contentLength != "" {
		if parsed, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			fileSize = parsed
		} else {
			log.Printf("Error parsing Content-Length: %v", err)
		}
	} else {
		log.Println("Content-Length header not present")
	}

	// Get the Range header from the request
	rangeHeader := c.GetReqHeaders()["Range"]

	var startInt, endInt int64

	// Check if a Range header was provided in the request
	if len(rangeHeader) > 0 {
		parts := strings.SplitN(rangeHeader[0], "=", 2)
		if len(parts) != 2 {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid Range Header format")
		}

		start, end := parts[1], ""
		if start != "" && !strings.ContainsAny(start, "0-9") {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid start byte position")
		}

		if start != "" {
			startInt, err = strconv.ParseInt(start, 10, 64)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}

			if end != "" {
				endInt, err = strconv.ParseInt(end, 10, 64)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
				}
			} else {
				endInt = fileSize - 1
			}
		}

		// Set headers for partial content
		c.Set(fiber.HeaderContentLength, strconv.FormatInt(endInt-startInt+1, 10))
		c.Set(fiber.HeaderAcceptRanges, "bytes")
		c.Status(fiber.StatusPartialContent)

		// Seek to the start position in the response body
		_, err = resp.Body.(io.Seeker).Seek(startInt, io.SeekStart)
		if err != nil {
			log.Printf("Error seeking to start position: %v", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}

		// Set Content-Length for partial content
		c.Set("Content-Length", strconv.FormatInt(fileSize, 10))

		// Copy bytes from response body to client
		_, copyErr := io.Copy(c.Response().BodyWriter(), resp.Body)
		if copyErr != nil {
			log.Printf("Error copying entire file to response: %v", copyErr)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}
	}

	// If we reach here, everything went fine, so return nil
	return nil
}
