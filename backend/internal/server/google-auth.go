package server

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/twinspeak/backend/internal/googleauth"
)

const sessionStateCookie = "session_state"

func (r *RestApi) googleSignIn(c *fiber.Ctx) error {
	url, state, err := r.users.GoogleRedirect()
	if err != nil {
		log.Errorf("Error generation sign-in url: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}
	stateExpiresAt := time.Now().Add(time.Minute * 10)
	c.Cookie(getSecureCookie("session_state", state, stateExpiresAt))
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

func (r *RestApi) googleCallback(c *fiber.Ctx) error {
	type request struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}

	var req request
	err := json.Unmarshal(c.Request().Body(), &req)
	if err != nil {
		log.Errorf("Error unmarshalling request body: %s", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, invalidRequestBody)
	}

	sessionState := c.Cookies("session_state")
	// userAgent := string(c.Context().UserAgent())
	// ip, _ := netip.ParseAddr(c.IP())

	accessToken, refreshToken, err := r.users.GoogleCallback(c.Context(), time.Now(), req.Code, sessionState, req.State)
	if errors.Is(err, googleauth.ErrGoogleInvalidState) ||
		errors.Is(err, googleauth.ErrGoogleCannotExchange) ||
		errors.Is(err, googleauth.ErrGoogleInvalidIdToken) {
		log.Errorf("Error while processing redirect: %s", err.Error())
		return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired creds")
	} else if err != nil {
		log.Errorf("Error while processing redirect: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}

	c.Cookie(getSecureCookie(refreshTokenCookie, refreshToken.Value, refreshToken.ExpiresAt))
	c.Cookie(getSecureCookie(accessTokenCookie, accessToken.Value, accessToken.ExpiresAt))
	// Killing session state cookie
	c.Cookie(getSecureCookie(sessionStateCookie, "", time.Unix(0, 0)))

	return c.SendStatus(fiber.StatusOK)
}
