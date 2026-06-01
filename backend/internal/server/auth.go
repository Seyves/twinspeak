package server

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/twinspeak/backend/internal/auth"
)

const (
	accessTokenCookie  = "access_token"
	refreshTokenCookie = "refresh_token"
)

func getSecureCookie(key string, value string, expiresAt time.Time) *fiber.Cookie {
	cookie := new(fiber.Cookie)
	cookie.Name = key
	cookie.Value = value
	cookie.HTTPOnly = true
	cookie.Secure = true
	cookie.SameSite = fiber.CookieSameSiteStrictMode
	cookie.Expires = expiresAt
	return cookie
}

func (r *RestApi) signIn(c *fiber.Ctx) error {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req request
	err := json.Unmarshal(c.Request().Body(), &req)
	if err != nil {
		log.Errorf("Error unmarshalling request body: %s", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, invalidRequestBody)
	}

	// userAgent := string(c.Context().UserAgent())
	// ip, _ := netip.ParseAddr(c.IP())

	accessToken, refreshToken, err := r.users.SignIn(c.Context(), time.Now(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		} else {
			log.Errorf("Error during sign in: %s", err.Error())
			return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
		}
	}
	c.Cookie(getSecureCookie(refreshTokenCookie, refreshToken.Value, refreshToken.ExpiresAt))
	c.Cookie(getSecureCookie(accessTokenCookie, accessToken.Value, accessToken.ExpiresAt))

	return c.SendStatus(fiber.StatusOK)
}

func (r *RestApi) signUp(c *fiber.Ctx) error {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req request
	err := json.Unmarshal(c.Request().Body(), &req)
	if err != nil {
		log.Errorf("Error unmarshalling request body: %s", err.Error())
		return fiber.NewError(fiber.StatusBadRequest, invalidRequestBody)
	}

	// userAgent := string(c.Context().UserAgent())
	// ip, _ := netip.ParseAddr(c.IP())

	accessToken, refreshToken, err := r.users.SignUp(c.Context(), time.Now(), req.Email, req.Password)
	if err != nil {
		log.Errorf("Error during sign in: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}
	c.Cookie(getSecureCookie(refreshTokenCookie, refreshToken.Value, refreshToken.ExpiresAt))
	c.Cookie(getSecureCookie(accessTokenCookie, accessToken.Value, accessToken.ExpiresAt))

	return c.SendStatus(fiber.StatusOK)
}

func (r *RestApi) refresh(c *fiber.Ctx) error {
	refreshTokenStr := c.Cookies(refreshTokenCookie)
	if refreshTokenStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
	}

	// userAgent := string(c.Context().UserAgent())
	// ip, _ := netip.ParseAddr(c.IP())
	//
	accessToken, refreshToken, err := r.users.RotateSession(
		c.Context(),
		time.Now(),
		refreshTokenStr,
	)
	if err != nil {
		if errors.Is(err, auth.ErrMaliciousSuspicion) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
		} else {
			log.Errorf("Error during refreshing token: %s", err.Error())
			return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
		}
	}

	c.Cookie(getSecureCookie(refreshTokenCookie, refreshToken.Value, refreshToken.ExpiresAt))
	c.Cookie(getSecureCookie(accessTokenCookie, accessToken.Value, accessToken.ExpiresAt))

	return c.SendStatus(fiber.StatusOK)
}

func (r *RestApi) getWSTiket(c *fiber.Ctx) error {
	type response struct {
		Ticket string `json:"ticket"`
	}
	userId := c.Locals("userId").(uuid.UUID)
	ticket, err := r.users.GetWSTicket(c.Context(), time.Now(), userId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}
	return c.JSON(response{
		Ticket: ticket.Value,
	})
}

func (r *RestApi) logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies(refreshTokenCookie)
	if refreshToken != "" {
		_ = r.users.Logout(c.Context(), refreshToken)
	}

	// Expire cookies by setting them with a past expiration date
	expiredTime := time.Unix(0, 0)
	c.Cookie(getSecureCookie(accessTokenCookie, "", expiredTime))
	c.Cookie(getSecureCookie(refreshTokenCookie, "", expiredTime))

	return c.SendStatus(fiber.StatusOK)
}
