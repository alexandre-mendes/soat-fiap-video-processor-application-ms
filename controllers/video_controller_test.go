package controllers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleVideoUpload_BadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/upload", nil)

	HandleVideoUpload(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Esperado status 400, obtido %d", w.Code)
	}
}

func TestHandleVideoUpload_InvalidFormat(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "arquivo.txt")
	part.Write([]byte("conteudo"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	HandleVideoUpload(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Esperado status 400 para formato inválido, obtido %d", w.Code)
	}
}

func TestHandleVideoUpload_InternalError(t *testing.T) {
	// Simula erro ao criar arquivo
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "video.mp4")
	part.Write([]byte("conteudo"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Simula diretório outputs inexistente para forçar erro
	os.RemoveAll("uploads")

	HandleVideoUpload(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Esperado status 500, obtido %d", w.Code)
	}
	if w.Body.String() == "" {
		t.Error("Esperado mensagem de erro ao salvar arquivo")
	}
}

func TestHandleVideoUpload_ErroAoCopiarArquivo(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("video", "video.mp4")
	part.Write([]byte("conteudo"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Simula erro ao copiar arquivo (arquivo já fechado)
	os.RemoveAll("uploads")
	os.MkdirAll("uploads", 0755)
	// Remove permissão de escrita
	os.Chmod("uploads", 0555)
	defer os.Chmod("uploads", 0755)

	HandleVideoUpload(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Esperado status 500, obtido %d", w.Code)
	}
	if w.Body.String() == "" {
		t.Error("Esperado mensagem de erro ao salvar arquivo")
	}
}
