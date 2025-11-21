package seaweed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"shared/constants"
)

type AssignResponse struct {
	Fid       string `json:"fid"`
	Url       string `json:"url"`
	PublicUrl string `json:"publicUrl"`
	Count     int    `json:"count"`
}

type UploadResult struct {
	Fid       string `json:"fid"`
	PublicURL string `json:"public_url"`
	URL       string `json:"url"`
}

func UploadToSeaweedFS(file multipart.File, fileName string) (*UploadResult, error) {
	// Step 1: Get assignment from Seaweed master
	fmt.Println("📦 Uploading file to Seaweed")
	fmt.Println("📦 Getting assignment from Seaweed master")
	assignResp, err := http.Get(fmt.Sprintf("http://%s:%s/dir/assign", constants.SeaweedFSHost, constants.SeaweedFSPort))
	if err != nil {
		return nil, fmt.Errorf("failed to assign fid: %w", err)
	}
	defer assignResp.Body.Close()

	var assign AssignResponse
	if err := json.NewDecoder(assignResp.Body).Decode(&assign); err != nil {
		return nil, fmt.Errorf("failed to parse assign response: %w", err)
	}

	// Step 2: Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}
	writer.Close()

	// Step 3: Upload to volume server
	fmt.Println("📦 Uploading file to Seaweed volume server")
	uploadURL := fmt.Sprintf("http://%s/%s", assign.Url, assign.Fid)
	req, err := http.NewRequest("POST", uploadURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload requests: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(body))
	}

	fmt.Println("📦 File uploaded successfully")
	return &UploadResult{
		Fid:       assign.Fid,
		PublicURL: fmt.Sprintf("http://%s/%s", assign.PublicUrl, assign.Fid),
		URL:       fmt.Sprintf("http://%s/%s", assign.Url, assign.Fid),
	}, nil
}

func GetSeaweedPublicURL(fid string) string {
	return fmt.Sprintf("http://%s:%s/%s", constants.SeaweedFSHost, constants.SeaweedFSPort, fid)
}

func DownloadFromSeaweedFS(fileURL string) ([]byte, error) {
	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("file fetch failed (%d): %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return data, nil
}
