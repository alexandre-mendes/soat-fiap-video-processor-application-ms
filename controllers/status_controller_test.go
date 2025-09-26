package controllers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func containsStatus(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || (len(s) > len(substr) && (s[0:len(substr)] == substr || containsStatus(s[1:], substr))))
}

func TestHandleStatus(t *testing.T) {
	filename := "test_status.zip"
	filePath := "outputs/" + filename
	_ = os.MkdirAll("outputs", 0755)
	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Erro ao criar arquivo de teste: %v", err)
	}
	f.Close()
	defer os.Remove(filePath)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	HandleStatus(c)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", w.Code)
	}
	if w.Body.String() == "" {
		t.Error("Esperado corpo não vazio")
	}
	if !containsStatus(w.Body.String(), filename) {
		t.Errorf("Resposta não contém o arquivo esperado: %s", filename)
	}
}

func TestHandleHealth(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	HandleHealth(c)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", w.Code)
	}
	if w.Body.String() == "" {
		t.Error("Esperado corpo não vazio")
	}
	if !containsStatus(w.Body.String(), "healthy") {
		t.Error("Resposta não contém 'healthy'")
	}
}
