package drone

import (
	"context"
	"encoding/json"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	gcommon "172.21.5.249/air-trans/at-drone/internal/gapi/common"
	service "172.21.5.249/air-trans/at-drone/internal/service"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"

	jsonpatch "github.com/evanphx/json-patch"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type TrackHistoryHandler struct {
	pb.UnimplementedTrackHistoryServiceServer
	MainService *service.MainService
}

func NewTrackHistoryHandler(svc *service.MainService) *TrackHistoryHandler {
	return &TrackHistoryHandler{
		MainService: svc,
	}
}

func (h *TrackHistoryHandler) Create(ctx context.Context, ass *pb.TrackHistory) (*pb.TrackHistory, error) {
	requestID := uuid.NewString()
	ctx = log.With().Str("x-request-id", requestID).Logger().WithContext(ctx)

	eventAPI := gcommon.GetEventAPIFromContext(ctx)
	if !eventAPI {
		ass.ID = requestID
	}

	config.PrintDebugLog(ctx, "Create drone: %+v", ass)

	result, err := h.MainService.CreateTrackHistory(ctx, ass, eventAPI)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to create drone: %+v", ass)

		return nil, err
	}

	return result, nil
}

func (h *TrackHistoryHandler) Update(ctx context.Context, ass *pb.TrackHistory) (*pb.TrackHistory, error) {
	requestID := uuid.NewString()
	ctx = log.With().Str("x-request-id", requestID).Logger().WithContext(ctx)

	eventAPI := gcommon.GetEventAPIFromContext(ctx)

	config.PrintDebugLog(ctx, "Update drone by id: %s: %+v", ass.ID, ass)

	result, err := h.MainService.UpdateTrackHistoryByID(ctx, ass, ass.ID, eventAPI)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to update drone by id: %s: %+v", ass.ID, ass)

		return nil, err
	}

	return result, nil
}

func (h *TrackHistoryHandler) Search(ctx context.Context, so *pb.SearchOptions) (*pb.SearchTrackHistoryResponse, error) {
	requestID := uuid.NewString()
	ctx = log.With().Str("x-request-id", requestID).Logger().WithContext(ctx)

	opt := gcommon.ParseQueryOptions(so)

	config.PrintDebugLog(ctx, "Search drone: %+v", opt)

	result, total := h.MainService.SearchTrackHistory(ctx, opt)

	config.PrintDebugLog(ctx, "Search drone result: %d", total)

	response := searchResponse(result, total)

	return response, nil
}

func (h *TrackHistoryHandler) Patch(ctx context.Context, po *pb.PatchOptions) (*pb.PatchResponse, error) {
	requestID := uuid.NewString()
	ctx = log.With().Str("x-request-id", requestID).Logger().WithContext(ctx)
	var patch jsonpatch.Patch
	err := json.Unmarshal(po.Operations, &patch)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to parse patch option: %+v", po)

		return &pb.PatchResponse{
			IsOk:    false,
			Message: err.Error(),
		}, err
	}

	eventAPI := gcommon.GetEventAPIFromContext(ctx)

	config.PrintDebugLog(ctx, "Patch drone by id: %s: %+v", po.ID, patch)

	_, err = h.MainService.PatchDroneByID(ctx, &patch, po.ID, eventAPI)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to patch drone by id: %s: %+v", po.ID, patch)

		return &pb.PatchResponse{
			IsOk:    false,
			Message: err.Error(),
		}, err
	}

	return &pb.PatchResponse{
		IsOk: true,
	}, nil
}

// func (h *TrackHistoryHandler) Delete(ctx context.Context, do *pb.DeleteOptions) (*pb.DeleteResponse, error) {
// 	requestID := uuid.NewString()
// 	ctx = log.With().Str("x-request-id", requestID).Logger().WithContext(ctx)

// 	eventAPI := gcommon.GetEventAPIFromContext(ctx)

// 	config.PrintDebugLog(ctx, "Delete x by id: %s", do.ID)

// 	err := h.MainService.DeleteDroneByID(ctx, do.ID, eventAPI)
// 	if err != nil {
// 		config.PrintErrorLog(ctx, err, "Failed to delete x by id: %s", do.ID)

// 		return &pb.DeleteResponse{
// 			IsOk:    false,
// 			Message: err.Error(),
// 		}, err
// 	}

// 	return &pb.DeleteResponse{
// 		IsOk: true,
// 	}, nil
// }

/* Create GRPC search response */
func searchResponse(r []*pb.TrackHistory, t int64) *pb.SearchTrackHistoryResponse {

	return &pb.SearchTrackHistoryResponse{
		TrackHistorys: r,
		TotalCount:  int32(t),
	}
}
