package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleHTML(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	HandleHTML(c)

	if w.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "text/html" {
		t.Errorf("Esperado Content-Type 'text/html', obtido '%s'", w.Header().Get("Content-Type"))
	}
	if w.Body.String() == "" {
		t.Error("Esperado corpo HTML não vazio")
	}
}

func TestHandleHTML_ContentType(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	HandleHTML(c)
	if w.Header().Get("Content-Type") != "text/html" {
		t.Errorf("Esperado Content-Type 'text/html', obtido '%s'", w.Header().Get("Content-Type"))
	}
	if w.Code != http.StatusOK {
		t.Errorf("Esperado status 200, obtido %d", w.Code)
	}
	if w.Body.String() == "" {
		t.Error("Esperado corpo HTML não vazio")
	}
}

func TestGetHTMLForm(t *testing.T) {
	html := GetHTMLForm()
	if html == "" {
		t.Error("Esperado HTML não vazio")
	}
	if !containsHTML(html, "<form") {
		t.Error("Esperado HTML de formulário")
	}
}

func TestGetHTMLForm_ContemTitulo(t *testing.T) {
	html := GetHTMLForm()
	if !containsHTML(html, "<title>FIAP X - Processador de Vídeos</title>") {
		t.Error("Esperado título FIAP X no HTML gerado")
	}
}

func TestHandleHTML_FormCompleto(t *testing.T) {
	html := GetHTMLForm()
	// Verifica se contém campos essenciais do formulário
	if !containsHTML(html, "<form id=\"uploadForm\"") {
		t.Error("Esperado form com id uploadForm")
	}
	if !containsHTML(html, "input type=\"file\"") {
		t.Error("Esperado campo input file no HTML")
	}
	if !containsHTML(html, "button type=\"submit\"") {
		t.Error("Esperado botão submit no HTML")
	}
}

func containsHTML(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[0:len(substr)] == substr || containsHTML(s[1:], substr)))
}
