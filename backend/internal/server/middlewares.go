package server

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/twinspeak/backend/internal/db"
)

func (r *RestApi) requestIdMiddleware(c *fiber.Ctx) error {
	requestID, err := uuid.NewV7()
	if err != nil {
		log.Errorf("Error creating request id: %w", err)
		return fiber.NewError(fiber.StatusInternalServerError, "internal server error")
	}
	c.Response().Header.Add("X-Request-ID", requestID.String())
	c.Locals("requestId", requestID)
	return c.Next()
}

func (r *RestApi) authMiddleware(c *fiber.Ctx) error {
	accessToken := c.Cookies(accessTokenCookie)
	userId, err := r.service.ValidateAccessToken(c.Context(), time.Now(), accessToken)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid access token")
	}
	c.Locals("userId", userId)
	return c.Next()
}

func (r *RestApi) wsAuthMiddleware(c *fiber.Ctx) error {
	ticket := c.Query("ticket")
	userId, err := r.service.ValidateWSTicket(c.Context(), time.Now(), ticket)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid ws ticket")
	}
	c.Locals("userId", userId)
	return c.Next()
}

func (r *RestApi) emailVerifiedMiddleware(c *fiber.Ctx) error {
	emailUnverifiedCookieValue := c.Cookies(emailUnverifiedCookie)

	// If cookie says verified, trust it and continue
	// In theory user can delete that cookie and navigate to / just fine
	// But on the next token rotation user will get it again
	if emailUnverifiedCookieValue != "true" {
		return c.Next()
	}

	// If cookie says unverified, double-check with DB
	// This handles: user verified in another tab, cookie manipulation
	userId := c.Locals("userId").(uuid.UUID)
	user, err := r.service.GetCurrentUser(c.Context(), userId)
	if err != nil {
		log.Errorf("Error getting user in email verification middleware: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}

	if !user.EmailVerified {
		return fiber.NewError(fiber.StatusForbidden, "email not verified")
	}

	c.Cookie(getEmailUnverifiedCookie(false, time.Now().Add(time.Minute*10)))

	return c.Next()
}

func (r *RestApi) metricsMiddleware(c *fiber.Ctx) error {
	var (
		code     int
		errorStr string
	)

	start := time.Now()
	err := c.Next()
	if err != nil {
		var e *fiber.Error
		if errors.As(err, &e) {
			code = e.Code
		} else {
			code = fiber.StatusInternalServerError
		}
		errorStr = err.Error()
	} else {
		code = c.Response().StatusCode()
	}

	end := time.Now()
	duration := end.Sub(start)

	reqId := c.Locals("requestId").(uuid.UUID)
	ip, _ := netip.ParseAddr(c.IP())
	userAgent := string(c.Context().UserAgent())
	reqHeadersBytes := len(c.Request().Header.Header())
	reqBodyBytes := len(c.Body())
	respHeadersBytes := len(c.Response().Header.Header())
	respBodyBytes := len(c.Response().Body())

	go func() {
		err := r.service.CreateHttpRequestMetric(
			context.Background(),
			db.InsertHttpRequestParams{
				RequestID:            reqId,
				Method:               c.Method(),
				Route:                c.Route().Path,
				Path:                 c.Path(),
				RecievedAt:           start,
				DurationMs:           int32(duration.Milliseconds()),
				ResponseCode:         int16(code),
				RequestHeadersBytes:  int32(reqHeadersBytes),
				RequestBodyBytes:     int32(reqBodyBytes),
				ResponseHeadersBytes: int32(respHeadersBytes),
				ResponseBodyBytes:    int32(respBodyBytes),
				Ip:                   ip,
				UserAgent:            &userAgent,
				Error:                &errorStr,
			},
		)
		if err != nil {
			log.Errorf("Error creating HTTP request metrics: %s", err.Error())
		}
	}()

	return err
}
