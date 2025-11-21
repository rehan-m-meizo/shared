package fileutils

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"shared/constants"
	"shared/pkgs/uuids"
	"shared/seaweed"
	"strings"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/chai2010/webp"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ExcelFile struct {
	SheetName string                   `json:"sheet_name"`
	Headers   []string                 `json:"headers"`
	Rows      []map[string]interface{} `json:"rows"`
}

type nOpCloser struct {
	*bytes.Reader
}

var imageEncodeSem = make(chan struct{}, 4) // allows 4 concurrent encodes
const MaxPixels = 40_000_000

func (nOpCloser) Close() error {
	return nil
}

func sanitize(s string) string {
	return strings.TrimSpace(strings.TrimSuffix(s, "*"))
}

// sanitizeHeader cleans and normalizes column headers
func sanitizeHeader(header string) string {
	return strings.TrimSpace(strings.TrimSuffix(header, "*"))
}

// sanitizeCellValue trims spaces from cell content
func sanitizeCellValue(cell string) string {
	return strings.TrimSpace(cell)
}

// parseSheet extracts structured data from a given sheet
func parseSheet(sheetName string, rows [][]string) (ExcelFile, error) {
	if len(rows) == 0 {
		return ExcelFile{}, fmt.Errorf("sheet %s has no rows", sheetName)
	}

	headers := make([]string, len(rows[0]))
	for i, h := range rows[0] {
		headers[i] = sanitizeHeader(h)
	}

	var parsedRows []map[string]interface{}
	for _, row := range rows[1:] {
		rowMap := make(map[string]interface{})
		empty := true

		for i, header := range headers {
			var val string
			if i < len(row) {
				val = sanitizeCellValue(row[i])
			}
			if val != "" {
				empty = false
			}
			rowMap[header] = val
		}

		if !empty {
			parsedRows = append(parsedRows, rowMap)
		}
	}

	return ExcelFile{
		SheetName: sheetName,
		Headers:   headers,
		Rows:      parsedRows,
	}, nil
}

// ProcessExcelFile reads the Excel file and returns structured data per sheet
func ProcessExcelFileFromMultiPart(file multipart.File) ([]ExcelFile, error) {
	return ProcessExcelFile(file)
}

// You can also use this for non-multipart streams (i.e., from disk or io.Reader)
func ProcessExcelFile(reader io.Reader) ([]ExcelFile, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	sheetNames := f.GetSheetList()
	if len(sheetNames) == 0 {
		return nil, fmt.Errorf("no sheets found in the Excel file")
	}

	// --- Case 1: Single sheet ---
	if len(sheetNames) == 1 {
		rows, err := f.GetRows(sheetNames[0])
		if err != nil {
			return nil, fmt.Errorf("failed to read sheet %s: %w", sheetNames[0], err)
		}

		sheet, err := parseSheet(sheetNames[0], rows)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sheet %s: %w", sheetNames[0], err)
		}
		return []ExcelFile{sheet}, nil
	}

	// --- Case 2: Multiple sheets ---
	var excelFiles []ExcelFile
	for _, sheetName := range sheetNames {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			fmt.Printf("error reading sheet %s: %v\n", sheetName, err)
			continue
		}

		sheet, err := parseSheet(sheetName, rows)
		if err != nil {
			fmt.Printf("skipping sheet %s due to parse error: %v\n", sheetName, err)
			continue
		}

		excelFiles = append(excelFiles, sheet)
	}

	if len(excelFiles) == 0 {
		return nil, fmt.Errorf("no valid sheets could be parsed")
	}

	return excelFiles, nil
}

func ProcessMultiPartFile(c *gin.Context, formName string, allowedTypes []string) (multipart.File, *multipart.FileHeader, error) {
	file, fileHeader, err := c.Request.FormFile(formName)
	if err != nil {
		return nil, nil, err
	}

	if file == nil {
		return nil, nil, fmt.Errorf("file cannot be read")
	}

	if !isAllowed(fileHeader, allowedTypes) {
		return nil, nil, fmt.Errorf("file type not supported")
	}

	// Generate a new filename using UUID + original extension
	ext := filepath.Ext(fileHeader.Filename)
	fileName, err := uuids.NewUUID5(uuids.NewUUID(), uuids.DnsNamespace)
	if err != nil {
		return nil, nil, err
	}
	newFileName := fmt.Sprintf("%s%s", fileName, ext)

	// Update the filename in fileHeader
	fileHeader.Filename = newFileName

	return file, fileHeader, nil
}

func isAllowed(file *multipart.FileHeader, allowedTypes []string) bool {
	for _, allowedType := range allowedTypes {
		if file.Header.Get("Content-Type") == allowedType {
			return true
		}
	}
	return false
}

func IsAllowedType(mime string) bool {
	for _, t := range constants.AllowedTypes {
		if mime == t {
			return true
		}
	}
	return false
}

func CreateMultipartForm(file *os.File, fileName string) (*bytes.Buffer, string, error) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, "", err
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, "", err
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return &requestBody, writer.FormDataContentType(), nil
}

func WriteMultipartFile(data []byte) multipart.File {
	return nOpCloser{
		bytes.NewReader(data),
	}
}

// ValidateFileExtension checks if a filename has one of the allowed extensions (case-insensitive)
// Returns true if the file extension is valid, false otherwise
func ValidateFileExtension(filename string, allowedExtensions []string) bool {
	if len(filename) == 0 {
		return false
	}

	filenameLower := strings.ToLower(filename)
	for _, ext := range allowedExtensions {
		extLower := strings.ToLower(ext)
		if len(filenameLower) >= len(extLower) && filenameLower[len(filenameLower)-len(extLower):] == extLower {
			return true
		}
	}
	return false
}

// ValidateCSVOrExcelFile validates if a file is CSV or Excel format (case-insensitive)
// Returns true if valid, false otherwise
func ValidateCSVOrExcelFile(filename string) bool {
	csvExts := []string{".csv"}
	excelExts := []string{".xls", ".xlsx", ".xlsm", ".xlsb"}
	return ValidateFileExtension(filename, csvExts) || ValidateFileExtension(filename, excelExts)
}

// FileUploadResult represents the result of a file upload operation
type FileUploadResult struct {
	FileName  string // The renamed filename (with timestamp prefix)
	PublicURL string // The public URL from SeaweedFS
	Fid       string // The file ID from SeaweedFS
}

// ProcessAndUploadFile validates file extension, processes the file, and uploads to SeaweedFS
// Returns FileUploadResult with the upload details
func ProcessAndUploadFile(fileHeader *multipart.FileHeader, allowedExtensions []string) (*FileUploadResult, error) {

	filename := fileHeader.Filename

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Read into memory (once)
	originalBytes, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}

	trueMime := sniffMime(originalBytes)
	isImage := strings.HasPrefix(trueMime, "image/")

	var finalBytes []byte
	var uploadName string

	if isImage {
		// Convert to safe WEBP
		file := WriteMultipartFile(originalBytes)
		finalBytes, err = ConvertToWebPClean(file)
		if err != nil {
			return nil, err
		}

		uploadName = fmt.Sprintf("%d_%s.webp",
			time.Now().UnixNano(),
			strings.TrimSuffix(filename, filepath.Ext(filename)),
		)
	} else {
		// Extension validation for non-image files
		if !ValidateFileExtension(filename, allowedExtensions) {
			extList := strings.Join(allowedExtensions, ", ")
			return nil, fmt.Errorf("unsupported file extension: only %s allowed", extList)
		}

		finalBytes = originalBytes
		uploadName = fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)
	}

	// Convert back to multipart.File
	file := WriteMultipartFile(finalBytes)

	// Upload to SeaweedFS
	seaweedResult, err := seaweed.UploadToSeaweedFS(file, uploadName)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to SeaweedFS: %w", err)
	}

	return &FileUploadResult{
		FileName:  uploadName,
		PublicURL: seaweedResult.PublicURL,
		Fid:       seaweedResult.Fid,
	}, nil
}

// ProcessAndUploadCSVOrExcelFile is a convenience function for CSV/Excel files
func ProcessAndUploadCSVOrExcelFile(fileHeader *multipart.FileHeader) (*FileUploadResult, error) {
	allowedExts := []string{".csv", ".xls", ".xlsx", ".xlsm", ".xlsb"}
	return ProcessAndUploadFile(fileHeader, allowedExts)
}

func ConvertToWebPClean(src multipart.File) ([]byte, error) {
	imageEncodeSem <- struct{}{} // concurrency limit
	defer func() { <-imageEncodeSem }()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, src); err != nil {
		return nil, err
	}

	mime := sniffMime(buf.Bytes())
	if !strings.HasPrefix(mime, "image/") {
		return nil, fmt.Errorf("file is not an image")
	}

	img, _, err := image.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("invalid or corrupted image: %v", err)
	}

	if err := validateImageDimensions(img); err != nil {
		return nil, err
	}

	// Lossy WebP 80% quality (no metadata, EXIF destroyed)
	out, err := webp.EncodeRGBA(img, float32(80))
	if err != nil {
		return nil, fmt.Errorf("failed to encode webp: %w", err)
	}

	return out, nil
}

func sniffMime(data []byte) string {
	if len(data) < 512 {
		return http.DetectContentType(data)
	}
	return http.DetectContentType(data[:512])
}

func validateImageDimensions(img image.Image) error {
	bounds := img.Bounds()
	pixels := bounds.Dx() * bounds.Dy()

	if pixels > MaxPixels {
		return fmt.Errorf("image resolution too large")
	}
	return nil
}

func isImageMime(m string) bool {
	return strings.HasPrefix(m, "image/")
}
