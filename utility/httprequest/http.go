package httprequest

import (
	"bytes"
	"net/http"
)

//Send ...
func Send(url, method string, header http.Header, body []byte) (*http.Response, error) {

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}

	req.Header = header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
