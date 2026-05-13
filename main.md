## Project overview
I want to create real-time translation app like echotalk.io. Lets start only with real-time transcribation

## Tech requirments (STRICT)
- frontent package manager: pnpm
- frontent framework: react with shadcn components
- backend: golang
- icons: lucide
- real-time transcribation: https://github.com/collabora/WhisperLive (docker container with turbo gpu model)
- testing and development should be in docker

## Core MVP transcribation features
- Converting in real time my speach into transcribed text by pressing a button. Changes should apper in browser on the fly

## Project structure 
- frontent
- backend
- services
- - services/whisper-live
