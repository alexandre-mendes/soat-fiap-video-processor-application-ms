package controllers

import (
	"context"
	"net/http"
	"time"
	"video-processor/services"
	"video-processor/utils"

	"github.com/gin-gonic/gin"
)

var messageProcessor *services.MessageProcessor

// Inicializar o processador de mensagens (chamado no main.go)
func InitMessageProcessor() error {
	// Configuração para o processador de mensagens
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

	var err error
	messageProcessor, err = services.NewMessageProcessor(config)
	if err != nil {
		return err
	}

	// Iniciar processamento em background
	go messageProcessor.StartProcessing(context.Background())

	return nil
}

// Endpoint para processar uma mensagem específica (para testes)
func HandleProcessMessage(c *gin.Context) {
	var request struct {
		FileID    string `json:"fileId" binding:"required"`
		ProcessID string `json:"processId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Dados inválidos: " + err.Error(),
		})
		return
	}

	if messageProcessor == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "Processador de mensagens não inicializado",
		})
		return
	}

	// Simular mensagem SQS
	// Criar mensagem para processamento
	message := services.VideoProcessingMessage{
		FileID:    request.FileID,
		ProcessID: request.ProcessID,
	}

	// Baixar e processar
	ctx := context.Background()
	localPath, err := messageProcessor.DownloadFromS3(ctx, messageProcessor.GetSourceBucket(), message.FileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Erro ao baixar do S3: " + err.Error(),
		})
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	result := services.ProcessVideo(localPath, timestamp)

	c.JSON(http.StatusOK, result)
}

// Endpoint para status do processador de mensagens
func HandleMessageProcessorStatus(c *gin.Context) {
	status := gin.H{
		"processor_active": messageProcessor != nil,
		"sqs_queue":        utils.GetEnv("SQS_QUEUE_URL", "http://localhost:4566/000000000000/video-processing-queue"),
		"localstack_url":   utils.GetEnv("LOCALSTACK_URL", "http://localhost:4566"),
		"aws_region":       utils.GetEnv("AWS_REGION", "us-east-1"),
	}

	c.JSON(http.StatusOK, status)
}
