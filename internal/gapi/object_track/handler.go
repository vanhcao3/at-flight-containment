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

type ObjectTrackHandler struct {
	pb.UnimplementedObjectTrackServiceServer
	MainService *service.MainService
}

func NewObjectTrackHandler(svc *service.MainService) *ObjectTrackHandler {
	return &ObjectTrackHandler{
		MainService: svc,
	}
}

func (h *ObjectTrackHandler) Create(ctx context.Context, ass *pb.ObjectTrack) (*pb.ObjectTrack, error) {
	requestID := uuid.NewString()
	ctx = log.With().Str("x-request-id", requestID).Logger().WithContext(ctx)

	eventAPI := gcommon.GetEventAPIFromContext(ctx)
	if !eventAPI {
		ass.ID = requestID
	}

	config.PrintDebugLog(ctx, "Create object_track: %+v", ass)

	result, err := h.MainService.CreateObjectTrack(ctx, ass, eventAPI)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to create object_track: %+v", ass)

		return nil, err
	}

	return result, nil
}

func (h *ObjectTrackHandler) Update(ctx context.Context, ass *pb.ObjectTrack) (*pb.ObjectTrack, error) {
	requestID := uuid.NewString()
	ctx = log.With().Str("x-request-id", requestID).Logger().WithContext(ctx)

	eventAPI := gcommon.GetEventAPIFromContext(ctx)

	config.PrintDebugLog(ctx, "Update object_track by id: %s: %+v", ass.ID, ass)

	result, err := h.MainService.UpdateObjectTrackByID(ctx, ass, ass.ID, eventAPI)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to update object_track by id: %s: %+v", ass.ID, ass)

		return nil, err
	}

	return result, nil
}

func (h *ObjectTrackHandler) Search(ctx context.Context, so *pb.SearchOptions) (*pb.SearchObjectTrackResponse, error) {
	requestID := uuid.NewString()
	ctx = log.With().Str("x-request-id", requestID).Logger().WithContext(ctx)

	opt := gcommon.ParseQueryOptions(so)

	config.PrintDebugLog(ctx, "Search object_track: %+v", opt)

	result, total := h.MainService.SearchObjectTrack(ctx, opt)

	config.PrintDebugLog(ctx, "Search object_track result: %d", total)

	response := searchResponse(*result, total)

	return response, nil
}

func (h *ObjectTrackHandler) Patch(ctx context.Context, po *pb.PatchOptions) (*pb.PatchResponse, error) {
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

	config.PrintDebugLog(ctx, "Patch object_track by id: %s: %+v", po.ID, patch)

	_, err = h.MainService.PatchObjectTrackByID(ctx, &patch, po.ID, eventAPI)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to patch object_track by id: %s: %+v", po.ID, patch)

		return &pb.PatchResponse{
			IsOk:    false,
			Message: err.Error(),
		}, err
	}

	return &pb.PatchResponse{
		IsOk: true,
	}, nil
}

// func (h *ObjectTrackHandler) Delete(ctx context.Context, do *pb.DeleteOptions) (*pb.DeleteResponse, error) {
// 	requestID := uuid.NewString()
// 	ctx = log.With().Str("x-request-id", requestID).Logger().WithContext(ctx)

// 	eventAPI := gcommon.GetEventAPIFromContext(ctx)

// 	config.PrintDebugLog(ctx, "Delete x by id: %s", do.ID)

// 	err := h.MainService.DeleteObjectTrackByID(ctx, do.ID, eventAPI)
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
func searchResponse(r []*pb.ObjectTrack, t int64) *pb.SearchObjectTrackResponse {
	gresult := make([]*pb.ObjectTrack, len(r))
	for i := range r {
		gresult[i] = r[i]
	}

	return &pb.SearchObjectTrackResponse{
		ObjectTracks: gresult,
		TotalCount:   int32(t),
	}
}
