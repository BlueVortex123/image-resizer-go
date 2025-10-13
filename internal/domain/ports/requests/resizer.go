package requests

import "fmt"

type RequestGetResize struct {
	URL    string
	Width  int
	Height int
}

type ResponseGetResize struct {
	Data     []byte
	MimeType string
	Error    error
}

func (req *RequestGetResize) Validate() error {
	if req.Width <= 0 {
		return fmt.Errorf("width must be greater than 0")
	}

	if req.Height <= 0 {
		return fmt.Errorf("height must be greater than 0")
	}
	return nil
}
