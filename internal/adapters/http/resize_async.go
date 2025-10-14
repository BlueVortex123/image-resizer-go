package httpadapter

import (
	"fmt"
	"net/http"
	"strconv"
)

// ResizeAsync handles async resize requests by queuing them
func (serverReference *Server) ResizeAsync(response http.ResponseWriter, request *http.Request) {
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

	// Check if async processing is supported
	if !serverReference.resizerSvc.SupportsAsync() {
		http.Error(response, "Async processing not configured", http.StatusServiceUnavailable)
		return
	}

	// Enqueue the resize task
	err := serverReference.resizerSvc.EnqueueResizeTask(url, widthValue, heightValue)
	if err != nil {
		http.Error(response, fmt.Sprintf("Failed to enqueue task: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusAccepted)
	_, _ = response.Write([]byte(`{"status":"queued","message":"Resize task queued successfully"}`))
}
