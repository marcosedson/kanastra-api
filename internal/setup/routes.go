package setup

import (
	"github.com/gin-gonic/gin"

	"kanastra-api/internal/core/usecase"
	"kanastra-api/internal/handler"
)

func Routes(useCase *usecase.ProcessFileUseCase) *gin.Engine {
	router := gin.Default()
	processFileHandler := handler.NewProcessFileHandler(useCase)
	processFileHandler.RegisterRoutes(router)

	return router
}
