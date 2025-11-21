package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"shared/fileutils"
	"strings"
	"time"
)

type HttpClient struct {
	ctx    context.Context
	client *http.Client
}

func NewHttpClient(ctx context.Context) *HttpClient {

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &HttpClient{
		ctx:    ctx,
		client: client,
	}
}

func (c *HttpClient) Get(uri string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *HttpClient) Post(uri string, body interface{}) (*http.Response, error) {

	if body == nil {
		req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, uri, nil)
		if err != nil {
			return nil, err
		}
		return c.client.Do(req)
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, uri, bytes.NewReader(jsonBytes))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

func (c *HttpClient) Put(uri string, body interface{}) (*http.Response, error) {

	if body == nil {
		req, err := http.NewRequestWithContext(c.ctx, http.MethodPut, uri, nil)
		if err != nil {
			return nil, err
		}
		return c.client.Do(req)
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(c.ctx, http.MethodPut, uri, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

func (c *HttpClient) Delete(uri string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *HttpClient) Patch(uri string, body interface{}) (*http.Response, error) {
	if body == nil {
		req, err := http.NewRequestWithContext(c.ctx, http.MethodPatch, uri, nil)
		if err != nil {
			return nil, err
		}

		return c.client.Do(req)
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(c.ctx, http.MethodPatch, uri, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

func (c *HttpClient) PostForm(uri string, form map[string]string) (*http.Response, error) {
	formValues := url.Values{}

	for key, value := range form {
		formValues.Add(key, value)
	}
	req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, uri, strings.NewReader(formValues.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.client.Do(req)
}

func (c *HttpClient) PostFormMultipart(uri string, fileName string, filePath string) (*http.Response, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	requestBody, contentType, err := fileutils.CreateMultipartForm(file, fileName)

	request, err := http.NewRequestWithContext(c.ctx, http.MethodPost, uri, requestBody)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", contentType)
	return c.client.Do(request)
}
