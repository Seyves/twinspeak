package services

import (
	"context"

	"github.com/twinspeak/backend/auth"
)

type UserService struct {
	auth    *auth.Auth
	billing *auth.Auth
}

func (s *UserService) SignUp(ctx context.Context) {
		
}
