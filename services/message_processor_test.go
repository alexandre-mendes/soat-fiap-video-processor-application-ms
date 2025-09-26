package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

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
