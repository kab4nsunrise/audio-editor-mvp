package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"regexp"

	"audio-editor-mvp/internal/audio"
	"audio-editor-mvp/internal/config"
	"audio-editor-mvp/internal/response"
)

var timeRe = regexp.MustCompile(`^(\d{1,2}:)?\d{1,2}:\d{2}(\.\d+)?$|^\d+(\.\d+)?$`)

type Handler struct {
	Audio *audio.Service
}

func New(svc *audio.Service) *Handler {
	return &Handler{Audio: svc}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response.OK(w, response.Body{Message: "ok"})
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, config.MaxUploadMB<<20)

	file, header, err := r.FormFile("audio")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Не удалось получить файл. Проверьте размер (макс. 50 МБ) и формат.")
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if !audio.AllowedExt(ext) {
		response.Error(w, http.StatusBadRequest, "Неподдерживаемый формат. Разрешены: mp3, wav, m4a, aac, flac, ogg, opus")
		return
	}

	filename, url, err := h.Audio.SaveUpload(file, header.Filename)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Ошибка сохранения файла на сервере")
		return
	}

	response.OK(w, response.Body{Filename: filename, URL: url})
}

type processRequest struct {
	Filename  string  `json:"filename"`
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Volume    float64 `json:"volume"`
	FadeIn    float64 `json:"fade_in"`
	FadeOut   float64 `json:"fade_out"`
	Format    string  `json:"format"`
	Normalize bool    `json:"normalize"`
}

func (h *Handler) Process(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	var req processRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Неверный JSON")
		return
	}

	if req.Filename == "" {
		response.Error(w, http.StatusBadRequest, "Не указан файл")
		return
	}
	if req.StartTime != "" && !timeRe.MatchString(req.StartTime) {
		response.Error(w, http.StatusBadRequest, "Неверный формат времени начала (пример: 00:00:05 или 5)")
		return
	}
	if req.EndTime != "" && !timeRe.MatchString(req.EndTime) {
		response.Error(w, http.StatusBadRequest, "Неверный формат времени конца")
		return
	}

	outName, url, err := h.Audio.Process(audio.ProcessParams{
		Filename:  req.Filename,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Volume:    req.Volume,
		FadeIn:    req.FadeIn,
		FadeOut:   req.FadeOut,
		Format:    req.Format,
		Normalize: req.Normalize,
	})
	if err != nil {
		if err.Error() == "file not found" {
			response.Error(w, http.StatusNotFound, "Файл не найден. Загрузите заново.")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Ошибка обработки аудио. Попробуйте другой файл или другие параметры.")
		return
	}

	response.OK(w, response.Body{Filename: outName, URL: url})
}

// CORS middleware
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
