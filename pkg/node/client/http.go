package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"

	"github.com/axelburling/dfs/pkg/node/constants"
	"github.com/valyala/fasthttp"
)

type http struct {
	addr string
}

func newHttp(addr string) *http {
	return &http{
		addr: addr,
	}
}

type Response struct {
	Error   string `json:"error,omitempty"`
	Hash    string `json:"hash,omitempty"`
	Size    int `size:"size,omitempty"`
	Success bool `json:"success,omitempty"`
}


func (h *http) SendChunk(file io.Reader, index int64, name string) (*Response, error) {
	// Create a buffer to store the multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add the file to the form
	part, err := writer.CreateFormFile("chunk", name)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy the file content into the form
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Add the index field
	err = writer.WriteField("index", strconv.FormatInt(index, 10))
	if err != nil {
		return nil, fmt.Errorf("failed to write index field: %w", err)
	}

	// Close the writer to finalize the form
	writer.Close()

	// Create fasthttp request
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(fmt.Sprintf("%s%s", h.addr, constants.HTTP_RECEIVE_CHUNK)) // Assuming "/upload" is the correct endpoint
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType(writer.FormDataContentType())
	// req.Header.SetContentType(fiber.MIMEApplicationJSON)

	// Set the request body
	req.SetBody(buf.Bytes())

	// Create response object
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	// Send the request
	client := &fasthttp.Client{}
	err = client.Do(req, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response body
	body := resp.Body()
	
	var res Response

	err = json.Unmarshal(body, &res)

	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(res.Error) > 0 {
		return nil, fmt.Errorf(res.Error)
	}

	return &res, nil
}
