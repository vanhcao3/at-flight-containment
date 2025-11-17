package hapi

import (
	"errors"
	"fmt"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	i18n "172.21.5.249/air-trans/at-drone/internal/i18n"
	service "172.21.5.249/air-trans/at-drone/internal/service"

	"github.com/labstack/echo/v4"

	echopprof "github.com/nguyenngodinh/echo-pprof"
)

type Router struct {
	Routes     []*echo.Route
	Root       *echo.Group
	Management *echo.Group
}

type Server struct {
	Echo        *echo.Echo
	Config      config.ServiceConfig
	Router      *Router
	I18n        *i18n.Service
	MainService *service.MainService
}

func NewServer(svc *service.MainService, cfg config.ServiceConfig) *Server {
	return &Server{
		Echo:        nil,
		Router:      nil,
		Config:      cfg,
		MainService: svc,
		I18n:        nil,
	}
}

func (s *Server) Ready() bool {
	return s.Echo != nil &&
		s.Router != nil &&
		s.MainService != nil
}

func (s *Server) InitI18n() error {
	i18nService, err := i18n.New(s.Config)
	if err != nil {
		return err
	}

	s.I18n = i18nService

	return nil
}

func (s *Server) Start(errs chan error) {
	if !s.Ready() {
		errs <- errors.New("server is not ready")
	}
	echopprof.Wrap(s.Echo)

	errs <- s.Echo.Start(fmt.Sprintf("%s:%d", s.Config.HttpConfig.HttpHost, s.Config.HttpConfig.HttpPort))
}
