package main

import (
	"fmt"
	"kanastra-api/internal/infra/config"
	"kanastra-api/internal/setup"
	"log"
)

func main() {
	repo := setup.Repository()
	email, invoice := setup.Services()
	producer, consumer := setup.Kafka(repo)
	defer setup.CloseKafka(producer, consumer)

	useCase := setup.UseCase(repo, email, invoice, producer)
	router := setup.Routes(useCase)

	if err := router.Run(fmt.Sprintf(":%v", config.GetEnv("HTTP_PORT", "8084"))); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
