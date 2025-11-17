package handlers

import (
	hapi "172.21.5.249/air-trans/at-drone/internal/hapi"
	common "172.21.5.249/air-trans/at-drone/internal/hapi/handlers/common"
	drone "172.21.5.249/air-trans/at-drone/internal/hapi/handlers/drone"
	objectTrack "172.21.5.249/air-trans/at-drone/internal/hapi/handlers/object_track"
	trackHistory "172.21.5.249/air-trans/at-drone/internal/hapi/handlers/track_history"
	"172.21.5.249/air-trans/at-drone/internal/hapi/handlers/websocket"

	"github.com/labstack/echo/v4"
)

func AttackAllRoutes(s *hapi.Server) {
	s.Router.Routes = []*echo.Route{
		// GET /-/version
		common.GetVersionRoute(s),
		// GET /-/ready
		common.GetReadyRoute(s),
		// GET /-/healthy
		common.GetHealthyRoute(s),

		drone.CreateRoute(s),
		drone.DeleteByIDRoute(s),
		drone.PatchByIDRoute(s),
		drone.FindAllRoute(s),
		drone.SearchRoute(s),
		drone.FindByIDRoute(s),
		drone.UpdateByIDRoute(s),

		trackHistory.CreateRoute(s),
		trackHistory.DeleteByIDRoute(s),
		trackHistory.PatchByIDRoute(s),
		trackHistory.FindAllRoute(s),
		trackHistory.SearchRoute(s),
		trackHistory.FindByIDRoute(s),
		trackHistory.UpdateByIDRoute(s),

		// objectTrack.CreateRoute(s),
		// objectTrack.DeleteByIDRoute(s),
		// objectTrack.PatchByIDRoute(s),
		objectTrack.FindAllRoute(s),
		// objectTrack.SearchRoute(s),
		objectTrack.FindByIDRoute(s),
		objectTrack.FindObjecTrackByDroneIDRoute(s),
		// objectTrack.UpdateByIDRoute(s),
	}

	s.Router.Routes = append(s.Router.Routes, websocket.RegisterRoutes(s)...)
}
