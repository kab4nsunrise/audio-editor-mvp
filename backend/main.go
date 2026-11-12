package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"audio-editor-mvp/internal/audio"
	"audio-editor-mvp/internal/config"
	"audio-editor-mvp/internal/handlers"
)

func main() {
	svc := audio.NewService(config.UploadDir)
	if err := svc.EnsureUploadDir(); err != nil {
		log.Fatal(err)
	}
	svc.StartCleanupLoop(config.CleanupEvery, config.FileMaxAge)

	h := handlers.New(svc)
	staticDir := findStaticDir()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(config.UploadDir))))
	mux.HandleFunc("/api/upload", h.Upload)
	mux.HandleFunc("/api/process", h.Process)
	mux.HandleFunc("/api/health", h.Health)

	fmt.Println("Audio Editor MVP → http://localhost" + config.ListenAddr)
	log.Fatal(http.ListenAndServe(config.ListenAddr, handlers.WithCORS(mux)))
}

func findStaticDir() string {
	candidates := []string{"./static", "../static", "/app/static"}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	log.Println("warning: static dir not found, using ./static")
	return "./static"
}
