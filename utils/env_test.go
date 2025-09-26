package utils

import (
	"os"
	"testing"
	"time"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_ENV", "valor")
	if GetEnv("TEST_ENV", "padrao") != "valor" {
		t.Errorf("Esperado 'valor', obtido '%s'", GetEnv("TEST_ENV", "padrao"))
	}
	os.Unsetenv("TEST_ENV")
	if GetEnv("TEST_ENV", "padrao") != "padrao" {
		t.Errorf("Esperado 'padrao', obtido '%s'", GetEnv("TEST_ENV", "padrao"))
	}
}

func TestGetEnvBool(t *testing.T) {
	os.Setenv("BOOL_ENV", "true")
	if !GetEnvBool("BOOL_ENV", false) {
		t.Error("Esperado true")
	}
	os.Setenv("BOOL_ENV", "false")
	if GetEnvBool("BOOL_ENV", true) {
		t.Error("Esperado false")
	}
	os.Unsetenv("BOOL_ENV")
	if !GetEnvBool("BOOL_ENV", true) {
		t.Error("Esperado default true")
	}
}

func TestGetEnvBool_Invalid(t *testing.T) {
	os.Setenv("BOOL_ENV", "talvez")
	if !GetEnvBool("BOOL_ENV", true) {
		t.Error("Esperado default true para valor inválido")
	}
	if GetEnvBool("BOOL_ENV", false) {
		t.Error("Esperado default false para valor inválido")
	}
	os.Unsetenv("BOOL_ENV")
}

func TestGetEnvInt(t *testing.T) {
	os.Setenv("INT_ENV", "42")
	if GetEnvInt("INT_ENV", 1) != 42 {
		t.Errorf("Esperado 42, obtido %d", GetEnvInt("INT_ENV", 1))
	}
	os.Unsetenv("INT_ENV")
	if GetEnvInt("INT_ENV", 7) != 7 {
		t.Errorf("Esperado default 7, obtido %d", GetEnvInt("INT_ENV", 7))
	}
}

func TestGetEnvInt_Invalid(t *testing.T) {
	os.Setenv("INT_ENV", "abc")
	if GetEnvInt("INT_ENV", 99) != 99 {
		t.Errorf("Esperado default 99 para valor inválido, obtido %d", GetEnvInt("INT_ENV", 99))
	}
	os.Unsetenv("INT_ENV")
}

func TestGetEnvDuration(t *testing.T) {
	os.Setenv("DUR_ENV", "2s")
	if GetEnvDuration("DUR_ENV", time.Second) != 2*time.Second {
		t.Errorf("Esperado 2s, obtido %v", GetEnvDuration("DUR_ENV", time.Second))
	}
	os.Unsetenv("DUR_ENV")
	if GetEnvDuration("DUR_ENV", 3*time.Second) != 3*time.Second {
		t.Errorf("Esperado default 3s, obtido %v", GetEnvDuration("DUR_ENV", 3*time.Second))
	}
}

func TestGetEnvDuration_Invalid(t *testing.T) {
	os.Setenv("DUR_ENV", "abc")
	if GetEnvDuration("DUR_ENV", 5*time.Second) != 5*time.Second {
		t.Errorf("Esperado default 5s para valor inválido, obtido %v", GetEnvDuration("DUR_ENV", 5*time.Second))
	}
	os.Unsetenv("DUR_ENV")
}

func TestLoadEnv_FileNotFound(t *testing.T) {
	err := LoadEnv("arquivo_inexistente.env")
	if err == nil {
		t.Error("Esperado erro ao carregar arquivo inexistente")
	}
}

func TestLoadEnv_FileComConteudo(t *testing.T) {
	file := "test.env"
	conteudo := "TESTE_ENV=valor\nOUTRO_ENV=123"
	os.WriteFile(file, []byte(conteudo), 0644)
	defer os.Remove(file)
	os.Unsetenv("TESTE_ENV")
	os.Unsetenv("OUTRO_ENV")
	if err := LoadEnv(file); err != nil {
		t.Errorf("Erro ao carregar env: %v", err)
	}
	if os.Getenv("TESTE_ENV") != "valor" {
		t.Errorf("Esperado 'valor', obtido '%s'", os.Getenv("TESTE_ENV"))
	}
	if os.Getenv("OUTRO_ENV") != "123" {
		t.Errorf("Esperado '123', obtido '%s'", os.Getenv("OUTRO_ENV"))
	}
}
