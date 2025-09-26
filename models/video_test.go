package models

import (
	"reflect"
	"testing"
)

func TestVideoRequestFields(t *testing.T) {
	vr := VideoRequest{VideoPath: "video.mp4", OutputDir: "out"}
	if vr.VideoPath != "video.mp4" {
		t.Errorf("Esperado 'video.mp4', obtido '%s'", vr.VideoPath)
	}
	if vr.OutputDir != "out" {
		t.Errorf("Esperado 'out', obtido '%s'", vr.OutputDir)
	}
}

func TestProcessingResultFields(t *testing.T) {
	pr := ProcessingResult{
		Success:    true,
		Message:    "ok",
		ZipPath:    "result.zip",
		FrameCount: 42,
		Images:     []string{"img1.jpg", "img2.jpg"},
	}
	if !pr.Success {
		t.Error("Esperado Success true")
	}
	if pr.Message != "ok" {
		t.Errorf("Esperado Message 'ok', obtido '%s'", pr.Message)
	}
	if pr.ZipPath != "result.zip" {
		t.Errorf("Esperado ZipPath 'result.zip', obtido '%s'", pr.ZipPath)
	}
	if pr.FrameCount != 42 {
		t.Errorf("Esperado FrameCount 42, obtido %d", pr.FrameCount)
	}
	if !reflect.DeepEqual(pr.Images, []string{"img1.jpg", "img2.jpg"}) {
		t.Errorf("Esperado Images ['img1.jpg', 'img2.jpg'], obtido %v", pr.Images)
	}
}
