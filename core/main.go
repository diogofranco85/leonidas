package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"leonidas/core/internal/infrastructure/http"
)

func main() {
	// Criar servidor
	server := http.NewServer()

	// Canal para capturar sinais do sistema
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar servidor em goroutine
	go func() {
		log.Println("Iniciando servidor...")
		if err := server.Start(); err != nil {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	// Aguardar sinal de parada
	<-sigChan
	log.Println("Recebido sinal de parada, encerrando servidor...")

	// Parar servidor graciosamente
	if err := server.Shutdown(); err != nil {
		log.Printf("Erro ao parar servidor: %v", err)
	}

	log.Println("Servidor encerrado com sucesso")
}
