package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"video-processor/services"
	"video-processor/utils"
)

func main() {
	// Carregar variáveis do arquivo .env
	if err := utils.LoadEnv(".env"); err != nil {
		log.Printf("⚠️ Aviso: Não foi possível carregar .env: %v", err)
		log.Println("📄 Usando apenas variáveis de ambiente do sistema")
	} else {
		log.Println("✅ Arquivo .env carregado com sucesso")
	}

	// Configuração do processador de mensagens usando .env
	config := services.MessageProcessorConfig{
		SQSQueueURL:     utils.GetEnv("SQS_QUEUE_URL", "http://localhost:4566/000000000000/video-processing-queue"),
		ResultsQueueURL: utils.GetEnv("RESULTS_QUEUE_URL", "http://localhost:4566/000000000000/video-results-queue"),
		LocalStackURL:   utils.GetEnv("LOCALSTACK_URL", "http://localhost:4566"),
		AWSRegion:       utils.GetEnv("AWS_REGION", "us-east-1"),
		SourceBucket:    utils.GetEnv("SOURCE_BUCKET", "video-bucket"),
		ResultsBucket:   utils.GetEnv("RESULTS_BUCKET", "video-results"),
		PollingInterval: utils.GetEnvDuration("POLLING_INTERVAL_SECONDS", 5*time.Second),
		MaxMessages:     int32(utils.GetEnvInt("MAX_MESSAGES", 10)),
	}

	// Criar processador
	processor, err := services.NewMessageProcessor(config)
	if err != nil {
		log.Fatalf("❌ Erro ao criar processador: %v", err)
	}

	// Criar contexto para controle de cancelamento
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capturar sinais do sistema para shutdown graceful
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar processamento em goroutine
	go processor.StartProcessing(ctx)

	// Aguardar sinal de shutdown
	<-sigChan
	log.Println("🛑 Recebido sinal de shutdown, parando aplicação...")
	cancel()

	// Aguardar um tempo para finalização graceful
	time.Sleep(2 * time.Second)
	log.Println("👋 Aplicação finalizada")
}
