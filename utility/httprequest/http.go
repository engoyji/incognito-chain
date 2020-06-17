package httprequest

import (
	"bytes"
	"net/http"
)

//Send ...
func Send(url, method string, header http.Header, body []byte) (*http.Response, error) {

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))

	req.Header = header

	// req.Header.Set("X-Custom-Header", "myvalue")
	// req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
