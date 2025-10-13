package httpadapter

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"example.com/img-resizer/internal/domain/ports/requests"
)

func (serverReference *Server) Resize(response http.ResponseWriter, request *http.Request) {
	// Get query params
	queryURL := request.URL.Query()
	url := queryURL.Get("url")
	width := queryURL.Get("width")
	height := queryURL.Get("height")

	// Check if we have valid params
	if url == "" || width == "" || height == "" {
		http.Error(response, "Missing params: url, width, height", http.StatusBadRequest)
		return
	}

	// Convert width and height to integers
	widthValue, widthValueErr := strconv.Atoi(width)
	if widthValueErr != nil {
		http.Error(response, "Invalid width parameter", http.StatusBadRequest)
		return
	}

	heightValue, heightValueErr := strconv.Atoi(height)
	if heightValueErr != nil {
		http.Error(response, "Invalid height parameter", http.StatusBadRequest)
		return
	}

	// Call service layer to get resized image
	resizedResponse := serverReference.resizerSvc.GetResized(context.Background(), requests.RequestGetResize{
		URL:    url,
		Width:  widthValue,
		Height: heightValue,
	}) // call GetResized from get_resized.go

	if resizedResponse.Error != nil {
		http.Error(response, fmt.Sprintf("Error resizing the image: %v", resizedResponse.Error), http.StatusInternalServerError)
		return
	}

	// Return the resized image
	response.Header().Set("Content-Type", resizedResponse.MimeType)
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(resizedResponse.Data)
}
