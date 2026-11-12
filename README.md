# Audio Editor MVP

Простой онлайн-редактор аудио на **Go + FFmpeg + WaveSurfer.js**.

Загрузи файл → выдели фрагмент на waveform → обрежь, измени громкость, добавь fade, нормализуй → послушай результат и скачай.

## Возможности

- Drag-and-drop и выбор файла
- Поддержка mp3, wav, m4a, flac, ogg, aac, opus
- Интерактивный waveform с выделением региона (WaveSurfer.js)
- Обрезка по времени, громкость, Fade In/Out
- Нормализация громкости (loudnorm)
- Выбор формата: MP3 / WAV / OGG
- Превью результата в браузере
- Индикатор прогресса
- История последних файлов (localStorage)
- Адаптивный интерфейс
- Автоочистка старых файлов
- Docker из коробки

## Стек

| Часть     | Технологии                     |
|-----------|--------------------------------|
| Backend   | Go (net/http), FFmpeg          |
| Frontend  | HTML/CSS/JS, WaveSurfer.js     |
| Infra     | Docker, Docker Compose         |
| CI        | GitHub Actions                 |

## Структура

```
audio-editor-mvp/
├── backend/
│   ├── main.go
│   ├── go.mod
│   └── internal/
│       ├── audio/       # бизнес-логика (FFmpeg, файлы)
│       ├── config/      # константы
│       ├── handlers/    # HTTP-обработчики + тесты
│       └── response/    # JSON-ответы
├── static/
│   └── index.html
├── .github/workflows/ci.yml
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## Быстрый старт

```bash
# Docker
make docker
# или
docker compose up --build

# Локально (нужны Go и FFmpeg)
make run

# Тесты
make test

# Сборка бинарника
make build
```

Открой http://localhost:8080

## API

### `POST /api/upload`
`multipart/form-data`, поле `audio`.

```json
{ "ok": true, "filename": "123_track.mp3", "url": "/uploads/123_track.mp3" }
```

### `POST /api/process`
```json
{
  "filename": "123_track.mp3",
  "start_time": "0",
  "end_time": "00:01:30",
  "volume": 1.0,
  "fade_in": 0.5,
  "fade_out": 1.0,
  "format": "mp3",
  "normalize": false
}
```

### `GET /api/health`
```json
{ "ok": true, "message": "ok" }
```

## Заметки

- Файлы старше 1 часа периодически удаляются.
- Максимальный размер загрузки — 50 МБ.
- История хранится в localStorage (до 8 файлов).
- CI прогоняет тесты и сборку при каждом пуше в main.

---

Pet-проект / прототип онлайн-редактора аудио.
