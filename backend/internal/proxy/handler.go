package proxy

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type Handler struct {
	whisperURL string
	upgrader   websocket.Upgrader
}

func NewHandler(whisperURL string) *Handler {
	return &Handler{
		whisperURL: whisperURL,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for development
				return true
			},
		},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade browser connection
	browserConn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade browser connection: %v", err)
		return
	}
	defer browserConn.Close()

	// Dial WhisperLive
	whisperConn, _, err := websocket.DefaultDialer.Dial(h.whisperURL, nil)
	if err != nil {
		fmt.Printf("Failed to connect to WhisperLive at %s: %v", h.whisperURL, err)
		browserConn.WriteMessage(websocket.TextMessage, []byte(`{"error":"Failed to connect to transcription service"}`))
		return
	}
	defer whisperConn.Close()

	done := make(chan struct{})

	// Browser → WhisperLive
	go func() {
		defer close(done)
		for {
			mt, data, err := browserConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("Browser WebSocket error: %v", err)
				}
				return
			}
			if err := whisperConn.WriteMessage(mt, data); err != nil {
				fmt.Printf("Failed to forward message to WhisperLive: %v", err)
				return
			}
		}
	}()

	// WhisperLive → Browser
	go func() {
		for {
			mt, data, err := whisperConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("WhisperLive WebSocket error: %v", err)
				}
				return
			}
			if err := browserConn.WriteMessage(mt, data); err != nil {
				fmt.Printf("Failed to forward message to browser: %v", err)
				return
			}
		}
	}()

	// Wait for browser disconnect
	<-done
	fmt.Printf("Browser disconnected, closing WhisperLive connection")
}
