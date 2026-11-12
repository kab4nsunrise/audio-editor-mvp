package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"audio-editor-mvp/internal/audio"
)

func setupTestHandler(t *testing.T) (*Handler, string) {
	t.Helper()
	dir := t.TempDir()
	svc := audio.NewService(dir)
	if err := svc.EnsureUploadDir(); err != nil {
		t.Fatal(err)
	}
	return New(svc), dir
}

func TestHealth(t *testing.T) {
	h, _ := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()
	h.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["ok"] != true {
		t.Fatalf("expected ok=true, got %v", body)
	}
}

func TestUpload_MethodNotAllowed(t *testing.T) {
	h, _ := setupTestHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/upload", nil)
	rr := httptest.NewRecorder()
	h.Upload(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestUpload_Success(t *testing.T) {
	h, dir := setupTestHandler(t)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("audio", "test.mp3")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(part, "ID3fake-audio-content-for-test"); err != nil {
		t.Fatal(err)
	}
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["ok"] != true {
		t.Fatalf("expected ok=true: %v", body)
	}
	filename, _ := body["filename"].(string)
	if filename == "" {
		t.Fatal("empty filename")
	}
	if _, err := os.Stat(filepath.Join(dir, filename)); err != nil {
		t.Fatalf("file not saved: %v", err)
	}
}

func TestUpload_BadExtension(t *testing.T) {
	h, _ := setupTestHandler(t)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, _ := w.CreateFormFile("audio", "virus.exe")
	io.WriteString(part, "not-audio")
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()
	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestProcess_MissingFilename(t *testing.T) {
	h, _ := setupTestHandler(t)
	body := strings.NewReader(`{"filename":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/process", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Process(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestProcess_FileNotFound(t *testing.T) {
	h, _ := setupTestHandler(t)
	body := strings.NewReader(`{"filename":"missing.mp3","start_time":"0","end_time":"5"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/process", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Process(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestProcess_BadTimeFormat(t *testing.T) {
	h, _ := setupTestHandler(t)
	body := strings.NewReader(`{"filename":"x.mp3","start_time":"not-a-time"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/process", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Process(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
