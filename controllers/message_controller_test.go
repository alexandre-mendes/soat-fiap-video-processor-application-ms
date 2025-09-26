package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"video-processor/services"

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

func TestHandleMessageProcessorStatus_Ativo(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	messageProcessor = &services.MessageProcessor{}
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
