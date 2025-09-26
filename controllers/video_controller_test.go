package controllers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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
