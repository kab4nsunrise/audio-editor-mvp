package config

import "time"

const (
	UploadDir   = "./uploads"
	MaxUploadMB = 50
	ListenAddr  = ":8080"
	FileMaxAge  = 1 * time.Hour
	CleanupEvery = 15 * time.Minute
)
