package service

import (
	"context"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

/*************************************************************************************************/

func (ms *MainService) FindAllInMemObjectTrack(ctx context.Context, empty *emptypb.Empty) ([]*pb.ObjectTrack, error) {
	var res []*pb.ObjectTrack
	gConn, err := ms.gClient.GetConn(config.SVC_EVENT_LISTENER)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to get grpc connection")

		return nil, err
	}

	client := pb.NewObjectTrackServiceClient(gConn)

	response, err := client.FindAll(ctx, empty)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find all object track service")

		return nil, err
	}
	res = append(res, response.ObjectTracks...)
	return res, nil
}

func (ms *MainService) FindInMemObjectTrackByID(ctx context.Context, id int32) (*pb.ObjectTrack, error) {

	gConn, err := ms.gClient.GetConn(config.SVC_EVENT_LISTENER)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to get grpc connection")

		return &pb.ObjectTrack{}, err
	}

	client := pb.NewObjectTrackServiceClient(gConn)

	response, err := client.FindByID(ctx, wrapperspb.Int32(id))
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find in_mem object track service by ID %+v", id)

		return response, err
	}

	return response, err
}
