package server

import (
	"encoding/json"
	"net/netip"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func (r *RestApi) googleSignIn(c *fiber.Ctx) error {
	return c.Redirect(r.googleOauth.GetSignInUrl(), fiber.StatusTemporaryRedirect)
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

	userAgent := string(c.Context().UserAgent())
	ip, _ := netip.ParseAddr(c.IP())

	accessToken, refreshToken, err := r.googleOauth.ProcessRedirect(c.Context(), req.Code, req.State, userAgent, &ip)
	if err != nil {
		log.Errorf("Error while processing redirect: %s", err.Error())
		return fiber.NewError(fiber.StatusInternalServerError, internalServerError)
	}

	c.Cookie(getSecureCookie("refresh_token", refreshToken.Value, refreshToken.ExpiresAt))
	c.Cookie(getSecureCookie("access_token", accessToken.Value, accessToken.ExpiresAt))

	return c.SendStatus(fiber.StatusOK)
}
