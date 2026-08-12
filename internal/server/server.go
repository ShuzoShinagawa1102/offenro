package server

import (
	"context"

	"github.com/ShuzoShinagawa1102/offenro/internal/api"
)

type Server struct{}

var _ api.StrictServerInterface = (*Server)(nil)

func New() *Server {
	return &Server{}
}

func (s *Server) GetHealth(
	ctx context.Context,
	request api.GetHealthRequestObject,
) (api.GetHealthResponseObject, error) {

	return api.GetHealth200JSONResponse{
		Status: "ok",
	}, nil
}
