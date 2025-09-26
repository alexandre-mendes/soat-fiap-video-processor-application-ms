package controllers

import (
	"errors"
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

func TestHandleStatus_ErroAoListarArquivos(t *testing.T) {
	// Simula erro no Glob usando diretório inexistente
	origGlob := globZipFiles
	globZipFiles = func(pattern string) ([]string, error) {
		return nil, errors.New("erro simulado")
	}
	defer func() { globZipFiles = origGlob }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	HandleStatus(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Esperado status 500, obtido %d", w.Code)
	}
	if !containsStatus(w.Body.String(), "Erro ao listar arquivos") {
		t.Error("Esperado mensagem de erro ao listar arquivos")
	}
}

func TestHandleHealth_ConteudoCompleto(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	HandleHealth(c)
	resp := w.Body.String()
	if w.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", w.Code)
	}
	if !containsStatus(resp, "healthy") {
		t.Error("Esperado status 'healthy' na resposta")
	}
	if !containsStatus(resp, "Service is running") {
		t.Error("Esperado mensagem de serviço rodando")
	}
	if !containsStatus(resp, "timestamp") {
		t.Error("Esperado campo timestamp na resposta")
	}
	if !containsStatus(resp, "version") {
		t.Error("Esperado campo version na resposta")
	}
}

func TestHandleStatus_SemArquivos(t *testing.T) {
	os.RemoveAll("outputs")
	os.MkdirAll("outputs", 0755)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	HandleStatus(c)
	if w.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", w.Code)
	}
	if !containsStatus(w.Body.String(), "files") {
		t.Error("Esperado campo 'files' na resposta")
	}
	if !containsStatus(w.Body.String(), "total") {
		t.Error("Esperado campo 'total' na resposta")
	}
}

func TestHandleStatus_ArquivoComErroDeStat(t *testing.T) {
	os.MkdirAll("outputs", 0755)
	file := "outputs/test_erro_stat.zip"
	os.WriteFile(file, []byte("conteudo"), 0644)
	// Remove permissão do arquivo para forçar erro no Stat
	os.Chmod(file, 0000)
	defer func() {
		os.Chmod(file, 0644)
		os.Remove(file)
	}()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	HandleStatus(c)
	if w.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", w.Code)
	}
	if !containsStatus(w.Body.String(), "files") {
		t.Error("Esperado campo 'files' na resposta mesmo com erro de Stat")
	}
}
