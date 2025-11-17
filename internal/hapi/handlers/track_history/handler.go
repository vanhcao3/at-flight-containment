package drone

import (
	"net/http"
	"strconv"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	hapi "172.21.5.249/air-trans/at-drone/internal/hapi"
	"172.21.5.249/air-trans/at-drone/internal/service/util"
	types "172.21.5.249/air-trans/at-drone/internal/types"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"
	"google.golang.org/protobuf/proto"

	jsonpatch "github.com/evanphx/json-patch"
	queryoptions "go.jtlabs.io/query"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func CreateRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.POST("/track_historys", createHandler(s))
}

// Create track_history godoc
//
//	@Summary		Create  track_history
//	@Description	Create a new track_history
//	@Tags			track_historys
//	@Accept			json
//	@Produce		json
//	@Param			track_history	body		pb.TrackHistory	true	"track_history body"
//	@Param			eventAPI		query		bool			true	"event api call flag"
//	@Param			lat				query		number			true	"event api call flag"
//	@Param			lon				query		number			true	"event api call flag"
//	@Param			alt				query		number			true	"event api call flag"
//	@Success		200				{object}	util.TrackHistoryByteAndJson
//	@Failure		400				{object}	types.ErrorResponse
//	@Router			/track_historys [post]
func createHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		latStr := c.QueryParam("lat")
		lonStr := c.QueryParam("lon")
		altStr := c.QueryParam("alt")

		lat, err := strconv.ParseFloat(latStr, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		lon, err := strconv.ParseFloat(lonStr, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		alt, err := strconv.ParseFloat(altStr, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

		var droneLocation pb.Location
		droneLocation.GeodeticPosition = &pb.GeodeticPosition{
			Latitude:  float32(lat),
			Longitude: float32(lon),
			Altitude:  float32(alt),
		}
		// if droneLocation.CreatedAt == 0 {
		// 	droneLocation.CreatedAt = uint64(time.Now().UnixMilli())
		// 	droneLocation.UpdatedAt = droneLocation.CreatedAt

		// } else {
		// 	droneLocation.UpdatedAt = uint64(time.Now().UnixMilli())
		// }
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)
		config.PrintDebugLog(ctx, "drone location query %v", &droneLocation)
		// if err := json.Unmarshal([]byte(droneLocationStr), &droneLocation); err != nil {
		// 	return c.JSON(http.StatusBadRequest, types.ErrorResponse{
		// 		Code:    http.StatusBadRequest,
		// 		Message: err.Error(),
		// 	})
		// }
		u := &pb.TrackHistory{}
		if err := c.Bind(u); err != nil {
			config.PrintErrorLog(ctx, err, "Failed to bind data")

			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
		}
		u.LocationByte, err = proto.Marshal(&droneLocation)
		if err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
		}

		eventAPI, _ := strconv.ParseBool(c.QueryParam("eventAPI"))
		if !eventAPI {
			u.ID = requestID
		}

		config.PrintDebugLog(ctx, "Create track_history: %+v", u)

		_, err = s.MainService.CreateTrackHistory(ctx, u, eventAPI)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to create track_history: %+v", u)

			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
		}

		return c.JSON(http.StatusCreated, util.ConvertToJSONResponse(ctx, []*pb.TrackHistory{u}))
	}
}

func DeleteByIDRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.DELETE("/track_historys/:id", deleteByIDHandler(s))
}

// Delete track_history by ID godoc
//
//	@Summary		Delete track_history by ID
//	@Description	Delete track_history by ID
//	@Tags			track_historys
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string	true	"track_history id"
//	@Param			eventAPI	query		bool	true	"event api call flag"
//	@Success		200			{object}	util.TrackHistoryByteAndJson
//	@Failure		400			{object}	types.ErrorResponse
//	@Router			/track_historys/{id} [delete]
func deleteByIDHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		id := c.Param("id")
		eventAPI, _ := strconv.ParseBool(c.QueryParam("eventAPI"))

		config.PrintDebugLog(ctx, "Delete track_history by id: %s", id)

		err := s.MainService.DeleteTrackHistoryByID(ctx, id, eventAPI)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to delete track_history by id: %s", id)

			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
		}

		return c.JSON(http.StatusOK, types.SucceedResponse{
			Success: true,
		})
	}
}

func PatchByIDRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.PATCH("/track_historys/:id", patchByIDHandler(s))
}

// Patch track_history by ID godoc
//
//	@Summary		Patch track_history by ID
//	@Description	Patch track_history by ID use standard of JSON PATCH https://jsonpatch.com/
//	@Tags			track_historys
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string			true	"track_history id"
//	@Param			drone		body		jsonpatch.Patch	true	"Patch operation format Array of Operation Add, Remove, Replace, Copy, Move, Test. Get example at https://jsonpatch.com/"
//	@Param			eventAPI	query		bool			true	"event api call flag"
//	@Success		200			{object}	pb.PatchResponse
//	@Failure		400			{object}	types.ErrorResponse
//	@Router			/track_historys/{id} [patch]
func patchByIDHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		patch := &jsonpatch.Patch{}
		if err := c.Bind(patch); err != nil {
			config.PrintErrorLog(ctx, err, "Failed to bind data")

			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
		}

		id := c.Param("id")
		eventAPI, _ := strconv.ParseBool(c.QueryParam("eventAPI"))

		config.PrintDebugLog(ctx, "Patch track_history by id: %s: %+v", id, patch)

		result, err := s.MainService.PatchTrackHistoryByID(ctx, patch, id, eventAPI)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to patch track_history by id: %s: %+v", id, patch)

			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusOK, util.ConvertToJSONResponse(ctx, []*pb.TrackHistory{result}))
	}
}

func SearchRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/track_historys/search", searchHandler(s))
}

// Search track_history godoc
//
//	@Summary		Search track_history
//	@Description	Search track_history use Query option https://github.com/jtlabsio/mongo/
//	@Tags			track_historys
//	@Accept			json
//	@Produce		json
//	@Param			page[page]	query		int	true	"page number"
//	@Param			page[size]	query		int	true	"page size"
//	@Success		200			{object}	util.TrackHistoryByteAndJson
//	@Failure		400			{object}	types.ErrorResponse
//	@Router			/track_historys/search [get]
func searchHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		opt, err := queryoptions.FromQuerystring(c.Request().URL.RequestURI())
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to get query option from string: %s", c.Request().URL.RequestURI())
		}

		config.PrintDebugLog(ctx, "Search track_history_track: %+v", opt)

		result, count := s.MainService.SearchTrackHistory(ctx, opt)

		config.PrintDebugLog(ctx, "Search track_history result: %d", count)

		c.Response().Header().Set("x-total-count", strconv.FormatInt(count, 10))
		return c.JSON(http.StatusOK,
			util.ConvertToJSONResponse(ctx, result),
			// result,
		)
	}
}

func FindByIDRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/track_historys/:id", findByIDHandler(s))
}

// Find track_history by ID godoc
//
//	@Summary		Find track_history by ID
//	@Description	Find track_history by ID
//	@Tags			track_historys
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"drone id"
//	@Success		200	{object}	util.TrackHistoryByteAndJson
//	@Failure		400	{object}	types.ErrorResponse
//	@Router			/track_historys/{id} [get]
func findByIDHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		config.PrintDebugLog(ctx, "Find track_history by id: %s", id)

		u, err := s.MainService.FindTrackHistoryByID(ctx, id)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to find track_history by id: %s", id)

			return c.JSON(http.StatusNotFound, types.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusOK, util.ConvertToJSONResponse(ctx, []*pb.TrackHistory{u}))
	}
}

func FindAllRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/track_historys", findAllHandler(s))
}

// Find all track_history godoc
//
//	@Summary		Find all track_history
//	@Description	Find all track_history
//	@Tags			track_historys
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	util.TrackHistoryByteAndJson
//	@Failure		400	{object}	types.ErrorResponse
//	@Router			/track_historys [get]
func findAllHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		config.PrintDebugLog(ctx, "Find track_history all")

		u, err := s.MainService.FindTrackHistoryAll(ctx)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to find track_history all")

			return c.JSON(http.StatusNotFound, types.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			})
		}

		return c.JSON(http.StatusOK, util.ConvertToJSONResponse(ctx, (u)))
	}
}

func UpdateByIDRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.PUT("/track_historys/:id", updateByIDHandler(s))
}

// Update track_history by ID godoc
//
//	@Summary		Update track_history by ID
//	@Description	Update track_history by ID
//	@Tags			track_historys
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string			true	"track_history id"
//	@Param			track_history	body		pb.TrackHistory	true	"track_history body"
//	@Param			eventAPI		query		bool			true	"event api call flag"
//	@Success		200				{object}	util.TrackHistoryByteAndJson
//	@Failure		400				{object}	types.ErrorResponse
//	@Router			/track_historys/{id} [put]
func updateByIDHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		u := &pb.TrackHistory{}
		if err := c.Bind(u); err != nil {
			config.PrintErrorLog(ctx, err, "Failed to bind data")

			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
		}

		id := c.Param("id")
		eventAPI, _ := strconv.ParseBool(c.QueryParam("eventAPI"))

		config.PrintDebugLog(ctx, "Update track_history by id: %s: %+v", id, u)

		result, err := s.MainService.UpdateTrackHistoryByID(ctx, u, id, eventAPI)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to update track_history by id: %s: %+v", id, u)

			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
		}

		return c.JSON(http.StatusOK, util.ConvertToJSONResponse(ctx, []*pb.TrackHistory{result}))
	}
}
