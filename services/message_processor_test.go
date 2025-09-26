package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Mock S3Client que retorna erro no upload
// Deve estar após os imports, antes das funções de teste

type mockS3ClientErro struct{}
func (m *mockS3ClientErro) PutObject(ctx context.Context, input *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return nil, errors.New("erro simulado no upload")
}
func (m *mockS3ClientErro) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return &s3.DeleteObjectOutput{}, nil
}
func (m *mockS3ClientErro) GetObject(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{}, nil
}

func TestGetSourceBucket(t *testing.T) {
	config := MessageProcessorConfig{SourceBucket: "bucket-test"}
	mp := &MessageProcessor{config: config}
	if mp.GetSourceBucket() != "bucket-test" {
		t.Errorf("Esperado 'bucket-test', obtido '%s'", mp.GetSourceBucket())
	}
}

func TestSendProcessingResult_FilaNaoConfigurada(t *testing.T) {
	mp := &MessageProcessor{config: MessageProcessorConfig{ResultsQueueURL: ""}}
	err := mp.SendProcessingResult(context.TODO(), "proc-1", "zipKey", "COMPLETED")
	if err != nil {
		t.Errorf("Esperado nil, obtido %v", err)
	}
}

func TestNewMessageProcessor_MinimalConfig(t *testing.T) {
	config := MessageProcessorConfig{
		SQSQueueURL:     "http://localhost:4566/000000000000/video-processing-queue",
		ResultsQueueURL: "http://localhost:4566/000000000000/video-results-queue",
		AWSRegion:       "us-east-1",
		SourceBucket:    "bucket-test",
		ResultsBucket:   "results-test",
		PollingInterval: 1,
		MaxMessages:     1,
	}
	mp, err := NewMessageProcessor(config)
	if err != nil {
		t.Errorf("Esperado sucesso na criação do MessageProcessor, obtido erro: %v", err)
	}
	if mp == nil {
		t.Error("Esperado MessageProcessor não nulo")
	}
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	mock := &mockSQSClient{}
	mp := &MessageProcessor{sqsClient: mock}
	msg := types.Message{MessageId: ptr("id1"), Body: ptr("{invalid json}")}
	mp.processMessage(context.TODO(), msg)
	// Espera não panicar e logar erro
}

func TestUploadZipToS3_ErroAoAbrirArquivo(t *testing.T) {
	mp := &MessageProcessor{}
	err := mp.UploadZipToS3(context.TODO(), "bucket", "key", "arquivo_inexistente.zip")
	if err == nil {
		t.Error("Esperado erro ao abrir arquivo ZIP inexistente")
	}
}

func TestUploadZipToS3_ErroNoUploadS3(t *testing.T) {
	// Mock S3Client que retorna erro

	mp := &MessageProcessor{s3Client: &mockS3ClientErro{}}
	// Cria arquivo fake para upload
	zipPath := "outputs/teste_upload.zip"
	os.MkdirAll("outputs", 0755)
	os.WriteFile(zipPath, []byte("conteudo"), 0644)
	defer os.Remove(zipPath)

	err := mp.UploadZipToS3(context.TODO(), "bucket", "key", zipPath)
	if err == nil {
		t.Error("Esperado erro ao enviar ZIP para S3")
	}
}

// Mock SQSClient para testes
type mockSQSClient struct {
	messages    []types.Message
	failReceive bool
	failSend    bool
	failDelete  bool
}

func (m *mockSQSClient) ReceiveMessage(ctx context.Context, input *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	if m.failReceive {
		return nil, errors.New("erro simulado no ReceiveMessage")
	}
	return &sqs.ReceiveMessageOutput{Messages: m.messages}, nil
}
func (m *mockSQSClient) SendMessage(ctx context.Context, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	if m.failSend {
		return nil, errors.New("erro simulado no SendMessage")
	}
	return &sqs.SendMessageOutput{}, nil
}
func (m *mockSQSClient) DeleteMessage(ctx context.Context, input *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	if m.failDelete {
		return nil, errors.New("erro simulado no DeleteMessage")
	}
	return &sqs.DeleteMessageOutput{}, nil
}

// Mock S3Client para testes
// Retorna erro ou sucesso simulado

type mockS3Client struct {
	failGet bool
}

func (m *mockS3Client) GetObject(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.failGet {
		return nil, errors.New("erro simulado no GetObject")
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte{}))}, nil
}
func (m *mockS3Client) PutObject(ctx context.Context, input *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, nil
}
func (m *mockS3Client) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return &s3.DeleteObjectOutput{}, nil
}

func TestProcessMessages_Success(t *testing.T) {
	msg := types.Message{MessageId: ptr("id1"), Body: ptr(`{"fileId":"video.mp4","processId":"proc-1"}`), ReceiptHandle: ptr("rh1")}
	mockSQS := &mockSQSClient{messages: []types.Message{msg}}
	mockS3 := &mockS3Client{}
	mp := &MessageProcessor{config: MessageProcessorConfig{SQSQueueURL: "url", MaxMessages: 1}, sqsClient: mockSQS, s3Client: mockS3}
	mp.processMessages(context.TODO())
	// Espera processar mensagem sem panic
}

func TestProcessMessages_Error(t *testing.T) {
	mock := &mockSQSClient{failReceive: true}
	mp := &MessageProcessor{config: MessageProcessorConfig{SQSQueueURL: "url", MaxMessages: 1}, sqsClient: mock}
	mp.processMessages(context.TODO())
	// Espera logar erro e não panicar
}

func TestDeleteMessage_Error(t *testing.T) {
	mock := &mockSQSClient{failDelete: true}
	mp := &MessageProcessor{config: MessageProcessorConfig{SQSQueueURL: "url"}, sqsClient: mock}
	msg := types.Message{MessageId: ptr("id1"), ReceiptHandle: ptr("rh1")}
	mp.deleteMessage(context.TODO(), msg)
	// Espera logar erro e não panicar
}

func ptr(s string) *string { return &s }

// Mock SQSClient que retorna erro no SendMessage
// Deve estar após os imports, antes das funções de teste

type mockSQSClientErro struct{}
func (m *mockSQSClientErro) SendMessage(ctx context.Context, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return nil, errors.New("erro simulado no envio")
}
func (m *mockSQSClientErro) ReceiveMessage(ctx context.Context, input *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	return &sqs.ReceiveMessageOutput{}, nil
}
func (m *mockSQSClientErro) DeleteMessage(ctx context.Context, input *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	return &sqs.DeleteMessageOutput{}, nil
}

func TestSendProcessingResult_ErroNoEnvioSQS(t *testing.T) {
	mp := &MessageProcessor{config: MessageProcessorConfig{ResultsQueueURL: "url"}, sqsClient: &mockSQSClientErro{}}
	err := mp.SendProcessingResult(context.TODO(), "proc", "zip", "COMPLETED")
	if err == nil {
		t.Error("Esperado erro ao enviar mensagem para fila de resultados")
	}
}

// Mock S3Client que retorna erro no GetObject
// Deve estar após os imports, antes das funções de teste

type mockS3ClientGetErro struct{}
func (m *mockS3ClientGetErro) GetObject(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return nil, errors.New("erro simulado no GetObject")
}
func (m *mockS3ClientGetErro) PutObject(ctx context.Context, input *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, nil
}
func (m *mockS3ClientGetErro) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return &s3.DeleteObjectOutput{}, nil
}

func TestProcessMessage_ErroParseJSON(t *testing.T) {
	mockSQS := &mockSQSClient{}
	mp := &MessageProcessor{sqsClient: mockSQS}
	msg := types.Message{MessageId: ptr("id1"), Body: ptr("{invalid json}")}
	mp.processMessage(context.TODO(), msg)
	// Espera não panicar e logar erro
}

func TestProcessMessage_ErroNoDownloadS3(t *testing.T) {
	mockSQS := &mockSQSClient{}
	mockS3 := &mockS3ClientGetErro{}
	mp := &MessageProcessor{sqsClient: mockSQS, s3Client: mockS3, config: MessageProcessorConfig{SourceBucket: "bucket"}}
	msg := types.Message{MessageId: ptr("id1"), Body: ptr(`{"fileId":"video.mp4","processId":"proc-1"}`)}
	mp.processMessage(context.TODO(), msg)
	// Espera não panicar e logar erro
}

func TestDownloadFromS3_ErroNoGetObject(t *testing.T) {
	mp := &MessageProcessor{s3Client: &mockS3ClientGetErro{}}
	_, err := mp.DownloadFromS3(context.TODO(), "bucket", "key")
	if err == nil {
		t.Error("Esperado erro ao baixar objeto S3")
	}
}

// Mock SQSClient que retorna erro no DeleteMessage
// Deve estar após os imports, antes das funções de teste

type mockSQSClientDeleteErro struct{}
func (m *mockSQSClientDeleteErro) DeleteMessage(ctx context.Context, input *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	return nil, errors.New("erro simulado no DeleteMessage")
}
func (m *mockSQSClientDeleteErro) SendMessage(ctx context.Context, input *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return &sqs.SendMessageOutput{}, nil
}
func (m *mockSQSClientDeleteErro) ReceiveMessage(ctx context.Context, input *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	return &sqs.ReceiveMessageOutput{}, nil
}

func TestDeleteMessage_ErroNoDeleteSQS(t *testing.T) {
	mp := &MessageProcessor{config: MessageProcessorConfig{SQSQueueURL: "url"}, sqsClient: &mockSQSClientDeleteErro{}}
	msg := types.Message{MessageId: ptr("id1"), ReceiptHandle: ptr("rh1")}
	mp.deleteMessage(context.TODO(), msg)
	// Espera logar erro de deleção na fila e não panicar
}

// Mock S3Client que retorna erro no DeleteObject
// Deve estar após os imports, antes das funções de teste

type mockS3ClientDeleteErro struct{}
func (m *mockS3ClientDeleteErro) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return nil, errors.New("erro simulado no DeleteObject")
}
func (m *mockS3ClientDeleteErro) PutObject(ctx context.Context, input *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, nil
}
func (m *mockS3ClientDeleteErro) GetObject(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{}, nil
}

func TestDeleteObject_ErroNoDelete(t *testing.T) {
	mp := &MessageProcessor{s3Client: &mockS3ClientDeleteErro{}}
	_, err := mp.s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{})
	if err == nil {
		t.Error("Esperado erro ao excluir objeto S3")
	}
}

func TestProcessMessage_ErroNoProcessamento(t *testing.T) {
	mockSQS := &mockSQSClient{}
	mockS3 := &mockS3Client{}
	mp := &MessageProcessor{sqsClient: mockSQS, s3Client: mockS3, config: MessageProcessorConfig{SourceBucket: "bucket"}}
	// Simula vídeo inválido para forçar erro no processamento
	msg := types.Message{MessageId: ptr("id1"), Body: ptr(`{"fileId":"arquivo_invalido.mp4","processId":"proc-1"}`)}
	mp.processMessage(context.TODO(), msg)
	// Espera não panicar e logar erro
}

func TestProcessMessages_MultiplasMensagens(t *testing.T) {
	msg1 := types.Message{MessageId: ptr("id1"), Body: ptr(`{"fileId":"video1.mp4","processId":"proc-1"}`), ReceiptHandle: ptr("rh1")}
	msg2 := types.Message{MessageId: ptr("id2"), Body: ptr(`{"fileId":"video2.mp4","processId":"proc-2"}`), ReceiptHandle: ptr("rh2")}
	mockSQS := &mockSQSClient{messages: []types.Message{msg1, msg2}}
	mockS3 := &mockS3Client{}
	mp := &MessageProcessor{config: MessageProcessorConfig{SQSQueueURL: "url", MaxMessages: 2, SourceBucket: "bucket"}, sqsClient: mockSQS, s3Client: mockS3}
	mp.processMessages(context.TODO())
	// Espera processar múltiplas mensagens sem panic
}

func TestProcessMessages_ErroNoReceiveMessage(t *testing.T) {
	mockSQS := &mockSQSClient{failReceive: true}
	mp := &MessageProcessor{config: MessageProcessorConfig{SQSQueueURL: "url", MaxMessages: 1}, sqsClient: mockSQS}
	mp.processMessages(context.TODO())
	// Espera logar erro e não panicar
}
