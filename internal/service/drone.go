package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	util "172.21.5.249/air-trans/at-drone/internal/service/util"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"

	jsonpatch "github.com/evanphx/json-patch"

	"go.mongodb.org/mongo-driver/bson"
)

func (us *MainService) CreateDrone(ctx context.Context, model *pb.Drone, eventAPI bool) (*pb.Drone, error) {
	model.CreatedAt = uint64(time.Now().Unix()) * 1000
	model.UpdatedAt = model.CreatedAt
	_, err := droneColl.InsertOne(ctx, model)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to create drone: %+v", model)

		return nil, err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(model, model),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_DRONE, config.ACT_CREATE, eventAPI, model.ID),
	)

	return model, err
}

func (us *MainService) UpdateDroneByID(ctx context.Context, updatedData *pb.Drone, id string, eventAPI bool) (*pb.Drone, error) {
	originData := &pb.Drone{}
	err := droneColl.Find(ctx, bson.M{"_id": id}).One(originData)
	updatedData.ID = originData.ID

	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find drone  by id: %s", id)

		return nil, err
	}
	updatedData.UpdatedAt = uint64(time.Now().Unix()) * 1000

	_, err = droneColl.UpsertId(ctx, id, updatedData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to upsert drone  by id: %s: %+v", id, updatedData)

		return nil, err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(originData, updatedData),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_DRONE, config.ACT_UPDATE, eventAPI, id),
	)

	return updatedData, err
}

func (us *MainService) DeleteDroneByID(ctx context.Context, id string, eventAPI bool) error {
	data := &pb.Drone{}
	err := droneColl.Find(ctx, bson.M{"_id": id}).One(data)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find drone  by id: %s", id)

		return err
	}

	err = droneColl.Remove(ctx, bson.M{"_id": id})
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to delete drone  by id: %s", id)

		return err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(data, data),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_DRONE, config.ACT_DELETE, eventAPI, id),
	)

	return err
}

func (us *MainService) PatchDroneByID(ctx context.Context, patch *jsonpatch.Patch, id string, eventAPI bool) (*pb.Drone, error) {
	originData := &pb.Drone{}
	err := droneColl.Find(ctx, bson.M{"_id": id}).One(originData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find drone  by id: %s", id)

		return nil, err
	}

	originDataJson, err := json.Marshal(originData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to decode data")

		return nil, err
	}

	updatedDataJson, err := patch.Apply(originDataJson)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to apply patch")

		return nil, err
	}

	updatedData := &pb.Drone{}
	err = json.Unmarshal(updatedDataJson, updatedData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to encode data")

		return nil, err
	}
	updatedData.UpdatedAt = uint64(time.Now().Unix()) * 1000
	_, err = droneColl.UpsertId(ctx, id, updatedData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to upsert drone  by id: %s: %+v", id, updatedData)

		return nil, err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(originData, updatedData),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_DRONE, config.ACT_PATCH, eventAPI, id),
	)

	return updatedData, err
}

func (us *MainService) FindDroneByID(ctx context.Context, id string) (*pb.Drone, error) {
	rs := pb.Drone{}
	err := droneColl.Find(ctx, bson.M{"_id": id}).One(&rs)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find drone  by id: %s", id)

		return nil, err
	}

	return &rs, err
}

func (us *MainService) FindDroneAll(ctx context.Context) ([]pb.Drone, error) {
	rs := []pb.Drone{}
	err := droneColl.Find(ctx, bson.M{}).All(&rs)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find drone  all")
	}

	return rs, err
}

// func (us *MainService) DroneSimulateJob() {
// 	ctx := context.Background()
// 	config.PrintInfoLog(ctx, "Start simulate 1 ")
// 	orders, err := us.SearchOrder(ctx, &pb.SearchOptions{Sorts: []string{"created_at"}})
// 	if err != nil || len(orders.Order) == 0 {
// 		config.PrintErrorLog(ctx, err, "Failed to get locker list")

// 	}
// 	for _, v := range orders.Order {
// 		// v.FlightRoute
// 		for _, j := range v.FlightRoute {
// 			config.PrintInfoLog(ctx, "simulate %v", j)
// 			us.UpdateDroneByID(ctx, &pb.Drone{
// 				ID:          "61a0d432-50c0-4c7f-96f1-f00f19d5f064",
// 				FlightRoute: j,
// 				Speed:       22,
// 				Battery:     22,
// 			}, "61a0d432-50c0-4c7f-96f1-f00f19d5f064", false)
// 		}
// 	}

// }
// func (us *MainService) StartScheduler(sv *MainService) {
// 	ctx := context.Background()

// 	config.PrintInfoLog(ctx, "Begin simulate ")

// 	scheduler := gocron.NewScheduler(time.UTC)
// 	config.PrintInfoLog(ctx, "Start simulate ")

// 	scheduler.Every(3).Second().Do(sv.DroneSimulateJob)
// scheduler.StartAsync()
// }

// func (us *MainService) FindDroneByStatusCondition(
// 	ctx context.Context,
// 	DroneType pb.DroneType,
// 	DroneStatus pb.DroneStatus,
// 	DroneDataStatus pb.DroneDataStatus,
// 	DroneTileStatus pb.DroneTileStatus,
// 	DroneRoutingStatus pb.DroneRoutingStatus,
// ) ([]pb.Drone, error) {
// 	rs := []pb.Drone{}
// 	condition := bson.M{}

// 	if DroneType != pb.DroneType_MST_UNKNOWN {
// 		condition["type"] = bson.M{"$eq": DroneType}
// 	}

// 	if DroneStatus != pb.DroneStatus_MSS_UNKNOWN {
// 		condition["status"] = bson.M{"$eq": DroneStatus}
// 	}

// 	if DroneDataStatus != pb.DroneDataStatus_MSDS_UNKNOWN {
// 		condition["data_status"] = bson.M{"$eq": DroneDataStatus}
// 	}

// 	if DroneTileStatus != pb.DroneTileStatus_MSTS_UNKNOWN {
// 		condition["tile_status"] = bson.M{"$eq": DroneTileStatus}
// 	}

// 	if DroneRoutingStatus != pb.DroneRoutingStatus_MSRS_UNKNOWN {
// 		condition["routing_status"] = bson.M{"$eq": DroneRoutingStatus}
// 	}

// 	err := droneColl.Find(ctx, condition).All(&rs)
// 	if err != nil {
// 		config.PrintErrorLog(ctx, err, "Failed to find drone  by condition")

// 		return nil, err
// 	}

// 	return rs, err
// }

// func (us *MainService) UpdateDroneStatusByID(
// 	ctx context.Context,
// 	id string,
// 	DroneStatus pb.DroneStatus,
// 	DroneDataStatus pb.DroneDataStatus,
// 	DroneTileStatus pb.DroneTileStatus,
// 	DroneRoutingStatus pb.DroneRoutingStatus,
// 	eventAPI bool,
// ) error {
// 	originData := &pb.Drone{}
// 	err := droneColl.Find(ctx, bson.M{"_id": id}).One(originData)
// 	if err != nil {
// 		config.PrintErrorLog(ctx, err, "Failed to find drone  by id: %s", id)

// 		return err
// 	}

// 	updatedData := &pb.Drone{}
// 	err = copier.Copy(updatedData, originData)
// 	if err != nil {
// 		config.PrintErrorLog(ctx, err, "Failed to create drone  updated")

// 		return err
// 	}

// 	if DroneStatus != pb.DroneStatus_MSS_UNKNOWN {
// 		updatedData.Status = DroneStatus
// 	}

// 	if DroneDataStatus != pb.DroneDataStatus_MSDS_UNKNOWN {
// 		updatedData.DataStatus = DroneDataStatus
// 	}

// 	if DroneTileStatus != pb.DroneTileStatus_MSTS_UNKNOWN {
// 		updatedData.TileStatus = DroneTileStatus
// 	}

// 	if DroneRoutingStatus != pb.DroneRoutingStatus_MSRS_UNKNOWN {
// 		updatedData.RoutingStatus = DroneRoutingStatus
// 	}

// 	_, err = droneColl.UpsertId(ctx, id, updatedData)
// 	if err != nil {
// 		config.PrintErrorLog(ctx, err, "Failed to upsert drone  status by id: %s", id)

// 		return err
// 	}

// 	us.publishEvent(
// 		ctx,
// 		util.CreatePublishEventData(originData, updatedData),
// 		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_DRONE, config.ACT_PATCH, eventAPI, id),
// 	)

// 	return err
// }
