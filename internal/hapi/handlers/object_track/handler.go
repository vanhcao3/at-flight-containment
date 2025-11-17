package drone

import (
	"net/http"
	"strconv"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	hapi "172.21.5.249/air-trans/at-drone/internal/hapi"
	types "172.21.5.249/air-trans/at-drone/internal/types"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// func CreateRoute(s *hapi.Server) *echo.Route {
// 	return s.Router.Root.POST("/object_tracks", createHandler(s))
// }

// // Create object_track godoc
// //
// //	@Summary		Create a object_track
// //	@Description	Create a new object_track
// //	@Tags			object_tracks
// //	@Accept			json
// //	@Produce		json
// //	@Param			object_track	body		pb.ObjectTrack	true	"object_track body"
// //	@Param			eventAPI		query		bool			true	"event api call flag"
// //	@Success		200				{object}	pb.ObjectTrack
// //	@Failure		400				{object}	types.ErrorResponse
// //	@Router			/object_tracks [post]
// func createHandler(s *hapi.Server) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		requestID := uuid.NewString()
// 		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
// 		c.Response().Header().Set("x-request-id", requestID)

// 		u := &pb.ObjectTrack{}
// 		if err := c.Bind(u); err != nil {
// 			config.PrintErrorLog(ctx, err, "Failed to bind data")

// 			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
// 				Code:    http.StatusBadRequest,
// 				Message: err.Error(),
// 			})
// 		}

// 		eventAPI, _ := strconv.ParseBool(c.QueryParam("eventAPI"))
// 		if !eventAPI {
// 			u.ID = requestID
// 		}

// 		config.PrintDebugLog(ctx, "Create object_track: %+v", u)

// 		_, err := s.MainService.CreateObjectTrack(ctx, u, eventAPI)
// 		if err != nil {
// 			config.PrintErrorLog(ctx, err, "Failed to create object_track: %+v", u)

// 			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
// 				Code:    http.StatusInternalServerError,
// 				Message: err.Error(),
// 			})
// 		}

// 		return c.JSON(http.StatusCreated, u)
// 	}
// }
// func DeleteByIDRoute(s *hapi.Server) *echo.Route {
// 	return s.Router.Root.DELETE("/object_tracks/:id", deleteByIDHandler(s))
// }

// // Delete object_track by ID godoc
// //
// //	@Summary		Delete object_track by ID
// //	@Description	Delete object_track by ID
// //	@Tags			object_tracks
// //	@Accept			json
// //	@Produce		json
// //	@Param			id			path		string	true	"object_track id"
// //	@Param			eventAPI	query		bool	true	"event api call flag"
// //	@Success		200			{object}	pb.ObjectTrack
// //	@Failure		400			{object}	types.ErrorResponse
// //	@Router			/object_tracks/{id} [delete]
// func deleteByIDHandler(s *hapi.Server) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		requestID := uuid.NewString()
// 		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
// 		c.Response().Header().Set("x-request-id", requestID)

// 		id := c.Param("id")
// 		eventAPI, _ := strconv.ParseBool(c.QueryParam("eventAPI"))

// 		config.PrintDebugLog(ctx, "Delete object_track by id: %s", id)

// 		err := s.MainService.DeleteObjectTrackByID(ctx, id, eventAPI)
// 		if err != nil {
// 			config.PrintErrorLog(ctx, err, "Failed to delete object_track by id: %s", id)

// 			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
// 				Code:    http.StatusInternalServerError,
// 				Message: err.Error(),
// 			})
// 		}

// 		return c.JSON(http.StatusOK, types.SucceedResponse{
// 			Success: true,
// 		})
// 	}
// }

// func PatchByIDRoute(s *hapi.Server) *echo.Route {
// 	return s.Router.Root.PATCH("/object_tracks/:id", patchByIDHandler(s))
// }

// // Patch object_track by ID godoc
// //
// //	@Summary		Patch object_track by ID
// //	@Description	Patch object_track by ID use standard of JSON PATCH https://jsonpatch.com/
// //	@Tags			object_tracks
// //	@Accept			json
// //	@Produce		json
// //	@Param			id				path		string			true	"object_track id"
// //	@Param			object_track	body		jsonpatch.Patch	true	"Patch operation format Array of Operation Add, Remove, Replace, Copy, Move, Test. Get example at https://jsonpatch.com/"
// //	@Param			eventAPI		query		bool			true	"event api call flag"
// //	@Success		200				{object}	pb.PatchResponse
// //	@Failure		400				{object}	types.ErrorResponse
// //	@Router			/object_tracks/{id} [patch]
// func patchByIDHandler(s *hapi.Server) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		requestID := uuid.NewString()
// 		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
// 		c.Response().Header().Set("x-request-id", requestID)

// 		patch := &jsonpatch.Patch{}
// 		if err := c.Bind(patch); err != nil {
// 			config.PrintErrorLog(ctx, err, "Failed to bind data")

// 			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
// 				Code:    http.StatusBadRequest,
// 				Message: err.Error(),
// 			})
// 		}

// 		id := c.Param("id")
// 		eventAPI, _ := strconv.ParseBool(c.QueryParam("eventAPI"))

// 		config.PrintDebugLog(ctx, "Patch object_track by id: %s: %+v", id, patch)

// 		result, err := s.MainService.PatchObjectTrackByID(ctx, patch, id, eventAPI)
// 		if err != nil {
// 			config.PrintErrorLog(ctx, err, "Failed to patch object_track by id: %s: %+v", id, patch)

// 			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
// 				Code:    http.StatusInternalServerError,
// 				Message: err.Error(),
// 			})
// 		}
// 		return c.JSON(http.StatusOK, result)
// 	}
// }

// func SearchRoute(s *hapi.Server) *echo.Route {
// 	return s.Router.Root.GET("/object_tracks/search", searchHandler(s))
// }

// // Search object_track godoc
// //
// //	@Summary		Search object_track
// //	@Description	Search object_track use Query option https://github.com/jtlabsio/mongo/
// //	@Tags			object_tracks
// //	@Accept			json
// //	@Produce		json
// //	@Param			page[page]	query		int	true	"page number"
// //	@Param			page[size]	query		int	true	"page size"
// //	@Success		200			{object}	pb.ObjectTrack
// //	@Failure		400			{object}	types.ErrorResponse
// //	@Router			/object_tracks/search [get]
// func searchHandler(s *hapi.Server) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		requestID := uuid.NewString()
// 		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
// 		c.Response().Header().Set("x-request-id", requestID)

// 		opt, err := queryoptions.FromQuerystring(c.Request().URL.RequestURI())
// 		if err != nil {
// 			config.PrintErrorLog(ctx, err, "Failed to get query option from string: %s", c.Request().URL.RequestURI())
// 		}

// 		config.PrintDebugLog(ctx, "Search object_track_track: %+v", opt)

// 		result, count := s.MainService.SearchObjectTrack(ctx, opt)

// 		config.PrintDebugLog(ctx, "Search object_track result: %d", count)

// 		c.Response().Header().Set("x-total-count", strconv.FormatInt(count, 10))
// 		return c.JSON(http.StatusOK,
// 			result,
// 			// result,
// 		)
// 	}
// }

func FindByIDRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/object_tracks/:id", findByIDHandler(s))
}

// Find object_track by ID godoc
//
//	@Summary		Find object_track by ID
//	@Description	Find object_track by ID
//	@Tags			object_tracks
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"object_track id"
//	@Success		200	{object}	pb.ObjectTrack
//	@Failure		400	{object}	types.ErrorResponse
//	@Router			/object_tracks/{id} [get]
func findByIDHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		id, err := strconv.ParseInt(c.Param("id"), 0, 32)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to find parse  id: %s", id)
			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
		}

		c.Response().Header().Set("x-request-id", requestID)

		config.PrintDebugLog(ctx, "Find object_track by id: %s", id)

		u, err := s.MainService.FindObjectTrackByID(ctx, int32(id))
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to find object_track by id: %v", id)

			return c.JSON(http.StatusNotFound, types.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusOK, u)
	}
}

func FindAllRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/object_tracks", findAllHandler(s))
}

// Find all object_track godoc
//
//	@Summary		Find all object_track
//	@Description	Find all object_track
//	@Tags			object_tracks
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	pb.ObjectTrack
//	@Failure		400	{object}	types.ErrorResponse
//	@Router			/object_tracks [get]
func findAllHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		config.PrintDebugLog(ctx, "Find object_track all")

		u, err := s.MainService.FindObjectTrackAll(ctx)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to find object_track all")

			return c.JSON(http.StatusNotFound, types.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			})
		}
		// fmt.Printf("XXXX heck %+v", u)
		return c.JSON(http.StatusOK, u)
	}
}

// func UpdateByIDRoute(s *hapi.Server) *echo.Route {
// 	return s.Router.Root.PUT("/object_tracks/:id", updateByIDHandler(s))
// }

// // Update object_track by ID godoc
// //
// //	@Summary		Update object_track by ID
// //	@Description	Update object_track by ID
// //	@Tags			object_tracks
// //	@Accept			json
// //	@Produce		json
// //	@Param			id				path		string			true	"object_track id"
// //	@Param			object_track	body		pb.ObjectTrack	true	"object_track body"
// //	@Param			eventAPI		query		bool			true	"event api call flag"
// //	@Success		200				{object}	pb.ObjectTrack
// //	@Failure		400				{object}	types.ErrorResponse
// //	@Router			/object_tracks/{id} [put]
// func updateByIDHandler(s *hapi.Server) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		requestID := uuid.NewString()
// 		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
// 		c.Response().Header().Set("x-request-id", requestID)

// 		u := &pb.ObjectTrack{}
// 		if err := c.Bind(u); err != nil {
// 			config.PrintErrorLog(ctx, err, "Failed to bind data")

// 			return c.JSON(http.StatusBadRequest, types.ErrorResponse{
// 				Code:    http.StatusBadRequest,
// 				Message: err.Error(),
// 			})
// 		}

// 		id := c.Param("id")
// 		eventAPI, _ := strconv.ParseBool(c.QueryParam("eventAPI"))

// 		config.PrintDebugLog(ctx, "Update object_track by id: %s: %+v", id, u)

// 		result, err := s.MainService.UpdateObjectTrackByID(ctx, u, id, eventAPI)
// 		if err != nil {
// 			config.PrintErrorLog(ctx, err, "Failed to update object_track by id: %s: %+v", id, u)

// 			return c.JSON(http.StatusInternalServerError, types.ErrorResponse{
// 				Code:    http.StatusInternalServerError,
// 				Message: err.Error(),
// 			})
// 		}

// 		return c.JSON(http.StatusOK, result)
// 	}
// }

func FindObjecTrackByDroneIDRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/mobile/drone/:id", findObjecTrackByDroneIDHandler(s))
}

// Find object_track by drone ID godoc
//
//	@Summary		Find object_track by drone ID
//	@Description	Find object_track by drone ID
//	@Tags			object_tracks
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"drone id"
//	@Success		200	{object}	pb.ObjectTrack
//	@Failure		400	{object}	types.ErrorResponse
//	@Router			/mobile/drone/{id} [get]
func findObjecTrackByDroneIDHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		id := c.Param("id")
		c.Response().Header().Set("x-request-id", requestID)

		config.PrintDebugLog(ctx, "Find object_track by drone id: %s", id)

		u, err := s.MainService.FindObjectTrackByDroneID(ctx, id)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to find object_track by drone id: %v", id)

			return c.JSON(http.StatusNotFound, types.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusOK, u)
	}
}
