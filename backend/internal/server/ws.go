package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/twinspeak/backend/internal/billing"
	"github.com/twinspeak/backend/internal/speechpipeline"
	"golang.org/x/sync/errgroup"
)

var expectedWSClosures = []int{
	websocket.CloseNormalClosure,
	websocket.CloseGoingAway,
	websocket.CloseNoStatusReceived,
	websocket.CloseAbnormalClosure,
}

func (r *RestApi) startSession(c *websocket.Conn) {
	ctx := context.Background()

	userId := c.Locals("userId").(uuid.UUID)
	inLang := c.Query("inLang")
	outLang := c.Query("outLang")
	start := time.Now()

	err := r.users.StartSpeech(ctx, start, userId)
	if err != nil {
		defer c.Close()
		var evt speechpipeline.Event
		if errors.Is(err, billing.ErrInsufficientCredits) {
			evt = speechpipeline.NewSpeechEvent(speechpipeline.ErrorEvt, err.Error())
		} else {
			log.Errorf("Error starting speech: %s", err.Error())
			evt = speechpipeline.NewSpeechEvent(speechpipeline.ErrorEvt, internalServerError)
		}
		if err = c.WriteJSON(evt); err != nil {
			log.Errorf("Error writing message to app: %s", err.Error())
		}
		return
	}

	defer func() {
		now := time.Now()
		err := r.users.EndSpeech(context.Background(), now, userId, inLang, outLang, start, now)
		if err != nil {
			log.Errorf("Error ending speech: %s", err.Error())
		}
	}()

	in := make(chan []byte)
	out := make(chan speechpipeline.Event)

	eg, ctx := errgroup.WithContext(ctx)
	timer := time.NewTimer(time.Second * billing.MaxCreditsPerSession)

	eg.Go(func() error {
		for {
			select {
			case <-timer.C:
				close(in)
				return nil
			default:
				mt, message, err := c.ReadMessage()
				if err != nil {
					c.Close()
					if websocket.IsCloseError(err, expectedWSClosures...) {
						return nil
					} else {
						return fmt.Errorf("cannot read message from: %w", err)
					}
				}

				switch mt {
				case websocket.BinaryMessage:
					in <- message
				case websocket.TextMessage:
					close(in)
					return nil
				}
			}
		}
	})

	eg.Go(func() error {
		defer c.Close()
		for evt := range out {
			err := c.WriteJSON(evt)
			if err != nil {
				return fmt.Errorf("cannot write message to app: %w", err)
			}
		}
		return nil
	})

	langConfig := speechpipeline.LangConfig{
		In:  inLang,
		Out: outLang,
	}

	err = r.pipeline.Pipe(ctx, langConfig, in, out)
	if err != nil {
		log.Errorf("Error during pipe: %s", err.Error())
	}

	err = eg.Wait()
	if err != nil {
		log.Errorf("Error client WS connection: %s", err.Error())
	}
}
