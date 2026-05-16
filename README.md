# TwinSpeak MVP - Real-Time Transcription

A real-time audio transcription application using WhisperLive (GPU-accelerated) with a React frontend and Go backend. Audio is captured in the browser, proxied through a Go WebSocket server, and transcribed in real-time using WhisperLive.

## Architecture

```
Browser (React)
    ↓ WS /ws/transcribe
Go Backend (Proxy)
    ↓ WS ws://whisper-live:9090
WhisperLive (GPU, Docker)
```

- **Frontend**: React 18 + Vite + TypeScript + Tailwind CSS + shadcn/ui
- **Backend**: Go with Gorilla WebSocket (stateless proxy)
- **WhisperLive**: GPU-accelerated speech-to-text using OpenAI's Whisper model

## Prerequisites

- Docker & Docker Compose
- NVIDIA Docker runtime (for GPU support)
- (Optional) Node.js 20+, Go 1.22+, pnpm for local development

## Quick Start

### With Docker Compose (Recommended)

```bash
docker compose up
```

This will:
1. Start WhisperLive (downloads the `turbo` model on first run, ~1-2 minutes)
2. Start the Go backend with hot-reload
3. Start the frontend dev server with pnpm

Open http://localhost:5173 in your browser.

### Manual Setup (Other Linux/macOS)

**Backend:**
```bash
cd backend
go install github.com/cosmtrek/air@latest
air -c .air.toml
```

**Frontend:**
```bash
cd frontend
pnpm install
pnpm dev
```

**WhisperLive:**
```bash
docker run --gpus all -p 9090:9090 ghcr.io/collabora/whisperlive-gpu:latest
```

## Usage

1. **Open the app**: http://localhost:5173
2. **Click "Start"**: Grants microphone permission and connects to the transcription service
3. **Speak**: Your speech appears in real-time in the transcript area
4. **Click "Stop"**: Ends the session cleanly

## Project Structure

```
twinspeak/
├── docker-compose.yml           # Service orchestration
├── backend/                     # Go WebSocket proxy
│   ├── Dockerfile
│   ├── .air.toml               # Hot-reload config
│   ├── go.mod
│   ├── main.go      # Entry point
│   └── internal/proxy/handler.go # WebSocket logic
├── frontend/                    # React + Vite app
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   ├── components.json         # shadcn config
│   ├── index.html
│   ├── public/
│   │   └── audio-processor.js  # AudioWorklet (MUST be static)
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── index.css           # Tailwind + CSS variables
│       ├── lib/utils.ts
│       ├── types/transcription.ts
│       ├── hooks/useTranscription.ts
│       └── components/
│           ├── TranscriptView.tsx
│           └── ui/             # shadcn components
└── services/whisper-live/      # (placeholder for local setup)
```

## Key Implementation Details

### Audio Processing (`public/audio-processor.js`)
- Custom AudioWorklet for efficient audio processing
- Resamples from browser sample rate to 16kHz using linear interpolation
- Converts float32 to int16 PCM encoding
- Sends 100ms chunks (1600 samples at 16kHz) to the backend
- Must be in `public/` folder (static, not bundled)

### WebSocket Protocol

**Browser → Backend → WhisperLive:**
1. Initial JSON config:
   ```json
   {
     "uid": "uuid",
     "task": "transcribe",
     "model": "turbo",
     "use_vad": true,
     "language": null,
     "output": "segments"
   }
   ```
2. Audio frames: Binary int16 PCM, 16kHz, mono
3. End signal: Text frame `"END_OF_AUDIO"`

**WhisperLive → Backend → Browser:**
- `{ "message": "SERVER_READY" }` - Server is ready for audio
- `{ "segments": [...] }` - Transcription updates with completed status
- `{ "error": "..." }` - Error messages

### Frontend Hook (`useTranscription.ts`)
- Manages WebSocket connection
- Handles audio capture and encoding
- Updates transcript in real-time
- Status states: idle, connecting, ready, recording, stopping, error

## Troubleshooting

### WhisperLive takes forever to start
- First run downloads the `turbo` model (~3GB)
- Check logs: `docker compose logs whisper-live`
- Use `docker compose up` without `-d` to see real-time logs

### No audio input
- Check browser microphone permissions
- Verify microphone works: `arecord -d 5 test.wav` (Linux)
- Check browser console for errors

### WebSocket connection fails
- Verify backend is running: `curl http://localhost:8080/healthz`
- Check backend logs: `docker compose logs backend`
- Verify WhisperLive is ready: `docker compose logs whisper-live`

### Transcription not appearing
- Check browser network tab for `/ws/transcribe` WebSocket
- Verify audio chunks are being sent (look for binary frames)
- Check WhisperLive logs for transcription events

## Development

### Hot Reload
- Backend: Air watches Go files, auto-rebuilds on changes
- Frontend: Vite HMR works automatically
- AudioWorklet: Static file, refresh browser to reload

### Building for Production
```bash
# Frontend
cd frontend
pnpm build

# Backend
cd backend
go build -o twinspeak-server ./cmd/server
```

## Notes

- The backend is a stateless proxy—all logic is in the browser or WhisperLive
- AudioWorklet runs on a separate thread, ideal for audio processing
- `crypto.randomUUID()` is used (no external UUID library needed)
- CSS variables approach from shadcn for easy theming

## License

MIT
