package websocket

import (
	"net/http"

	"172.21.5.249/air-trans/at-drone/internal/hapi"
	"172.21.5.249/air-trans/at-drone/internal/service"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func RegisterRoutes(s *hapi.Server) []*echo.Route {
	return []*echo.Route{
		s.Router.Root.GET("/ws/flight-containment", handler(s, service.EventFlightContainmentInfringement)),
	}
}

func handler(s *hapi.Server, event service.NotificationEvent) echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer conn.Close()
		messages, unsubscribe := s.MainService.Notifier().Subscribe(event)
		defer unsubscribe()
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()
		for {
			select {
			case <-done:
				return nil
			case <-c.Request().Context().Done():
				return nil
			case msg, ok := <-messages:
				if !ok {
					return nil
				}
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return err
				}
			}
		}
	}
}
