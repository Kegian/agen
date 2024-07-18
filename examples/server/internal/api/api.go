package api

import (
	"github.com/Kegian/agen/examples/server/internal/generated/oapi"
	"github.com/Kegian/agen/examples/server/internal/repo"
)

type Service struct {
	oapi.UnimplementedHandler

	repo *repo.Repo
}

func NewService(repo *repo.Repo) *Service {
	return &Service{repo: repo}
}
