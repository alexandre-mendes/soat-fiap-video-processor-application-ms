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
