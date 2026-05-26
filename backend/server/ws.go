package server

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/twinspeak/backend/billing"
	"github.com/twinspeak/backend/db"
	"github.com/twinspeak/backend/pipeline"
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

	defer func() {
		go func() {
			end := time.Now()
			err := r.metrics.CreateSpeechMetric(context.Background(), db.InsertSpeechParams{
				UserID:    userId,
				InLang:    inLang,
				OutLang:   outLang,
				StartedAt: start,
				EndedAt:   end,
			})
			if err != nil {
				log.Errorf("Error creating speech metrics: %s", err.Error())
			}
		}()
	}()

	in := make(chan []byte)
	out := make(chan pipeline.SpeechEvent)

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

	langConfig := pipeline.LangConfig{
		In:  inLang,
		Out: outLang,
	}

	duration, err := r.pipeline.Pipe(ctx, langConfig, in, out)
	if err != nil {
		log.Errorf("Error during pipe: %s", err.Error())
	}
	log.Infof("Duration: %d", duration)

	err = eg.Wait()
	if err != nil {
		log.Errorf("Error client WS connection: %s", err.Error())
	}
}
