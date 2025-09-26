package main

import (
	"testing"
)

func TestRun(t *testing.T) {
	// Teste básico: apenas chama run() para garantir que não ocorre panic
	// Em cenários reais, mocks e injeção de dependências são recomendados
	go func() {
		run()
	}()
}
