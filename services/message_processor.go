package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// Estrutura da mensagem SQS esperada
// Exemplo: { "fileId": "videos/video.mp4", "processId": "proc-123" }
type VideoProcessingMessage struct {
	FileID    string `json:"fileId"`
	ProcessID string `json:"processId"`
	MessageID string `json:"message_id,omitempty"`
}

// Estrutura da mensagem de resultado
type VideoProcessingResult struct {
	ProcessID string `json:"processId"`
	ZipKey    string `json:"zipKey"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Configuração do processador de mensagens
type MessageProcessorConfig struct {
	SQSQueueURL     string
	ResultsQueueURL string
	LocalStackURL   string
	AWSRegion       string
	SourceBucket    string
	ResultsBucket   string
	PollingInterval time.Duration
	MaxMessages     int32
}

// Processador principal de mensagens
type MessageProcessor struct {
	config    MessageProcessorConfig
	sqsClient SQSClient
	s3Client  S3Client
}

// Criar novo processador de mensagens
func NewMessageProcessor(config MessageProcessorConfig) (*MessageProcessor, error) {
	cfg, err := config.loadAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar configuração AWS: %w", err)
	}

	var sqsClient *sqs.Client
	var s3Client *s3.Client
	if config.LocalStackURL != "" {
		sqsClient = sqs.NewFromConfig(cfg, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(config.LocalStackURL)
		})
		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(config.LocalStackURL)
			o.UsePathStyle = true // Necessário para LocalStack
		})
	} else {
		sqsClient = sqs.NewFromConfig(cfg)
		s3Client = s3.NewFromConfig(cfg)
	}

	return &MessageProcessor{
		config:    config,
		sqsClient: sqsClient,
		s3Client:  s3Client,
	}, nil
}

// Carregar configuração AWS
func (c MessageProcessorConfig) loadAWSConfig() (aws.Config, error) {
	return config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(c.AWSRegion),
	)
}

// Iniciar o loop de processamento de mensagens
func (mp *MessageProcessor) StartProcessing(ctx context.Context) {
	log.Printf("🚀 Iniciando processamento de mensagens SQS")
	log.Printf("📡 Queue: %s", mp.config.SQSQueueURL)
	log.Printf("⏱️  Intervalo: %s", mp.config.PollingInterval)

	ticker := time.NewTicker(mp.config.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Parando processamento de mensagens")
			return
		case <-ticker.C:
			mp.processMessages(ctx)
		}
	}
}

// Processar mensagens da fila SQS
func (mp *MessageProcessor) processMessages(ctx context.Context) {
	resp, err := mp.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(mp.config.SQSQueueURL),
		MaxNumberOfMessages: mp.config.MaxMessages,
		WaitTimeSeconds:     5, // Long polling
	})

	if err != nil {
		log.Printf("❌ Erro ao receber mensagens SQS: %v", err)
		return
	}

	if len(resp.Messages) == 0 {
		log.Println("📭 Nenhuma mensagem na fila")
		return
	}

	log.Printf("📨 Recebidas %d mensagem(s)", len(resp.Messages))

	for _, message := range resp.Messages {
		mp.processMessage(ctx, message)
	}
}

func (mp *MessageProcessor) processMessage(ctx context.Context, message types.Message) {
	log.Printf("🔄 Processando mensagem: %s", *message.MessageId)

	var videoMsg VideoProcessingMessage
	if err := json.Unmarshal([]byte(*message.Body), &videoMsg); err != nil {
		log.Printf("❌ Erro ao fazer parse da mensagem: %v", err)
		mp.deleteMessage(ctx, message)
		return
	}

	videoMsg.MessageID = *message.MessageId
	log.Printf("📹 Processando vídeo: s3://%s/%s (ProcessID: %s)", mp.config.SourceBucket, videoMsg.FileID, videoMsg.ProcessID)

	// Enviar notificação de início do processamento
	err := mp.SendProcessingResult(ctx, videoMsg.ProcessID, "", "IN_PROGRESS")
	if err != nil {
		log.Printf("⚠️ Erro ao enviar notificação de início: %v", err)
	}

	// Baixar arquivo do S3
	localPath, err := mp.DownloadFromS3(ctx, mp.config.SourceBucket, videoMsg.FileID)
	if err != nil {
		log.Printf("❌ Erro ao baixar do S3: %v", err)
		// Enviar notificação de erro
		mp.SendProcessingResult(ctx, videoMsg.ProcessID, "", "FAILED")
		return
	}
	defer os.Remove(localPath) // Limpar arquivo local após processamento

	// Processar vídeo
	timestamp := time.Now().Format("20060102_150405")
	result := ProcessVideo(localPath, timestamp)

	if result.Success {
		log.Printf("✅ Vídeo processado com sucesso: %s", result.ZipPath)

		// Upload do ZIP para S3 com ProcessID único
		zipS3Key := fmt.Sprintf("processed/%s_%s", videoMsg.ProcessID, result.ZipPath)
		localZipPath := filepath.Join("outputs", result.ZipPath)

		log.Printf("📤 Iniciando upload: %s → s3://%s/%s", localZipPath, mp.config.ResultsBucket, zipS3Key)

		err := mp.UploadZipToS3(ctx, mp.config.ResultsBucket, zipS3Key, localZipPath)
		if err != nil {
			log.Printf("❌ Erro ao enviar ZIP para S3: %v", err)
			// Enviar notificação de erro
			mp.SendProcessingResult(ctx, videoMsg.ProcessID, "", "FAILED")
			return
		}

		// Remover ZIP local após upload bem-sucedido
		if err := os.Remove(localZipPath); err != nil {
			log.Printf("⚠️ Aviso: Erro ao remover ZIP local: %v", err)
		} else {
			log.Printf("🗑️ ZIP local removido: %s", localZipPath)
		}

		// Excluir arquivo original do S3 após processamento
		_, err = mp.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(mp.config.SourceBucket),
			Key:    aws.String(videoMsg.FileID),
		})
		if err != nil {
			log.Printf("⚠️ Aviso: Erro ao excluir arquivo original do S3: %v", err)
		} else {
			log.Printf("🗑️ Arquivo original excluído do S3: s3://%s/%s", mp.config.SourceBucket, videoMsg.FileID)
		}

		// Deletar mensagem da fila após sucesso completo
		mp.deleteMessage(ctx, message)

		// Enviar resultado para fila de resultados
		err = mp.SendProcessingResult(ctx, videoMsg.ProcessID, zipS3Key, "COMPLETED")
		if err != nil {
			log.Printf("⚠️ Erro ao enviar notificação de resultado: %v", err)
		}
	} else {
		log.Printf("❌ Erro no processamento: %s", result.Message)
		// Enviar notificação de erro
		mp.SendProcessingResult(ctx, videoMsg.ProcessID, "", "FAILED")
		// Mensagem voltará para a fila após visibility timeout
	}
}

// Baixar arquivo do S3
func (mp *MessageProcessor) DownloadFromS3(ctx context.Context, bucket, key string) (string, error) {
	log.Printf("⬇️  Baixando s3://%s/%s", bucket, key)

	resp, err := mp.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("erro ao baixar objeto S3: %w", err)
	}
	defer resp.Body.Close()

	// Criar arquivo local
	filename := filepath.Base(key)
	localPath := filepath.Join("uploads", fmt.Sprintf("sqs_%d_%s", time.Now().Unix(), filename))

	// Garantir que o diretório existe
	os.MkdirAll("uploads", 0755)

	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("erro ao criar arquivo local: %w", err)
	}
	defer file.Close()

	// Copiar dados do S3 para arquivo local
	_, err = file.ReadFrom(resp.Body)
	if err != nil {
		os.Remove(localPath)
		return "", fmt.Errorf("erro ao salvar arquivo: %w", err)
	}

	log.Printf("📁 Arquivo salvo em: %s", localPath)
	return localPath, nil
}

// Upload do ZIP processado para S3
func (mp *MessageProcessor) UploadZipToS3(ctx context.Context, bucket, key, localZipPath string) error {
	log.Printf("📤 Enviando ZIP para S3: s3://%s/%s", bucket, key)

	// Abrir arquivo ZIP local
	file, err := os.Open(localZipPath)
	if err != nil {
		return fmt.Errorf("erro ao abrir ZIP local: %w", err)
	}
	defer file.Close()

	// Upload para S3
	_, err = mp.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("erro ao enviar ZIP para S3: %w", err)
	}

	log.Printf("✅ ZIP enviado com sucesso: s3://%s/%s", bucket, key)
	return nil
}

func (mp *MessageProcessor) deleteMessage(ctx context.Context, message types.Message) {
	_, err := mp.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(mp.config.SQSQueueURL),
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		log.Printf("❌ Erro ao deletar mensagem: %v", err)
	} else {
		log.Printf("🗑️  Mensagem deletada: %s", *message.MessageId)
	}
}

// GetSourceBucket retorna o bucket configurado para origem
func (mp *MessageProcessor) GetSourceBucket() string {
	return mp.config.SourceBucket
}

// Enviar resultado do processamento para fila de resultados
func (mp *MessageProcessor) SendProcessingResult(ctx context.Context, processID, zipKey, status string) error {
	if mp.config.ResultsQueueURL == "" {
		log.Printf("⚠️ Fila de resultados não configurada, pulando notificação")
		return nil
	}

	result := VideoProcessingResult{
		ProcessID: processID,
		ZipKey:    zipKey,
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("erro ao serializar resultado: %w", err)
	}

	log.Printf("📨 Enviando resultado para fila: %s", string(resultJSON))

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(mp.config.ResultsQueueURL),
		MessageBody: aws.String(string(resultJSON)),
	}
	// Adiciona MessageGroupId se a fila for FIFO
	if len(mp.config.ResultsQueueURL) > 5 && mp.config.ResultsQueueURL[len(mp.config.ResultsQueueURL)-5:] == ".fifo" {
		input.MessageGroupId = aws.String("soat-fiap-x-group")
	}

	_, err = mp.sqsClient.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem para fila de resultados: %w", err)
	}

	log.Printf("✅ Resultado enviado com sucesso para fila de resultados")
	return nil
}

// Interfaces para facilitar mocks nos testes
// Mantém compatibilidade com os clients originais
type SQSClient interface {
	ReceiveMessage(ctx context.Context, input *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	SendMessage(ctx context.Context, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	DeleteMessage(ctx context.Context, input *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type S3Client interface {
	GetObject(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, input *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}
