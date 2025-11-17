package service

import (
	"context"
	"fmt"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	util "172.21.5.249/air-trans/at-drone/internal/service/util"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"
)

/*************************************************************************************************/

func (ms *MainService) CreateOrder(ctx context.Context, model *pb.Order) (*pb.Order, error) {
	gConn, err := ms.gClient.GetConn(config.SVC_ORDER)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to get grpc connection")

		return nil, err
	}

	client := pb.NewOrderServiceClient(gConn)

	response, err := client.Create(ctx, model)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to create order service: %+v", model)

		return nil, err
	}

	return response, nil
}

func (ms *MainService) UpdateOrderByID(ctx context.Context, model *pb.Order, id string) (*pb.Order, error) {
	gConn, err := ms.gClient.GetConn(config.SVC_ORDER)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to get grpc connection")

		return nil, err
	}

	client := pb.NewOrderServiceClient(gConn)

	response, err := client.Update(ctx, model)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to update order service by id: %s: %+v", id, model)

		return nil, err
	}

	return response, nil
}

func (ms *MainService) SearchOrder(ctx context.Context, searchOpt *pb.SearchOptions) (*pb.SearchOrderResponse, error) {
	gConn, err := ms.gClient.GetConn(config.SVC_ORDER)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to get grpc connection")

		return nil, err
	}

	client := pb.NewOrderServiceClient(gConn)

	response, err := client.Search(ctx, searchOpt)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to search order service: %+v", searchOpt)

		return nil, err
	}

	return response, nil
}

func (ms *MainService) PatchOrder(ctx context.Context, patchOpt *pb.PatchOptions) error {
	gConn, err := ms.gClient.GetConn(config.SVC_ORDER)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to get grpc connection")

		return err
	}

	client := pb.NewOrderServiceClient(gConn)

	_, err = client.Patch(ctx, patchOpt)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to patch order service: %+v", patchOpt)

		return err
	}

	return nil
}

// func (ms *MainService) UpdateOrderStatusByID(
// 	ctx context.Context,
// 	id string,
// 	OrderStatus pb.OrderStatus,
// 	OrderDataStatus pb.OrderDataStatus,
// 	OrderTileStatus pb.OrderTileStatus,
// 	OrderRoutingStatus pb.OrderRoutingStatus,
// ) error {
// 	if id == "" {
// 		return fmt.Errorf("id is not be empty")
// 	}

// 	patchOptions := []util.PatchOption{}

// 	if OrderStatus != pb.OrderStatus_MSS_UNKNOWN {
// 		patchOptions = append(patchOptions, util.PatchOption{Op: "replace", Path: "/status", Value: OrderStatus})
// 	}

// 	if OrderDataStatus != pb.OrderDataStatus_MSDS_UNKNOWN {
// 		patchOptions = append(patchOptions, util.PatchOption{Op: "replace", Path: "/data_status", Value: OrderDataStatus})
// 	}

// 	if OrderTileStatus != pb.OrderTileStatus_MSTS_UNKNOWN {
// 		patchOptions = append(patchOptions, util.PatchOption{Op: "replace", Path: "/tile_status", Value: OrderTileStatus})
// 	}

// 	if OrderRoutingStatus != pb.OrderRoutingStatus_MSRS_UNKNOWN {
// 		patchOptions = append(patchOptions, util.PatchOption{Op: "replace", Path: "/routing_status", Value: OrderRoutingStatus})
// 	}

// 	return ms.PatchOrder(ctx, &pb.PatchOptions{
// 		ID:         id,
// 		Operations: util.CreatePatchOptions(patchOptions),
// 	})
// }

func (ms *MainService) FindOrderByID(ctx context.Context, id string) (*pb.Order, error) {
	if id == "" {
		return nil, fmt.Errorf("id is not be empty")
	}

	response, err := ms.SearchOrder(ctx, util.CreateSearchOptions(map[string][]string{"_id": {id}}, 0, 1))
	if err != nil {
		return nil, err
	}
	if response.TotalCount == 0 {
		return nil, fmt.Errorf("order service is not exist")
	}

	return response.Order[0], nil
}

func (ms *MainService) FindOrderByIDs(ctx context.Context, ids []string) ([]*pb.Order, error) {
	if len(ids) == 0 {
		return []*pb.Order{}, nil
	} else {
		for _, id := range ids {
			if id == "" {
				return nil, fmt.Errorf("id is not be empty")
			}
		}
	}

	rsponse, err := ms.SearchOrder(ctx, util.CreateSearchOptions(map[string][]string{"_id": ids}, 0, 0))
	if err != nil {
		return nil, err
	}

	return rsponse.Order, nil
}

// func (ms *MainService) DeleteOrderByID(ctx context.Context, deleteOpt *pb.DeleteOptions) error {
// 	gConn, err := ms.gClient.GetConn(config.SVC_ORDER)
// 	if err != nil {
// 		config.PrintErrorLog(ctx, err, "Failed to get grpc connection")

// 		return err
// 	}

// 	client := pb.NewOrderServiceClient(gConn)

// 	_, err = client.Delete(ctx, deleteOpt)
// 	if err != nil {
// 		config.PrintErrorLog(ctx, err, "Failed to delete order service: %+v", deleteOpt)

// 		return err
// 	}

// 	return nil
// }
