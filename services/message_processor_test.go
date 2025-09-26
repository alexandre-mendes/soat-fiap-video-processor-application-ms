package services

import (
	"context"
	"testing"
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
