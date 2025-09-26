package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"context"
	"github.com/gin-gonic/gin"
)

func TestHandleProcessMessage_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/process-message", bytes.NewBuffer([]byte(`{"fileId":""}`)))
	c.Request.Header.Set("Content-Type", "application/json")

	HandleProcessMessage(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Esperado status 400, obtido %d", w.Code)
	}
}

func TestHandleProcessMessage_ServiceUnavailable(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/process-message", bytes.NewBuffer([]byte(`{"fileId":"abc","processId":"123"}`)))
	c.Request.Header.Set("Content-Type", "application/json")

	// Garante que messageProcessor está nil
	messageProcessor = nil
	HandleProcessMessage(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Esperado status 503, obtido %d", w.Code)
	}
}

func TestHandleProcessMessage_DadosInvalidos(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/process-message", bytes.NewBuffer([]byte(`{"fileId":123}`)))
	c.Request.Header.Set("Content-Type", "application/json")

	HandleProcessMessage(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Esperado status 400, obtido %d", w.Code)
	}
	if w.Body.String() == "" {
		t.Error("Esperado mensagem de erro para dados inválidos")
	}
}


// Mock para MessageProcessor
type mockProcessor struct{}
func (m *mockProcessor) DownloadFromS3(ctx context.Context, bucket, key string) (string, error) {
	return "video.mp4", nil
}
func (m *mockProcessor) GetSourceBucket() string { return "bucket" }
func (m *mockProcessor) StartProcessing(ctx context.Context) {}
func (m *mockProcessor) UploadZipToS3(ctx context.Context, bucket, key, localZipPath string) error { return nil }
func (m *mockProcessor) SendProcessingResult(ctx context.Context, processID, zipKey, status string) error { return nil }

func TestHandleMessageProcessorStatus_Ativo(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	messageProcessor = &mockProcessor{}
	HandleMessageProcessorStatus(c)
	if w.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", w.Code)
	}
	if w.Body.String() == "" {
		t.Error("Esperado corpo não vazio para status ativo")
	}
	if !containsStatus(w.Body.String(), "processor_active") {
		t.Error("Esperado campo processor_active na resposta")
	}
}
func TestHandleProcessMessage_Sucesso(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/process-message", bytes.NewBuffer([]byte(`{"fileId":"video.mp4","processId":"proc-1"}`)))
	c.Request.Header.Set("Content-Type", "application/json")

	messageProcessor = &mockProcessor{}

	HandleProcessMessage(c)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", w.Code)
	}
	if w.Body.String() == "" {
		t.Error("Esperado corpo não vazio para sucesso")
	}
}
