package api

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/Kegian/agen/examples/server/internal/generated/oapi"
	"github.com/Kegian/agen/examples/server/internal/generated/query"
)

func (s *Service) UsersUserIDGet(ctx context.Context, params oapi.UsersUserIDGetParams) (*oapi.UsersUserIDGetOK, error) {
	user, err := s.repo.GetUser(ctx, params.UserID)
	if s.repo.IsNotFound(err) {
		user, err = s.repo.CreateUser(
			ctx,
			query.CreateUserParams{
				ID:   params.UserID,
				Name: fmt.Sprintf("Default name %d", rand.Int31n(100)),
			},
		)
	}
	if err != nil {
		return nil, err
	}

	return &oapi.UsersUserIDGetOK{
		Data: oapi.User{
			ID:   user.ID,
			Name: user.Name,
		},
	}, nil
}
