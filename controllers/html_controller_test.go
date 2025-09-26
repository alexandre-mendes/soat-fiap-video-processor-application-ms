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

func TestGetHTMLForm(t *testing.T) {
   html := GetHTMLForm()
   if html == "" {
	   t.Error("Esperado HTML não vazio")
   }
   if !containsHTML(html, "<form") {
	   t.Error("Esperado HTML de formulário")
   }
}

func containsHTML(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[0:len(substr)] == substr || containsHTML(s[1:], substr)))
}
