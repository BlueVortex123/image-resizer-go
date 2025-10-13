package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	_ "golang.org/x/image/webp"

	"example.com/img-resizer/internal/domain/ports/requests"
	"github.com/disintegration/imaging"
)

func (resizerServiceRefference *ResizerService) GetResized(context context.Context, req requests.RequestGetResize) requests.ResponseGetResize {
	log.Printf("Resizing request received: %s (%dx%d)", req.URL, req.Width, req.Height)

	// Confirm request is valid
	if err := req.Validate(); err != nil {
		return requests.ResponseGetResize{Error: err}
	}

	// Fetch the image from the URL
	srcImg, format, err := resizerServiceRefference.makeHttpRequestAndGetImage(context, req.URL)
	if err != nil {
		return requests.ResponseGetResize{Error: err}
	}

	// Check cache by generated key: <URL>|<width>|<height>
	key := resizerServiceRefference.keyFor(req.URL, req.Width, req.Height)
	log.Printf("Resizing image. Key: %s", key)

	// If we have the image cached, return it
	if cached, cachedContentType, ok, err := resizerServiceRefference.cache.Get(context, key); err == nil && ok {
		log.Printf("Served from cache. Key: %s\n\n", key)
		return requests.ResponseGetResize{Data: cached, MimeType: cachedContentType, Error: nil}
	} else if err != nil {
		// log cache error but continue
		log.Println("cache get error:", err)
		_ = err
	}

	// Proceed to resizing the image
	data, contentType, err := resizerServiceRefference.resize(srcImg, format, req.Width, req.Height)
	if err != nil {
		return requests.ResponseGetResize{Error: err}
	}

	if err := resizerServiceRefference.cache.Set(context, key, data, contentType); err != nil {
		log.Println("cache set error:", err)
		_ = err
	}
	log.Printf("Image resized and cached. Key: %s\n\n", key)

	return requests.ResponseGetResize{Data: data, MimeType: contentType, Error: nil}
}

func (resizerServiceRefference *ResizerService) makeHttpRequestAndGetImage(ctx context.Context, url string) (image.Image, string, error) {
	client := &http.Client{
		Timeout: resizerServiceRefference.config.HTTP.Timeout * time.Second,
	}

	// Make the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("creating HTTP request failed: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("making HTTP request failed: %w", err)
	}
	defer resp.Body.Close() // Ensure the response body is closed

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("bad status from origin: %d", resp.StatusCode)
	}

	// Check Content-Type header to ensure it's an image
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, "", fmt.Errorf("URL did not return an image (got %s)", contentType)
	}

	// Limit the size of the response body to prevent memory issues
	limited := io.LimitReader(resp.Body, 10<<20) // 10 MB
	srcImg, format, err := image.Decode(limited)
	if err != nil {
		return nil, "", fmt.Errorf("decoding image failed: %w", err)
	}

	return srcImg, format, nil
}

func (resizerServiceRefference *ResizerService) resize(srcImg image.Image, format string, width, height int) ([]byte, string, error) {
	dst := imaging.Resize(srcImg, width, height, imaging.Lanczos) // use the package github.com/disintegration/imaging
	var buf bytes.Buffer

	// Encode based on original format, default to JPEG
	switch format {
	case "jpeg", "jpg":
		if err := imaging.Encode(&buf, dst, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/jpeg", nil
	case "png":
		if err := imaging.Encode(&buf, dst, imaging.PNG); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/png", nil
	case "webp":
		// imaging can't encode webp. Fallback to PNG
		if err := imaging.Encode(&buf, dst, imaging.PNG); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/png", nil
	default:
		// default to JPEG
		if err := imaging.Encode(&buf, dst, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/jpeg", nil
	}
}
