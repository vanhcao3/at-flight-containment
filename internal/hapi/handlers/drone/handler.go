package drone

import (
	"net/http"
	"strconv"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	hapi "172.21.5.249/air-trans/at-drone/internal/hapi"
	types "172.21.5.249/air-trans/at-drone/internal/types"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"

	jsonpatch "github.com/evanphx/json-patch"
	queryoptions "go.jtlabs.io/query"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func CreateRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.POST("/drones", createHandler(s))
}

// Create drone godoc
//
//	@Summary		Create a drone
//	@Description	Create a new drone
//	@Tags			drones
//	@Accept			json
//	@Produce		json
//	@Param			drone		body		pb.Drone	true	"drone body"
//	@Param			eventAPI	query		bool		true	"event api call flag"
//	@Success		200			{object}	pb.Drone
//	@Failure		400			{object}	types.ErrorResponse
//	@Router			/drones [post]
func createHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		u := &pb.Drone{}
		if err := c.Bind(u); err != nil {
			config.PrintErrorLog(ctx, err, "Failed to bind data")

			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
		}

		eventAPI, _ := strconv.ParseBool(c.QueryParam("eventAPI"))
		if !eventAPI {
			u.ID = requestID
		}

		config.PrintDebugLog(ctx, "Create drone: %+v", u)

		_, err := s.MainService.CreateDrone(ctx, u, eventAPI)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to create drone: %+v", u)

			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
		}

		return c.JSON(http.StatusCreated, u)
	}
}

func DeleteByIDRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.DELETE("/drones/:id", deleteByIDHandler(s))
}

// Delete drone by ID godoc
//
//	@Summary		Delete drone by ID
//	@Description	Delete drone by ID
//	@Tags			drones
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string	true	"drone id"
//	@Param			eventAPI	query		bool	true	"event api call flag"
//	@Success		200			{object}	pb.Drone
//	@Failure		400			{object}	types.ErrorResponse
//	@Router			/drones/{id} [delete]
func deleteByIDHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		id := c.Param("id")
		eventAPI, _ := strconv.ParseBool(c.QueryParam("eventAPI"))

		config.PrintDebugLog(ctx, "Delete drone by id: %s", id)

		err := s.MainService.DeleteDroneByID(ctx, id, eventAPI)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to delete drone by id: %s", id)

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
	return s.Router.Root.PATCH("/drones/:id", patchByIDHandler(s))
}

// Patch drone by ID godoc
//
//	@Summary		Patch drone by ID
//	@Description	Patch drone by ID use standard of JSON PATCH https://jsonpatch.com/
//	@Tags			drones
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string			true	"drone id"
//	@Param			drone		body		jsonpatch.Patch	true	"Patch operation format Array of Operation Add, Remove, Replace, Copy, Move, Test. Get example at https://jsonpatch.com/"
//	@Param			eventAPI	query		bool			true	"event api call flag"
//	@Success		200			{object}	pb.PatchResponse
//	@Failure		400			{object}	types.ErrorResponse
//	@Router			/drones/{id} [patch]
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

		config.PrintDebugLog(ctx, "Patch drone by id: %s: %+v", id, patch)

		result, err := s.MainService.PatchDroneByID(ctx, patch, id, eventAPI)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to patch drone by id: %s: %+v", id, patch)

			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusOK, result)
	}
}

func SearchRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/drones/search", searchHandler(s))
}

// Search drone godoc
//
//	@Summary		Search drone
//	@Description	Search drone use Query option https://github.com/jtlabsio/mongo/
//	@Tags			drones
//	@Accept			json
//	@Produce		json
//	@Param			page[page]	query		int	true	"page number"
//	@Param			page[size]	query		int	true	"page size"
//	@Success		200			{object}	pb.Drone
//	@Failure		400			{object}	types.ErrorResponse
//	@Router			/drones/search [get]
func searchHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		opt, err := queryoptions.FromQuerystring(c.Request().URL.RequestURI())
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to get query option from string: %s", c.Request().URL.RequestURI())
		}

		config.PrintDebugLog(ctx, "Search drone: %+v", opt)

		result, count := s.MainService.SearchDrone(ctx, opt)

		config.PrintDebugLog(ctx, "Search drone result: %d", count)

		c.Response().Header().Set("x-total-count", strconv.FormatInt(count, 10))
		return c.JSON(http.StatusOK, *result)
	}
}

func FindByIDRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/drones/:id", findByIDHandler(s))
}

// Find drone by ID godoc
//
//	@Summary		Find drone by ID
//	@Description	Find drone by ID
//	@Tags			drones
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"drone id"
//	@Success		200	{object}	pb.Drone
//	@Failure		400	{object}	types.ErrorResponse
//	@Router			/drones/{id} [get]
func findByIDHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		config.PrintDebugLog(ctx, "Find drone by id: %s", id)

		u, err := s.MainService.FindDroneByID(ctx, id)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to find drone by id: %s", id)

			return c.JSON(http.StatusNotFound, types.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusOK, u)
	}
}

func FindAllRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/drones", findAllHandler(s))
}

// Find all drone godoc
//
//	@Summary		Find all drone
//	@Description	Find all drone
//	@Tags			drones
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	pb.Drone
//	@Failure		400	{object}	types.ErrorResponse
//	@Router			/drones [get]
func findAllHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		config.PrintDebugLog(ctx, "Find drone all")

		u, err := s.MainService.FindDroneAll(ctx)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to find drone all")

			return c.JSON(http.StatusNotFound, types.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			})
		}

		return c.JSON(http.StatusOK, u)
	}
}

func UpdateByIDRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.PUT("/drones/:id", updateByIDHandler(s))
}

// Update drone by ID godoc
//
//	@Summary		Update drone by ID
//	@Description	Update drone by ID
//	@Tags			drones
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string		true	"drone id"
//	@Param			drone		body		pb.Drone	true	"drone body"
//	@Param			eventAPI	query		bool		true	"event api call flag"
//	@Success		200			{object}	pb.Drone
//	@Failure		400			{object}	types.ErrorResponse
//	@Router			/drones/{id} [put]
func updateByIDHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		u := &pb.Drone{}
		if err := c.Bind(u); err != nil {
			config.PrintErrorLog(ctx, err, "Failed to bind data")

			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			})
		}

		id := c.Param("id")
		eventAPI, _ := strconv.ParseBool(c.QueryParam("eventAPI"))

		config.PrintDebugLog(ctx, "Update drone by id: %s: %+v", id, u)

		result, err := s.MainService.UpdateDroneByID(ctx, u, id, eventAPI)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to update drone by id: %s: %+v", id, u)

			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
		}

		return c.JSON(http.StatusOK, result)
	}
}
