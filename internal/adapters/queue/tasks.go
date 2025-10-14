package queue

import "encoding/json"

const TypeResizeImage = "resize_image"

type ResizeImagePayload struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func NewResizeImageTask(url string, width, height int) ([]byte, error) {
	payload, err := json.Marshal(ResizeImagePayload{
		URL:    url,
		Width:  width,
		Height: height,
	})
	if err != nil {
		return nil, err
	}
	return payload, nil
}
