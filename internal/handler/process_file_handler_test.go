package handler

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockUseCase struct{}

func (m *MockUseCase) ProcessFileAsync(_ io.Reader, fileName string) int {
	if fileName == "error.csv" {
		return 0
	}

	return 100
}

func TestProcessFileHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUseCase := &MockUseCase{}
	processFileHandler := NewProcessFileHandler(mockUseCase)

	router := gin.Default()
	processFileHandler.RegisterRoutes(router)

	t.Run("Success with valid CSV", func(t *testing.T) {
		fileContent := `name,governmentId,email,debtAmount,debtDueDate,debtId\nJohn Doe,1234567890,john@example.com,1000.50,2025-01-01,abc123`
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("files", "valid.csv")
		assert.NoError(t, err)
		_, err = io.Copy(part, strings.NewReader(fileContent))
		assert.NoError(t, err)
		err = writer.Close()
		if err != nil {
			return
		}

		req := httptest.NewRequest(http.MethodPost, "/process-files", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusAccepted, resp.Code)
		assert.Contains(t, resp.Body.String(), "Files are being processed")
	})

	t.Run("Fail to parse multipart form", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/process-files", nil)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "Failed to parse form")
	})

	t.Run("No files provided", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		err := writer.Close()
		if err != nil {
			return
		}

		req := httptest.NewRequest(http.MethodPost, "/process-files", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "No files provided")
	})

	t.Run("CSV validation error", func(t *testing.T) {
		fileContent := `invalid-content`
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("files", "error.csv")
		assert.NoError(t, err)
		_, err = io.Copy(part, strings.NewReader(fileContent))
		assert.NoError(t, err)
		err = writer.Close()
		if err != nil {
			return
		}

		req := httptest.NewRequest(http.MethodPost, "/process-files", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusAccepted, resp.Code)
		assert.Contains(t, resp.Body.String(), "Files are being processed")
	})
}

func TestIsValidCSV(t *testing.T) {
	t.Run("Valid CSV", func(t *testing.T) {
		fileContent := `name,governmentId,email,debtAmount,debtDueDate,debtId
John Doe,1234567890,john@example.com,1000.50,2025-01-01,abc123`
		r := strings.NewReader(fileContent)
		err := IsValidCSV("valid.csv", r, expectedHeadersFile)

		assert.NoError(t, err)
	})

	t.Run("Invalid CSV extension", func(t *testing.T) {
		r := strings.NewReader(``)
		err := IsValidCSV("invalid.txt", r, expectedHeadersFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "arquivo não é um CSV")
	})

	t.Run("Empty CSV file", func(t *testing.T) {
		r := strings.NewReader(``)
		err := IsValidCSV("empty.csv", r, expectedHeadersFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "arquivo CSV está vazio")
	})

	t.Run("Invalid headers", func(t *testing.T) {
		fileContent := `wrong,header,format\n`
		r := strings.NewReader(fileContent)
		err := IsValidCSV("invalid.csv", r, expectedHeadersFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cabeçalho do CSV é inválido ou não corresponde ao esperado")
	})
}
