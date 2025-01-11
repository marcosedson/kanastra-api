package handler

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"kanastra-api/internal/handler/dto"
)

type ProcessFileUseCaseInterface interface {
	ProcessFileAsync(file io.Reader, fileName string) int
}

type ProcessFileHandler struct {
	useCase ProcessFileUseCaseInterface
}

var expectedHeadersFile = []string{"name", "governmentId", "email", "debtAmount", "debtDueDate", "debtId"}

func NewProcessFileHandler(useCase ProcessFileUseCaseInterface) *ProcessFileHandler {
	return &ProcessFileHandler{useCase: useCase}
}

func (h *ProcessFileHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/process-files", h.Handle)
}

func (h *ProcessFileHandler) Handle(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		c.JSON(http.StatusBadRequest, dto.ProcessFilesResponse{
			Message: "Failed to parse form",
		})

		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		log.Printf("No files provided")
		c.JSON(http.StatusBadRequest, dto.ProcessFilesResponse{
			Message: "No files provided",
		})

		return
	}

	for _, fileHeader := range files {
		go func(fileHeader *multipart.FileHeader) {
			file, err := fileHeader.Open()
			if err != nil {
				log.Printf("Failed to open file: %v", err)

				return
			}

			defer func(file multipart.File) {
				err := file.Close()
				if err != nil {

				}
			}(file)

			err = IsValidCSV(fileHeader.Filename, file, expectedHeadersFile)
			if err != nil {
				log.Printf("Arquivo CSV inválido: %v", err)
			}

			totalLines := h.useCase.ProcessFileAsync(file, fileHeader.Filename)
			log.Printf("Arquivo %s processado: Total de linhas: %d", fileHeader.Filename, totalLines)
		}(fileHeader)
	}

	c.JSON(http.StatusAccepted, dto.ProcessFilesResponse{
		Message: "Files are being processed",
	})
}

func IsValidCSV(fileName string, file io.Reader, expectedHeaders []string) error {
	if !strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		return errors.New("arquivo não é um CSV")
	}

	reader := csv.NewReader(file)

	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return errors.New("arquivo CSV está vazio")
		}

		return errors.New("erro ao ler o cabeçalho do CSV")
	}

	if len(expectedHeaders) > 0 {
		if !compareHeaders(headers, expectedHeaders) {
			return errors.New("cabeçalho do CSV é inválido ou não corresponde ao esperado")
		}
	}

	log.Printf("Arquivo CSV %s validado com sucesso", fileName)

	return nil
}

func compareHeaders(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i := range actual {
		if strings.TrimSpace(strings.ToLower(actual[i])) != strings.TrimSpace(strings.ToLower(expected[i])) {
			return false
		}
	}

	return true
}
