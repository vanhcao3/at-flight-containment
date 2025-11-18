package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	util "172.21.5.249/air-trans/at-drone/internal/service/util"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"

	jsonpatch "github.com/evanphx/json-patch"

	"go.mongodb.org/mongo-driver/bson"
)

func (us *MainService) CreateObjectTrack(ctx context.Context, model *pb.ObjectTrack, eventAPI bool) (*pb.ObjectTrack, error) {
	model.CreatedAt = uint64(time.Now().Unix()) * 1000
	model.UpdatedAt = model.CreatedAt
	_, err := objectTrackColl.InsertOne(ctx, model)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to create objectTrack: %+v", model)

		return nil, err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(model, model),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_OBJECT_TRACK, config.ACT_CREATE, eventAPI, model.ID),
	)

	return model, err
}

func (us *MainService) UpdateObjectTrackByID(ctx context.Context, updatedData *pb.ObjectTrack, id string, eventAPI bool) (*pb.ObjectTrack, error) {
	originData := &pb.ObjectTrack{}
	err := objectTrackColl.Find(ctx, bson.M{"_id": id}).One(originData)
	updatedData.ID = originData.ID

	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find objectTrack  by id: %s", id)

		return nil, err
	}
	updatedData.UpdatedAt = uint64(time.Now().Unix()) * 1000
	_, err = objectTrackColl.UpsertId(ctx, id, updatedData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to upsert objectTrack  by id: %s: %+v", id, updatedData)

		return nil, err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(originData, updatedData),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_OBJECT_TRACK, config.ACT_UPDATE, eventAPI, id),
	)

	return updatedData, err
}

func (us *MainService) DeleteObjectTrackByID(ctx context.Context, id string, eventAPI bool) error {
	data := &pb.ObjectTrack{}
	err := objectTrackColl.Find(ctx, bson.M{"_id": id}).One(data)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find objectTrack  by id: %s", id)

		return err
	}

	err = objectTrackColl.Remove(ctx, bson.M{"_id": id})
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to delete objectTrack  by id: %s", id)

		return err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(data, data),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_OBJECT_TRACK, config.ACT_DELETE, eventAPI, id),
	)

	return err
}

func (us *MainService) PatchObjectTrackByID(ctx context.Context, patch *jsonpatch.Patch, id string, eventAPI bool) (*pb.ObjectTrack, error) {
	originData := &pb.ObjectTrack{}
	err := objectTrackColl.Find(ctx, bson.M{"_id": id}).One(originData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find objectTrack  by id: %s", id)

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

	updatedData := &pb.ObjectTrack{}
	err = json.Unmarshal(updatedDataJson, updatedData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to encode data")

		return nil, err
	}
	updatedData.UpdatedAt = uint64(time.Now().Unix()) * 1000
	_, err = objectTrackColl.UpsertId(ctx, id, updatedData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to upsert objectTrack  by id: %s: %+v", id, updatedData)

		return nil, err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(originData, updatedData),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_OBJECT_TRACK, config.ACT_PATCH, eventAPI, id),
	)

	return updatedData, err
}

func (ms *MainService) FindObjectTrackByID(ctx context.Context, id int32) (*pb.ObjectTrack, error) {
	// rs := pb.ObjectTrack{}
	inMemObjectTrack, err := ms.FindInMemObjectTrackByID(ctx, id)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find in_mem object tracks by ID %v \n", id)
		return inMemObjectTrack, err
	}

	return inMemObjectTrack, err
}

func (ms *MainService) FindObjectTrackAll(ctx context.Context) ([]pb.ObjectTrack, error) {
	rs := []pb.ObjectTrack{}
	inMemObjectTracks, err := ms.FindAllInMemObjectTrack(ctx, &emptypb.Empty{})
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to get all in_mem object tracks")
		return rs, err
	}
	fmt.Printf("Got all data %v", len(inMemObjectTracks))
	for _, v := range inMemObjectTracks {
		v.PolarVelocity.Speed = nil
		v.SourceTracks[0].PolarVelocity.Speed = nil
		rs = append(rs, *v)
	}

	_, err = json.Marshal(rs)
	if err != nil {
		fmt.Printf("Error in json marshal %v", err)
	}

	problems := util.FindInvalidFloats(rs, "objecttrack")
	if len(problems) > 0 {
		fmt.Println("Found invalid float values:")
		for _, p := range problems {
			fmt.Println("  -", p)
		}
	}

	return rs, err
}

type FlightContainmentAlert struct {
	HorizontalDeviation float64 `json:"horizontal_deviation"`
	AltitudeDeviation   float64 `json:"altitude_deviation"`
}

func (ms *MainService) CheckFlightContainmentAll(ctx context.Context) error {
	inMemObjectTracks, err := ms.FindAllInMemObjectTrack(ctx, &emptypb.Empty{})
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to get all in_mem object tracks")
		return err
	}
	infringed := make(map[int32]FlightContainmentAlert)
	for _, track := range inMemObjectTracks {
		if track == nil || track.Position == nil {
			continue
		}
		eval, ok := ms.evaluateFlightContainment(float64(track.Position.Latitude), float64(track.Position.Longitude), float64(track.Position.Altitude))
		if ok && (eval.horizontalExceeded || eval.verticalExceeded) {
			if track.ObjectTrackID == 0 {
				continue
			}
			alert := buildContainmentAlert(eval)
			infringed[track.ObjectTrackID] = alert
		}
	}
	newInfringements := ms.filterNewInfringements(infringed)
	if len(newInfringements) == 0 {
		return nil
	}
	if err := ms.Notifier().Publish(EventFlightContainmentInfringement, newInfringements); err != nil {
		config.PrintErrorLog(ctx, err, "Failed to publish flight containment notification")
		return err
	}
	return nil
}

func buildContainmentAlert(eval containmentEvaluation) FlightContainmentAlert {
	alert := FlightContainmentAlert{}
	if eval.horizontalExceeded {
		alert.HorizontalDeviation = signedHorizontalDeviation(eval.offset)
	}
	if eval.verticalExceeded {
		alert.AltitudeDeviation = eval.verticalDeviation
	}
	return alert
}

func signedHorizontalDeviation(offset Vec) float64 {
	magnitude := math.Hypot(offset.x, offset.y)
	if magnitude == 0 {
		return 0
	}
	dominant := offset.x
	if math.Abs(offset.y) > math.Abs(offset.x) {
		dominant = offset.y
	}
	return math.Copysign(magnitude, dominant)
}

func (ms *MainService) filterNewInfringements(current map[int32]FlightContainmentAlert) map[int32]FlightContainmentAlert {
	ms.infringedMu.Lock()
	defer ms.infringedMu.Unlock()

	for id := range ms.activeContainment {
		if _, still := current[id]; !still {
			delete(ms.activeContainment, id)
			delete(ms.notifiedTracks, id)
		}
	}

	renotifyInterval := ms.flightContainmentRenotifyInterval()
	now := time.Now()
	newOnes := make(map[int32]FlightContainmentAlert)
	for id, alert := range current {
		ms.activeContainment[id] = struct{}{}
		lastNotified, alreadyNotified := ms.notifiedTracks[id]
		if !alreadyNotified {
			ms.notifiedTracks[id] = now
			newOnes[id] = alert
			continue
		}
		if renotifyInterval <= 0 {
			continue
		}
		if now.Sub(lastNotified) >= renotifyInterval {
			ms.notifiedTracks[id] = now
			newOnes[id] = alert
		}
	}

	return newOnes
}

func (ms *MainService) flightContainmentRenotifyInterval() time.Duration {
	if ms == nil || ms.SvcConfig == nil {
		return 0
	}
	seconds := ms.SvcConfig.FlightContainment.RenotifySeconds
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}

func (ms *MainService) StartFlightContainmentMonitor(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := ms.CheckFlightContainmentAll(ctx); err != nil && ctx.Err() == nil {
					config.PrintErrorLog(ctx, err, "Flight containment monitor failed")
				}
			}
		}
	}()
}

type MobileDroneResponse struct {
	ObjectID      string              `json:"drone_id"`
	PolarVelocity pb.PolarVelocity    `json:"polar_velocity"`
	Position      pb.GeodeticPosition `json:"position"`
	UpdatedAt     uint64              `json:"updated_at"`
}

func (ms *MainService) FindObjectTrackByDroneID(ctx context.Context, id string) (*MobileDroneResponse, error) {
	rs := MobileDroneResponse{}
	inMemObjectTracks, err := ms.FindAllInMemObjectTrack(ctx, &emptypb.Empty{})
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to get all in_mem object tracks")
		return &rs, err
	}

	for _, v := range inMemObjectTracks {
		if v.ObjectID == id {
			rs.ObjectID = v.ObjectID
			rs.PolarVelocity = *v.PolarVelocity
			rs.Position = *v.Position
			rs.UpdatedAt = v.UpdatedAt
		}
	}
	return &rs, err
}

// func (us *MainService) ObjectTrackSimulateJob() {
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
// 			us.UpdateObjectTrackByID(ctx, &pb.ObjectTrack{
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

// 	scheduler.Every(3).Second().Do(sv.ObjectTrackSimulateJob)
// scheduler.StartAsync()
// }

// func (us *MainService) FindObjectTrackByStatusCondition(
// 	ctx context.Context,
// 	ObjectTrackType pb.ObjectTrackType,
// 	ObjectTrackStatus pb.ObjectTrackStatus,
// 	ObjectTrackDataStatus pb.ObjectTrackDataStatus,
// 	ObjectTrackTileStatus pb.ObjectTrackTileStatus,
// 	ObjectTrackRoutingStatus pb.ObjectTrackRoutingStatus,
// ) ([]pb.ObjectTrack, error) {
// 	rs := []pb.ObjectTrack{}
// 	condition := bson.M{}

// 	if ObjectTrackType != pb.ObjectTrackType_MST_UNKNOWN {
// 		condition["type"] = bson.M{"$eq": ObjectTrackType}
// 	}

// 	if ObjectTrackStatus != pb.ObjectTrackStatus_MSS_UNKNOWN {
// 		condition["status"] = bson.M{"$eq": ObjectTrackStatus}
// 	}

// 	if ObjectTrackDataStatus != pb.ObjectTrackDataStatus_MSDS_UNKNOWN {
// 		condition["data_status"] = bson.M{"$eq": ObjectTrackDataStatus}
// 	}

// 	if ObjectTrackTileStatus != pb.ObjectTrackTileStatus_MSTS_UNKNOWN {
// 		condition["tile_status"] = bson.M{"$eq": ObjectTrackTileStatus}
// 	}

// 	if ObjectTrackRoutingStatus != pb.ObjectTrackRoutingStatus_MSRS_UNKNOWN {
// 		condition["routing_status"] = bson.M{"$eq": ObjectTrackRoutingStatus}
// 	}

// 	err := objectTrackColl.Find(ctx, condition).All(&rs)
// 	if err != nil {
// 		config.PrintErrorLog(ctx, err, "Failed to find objectTrack  by condition")

// 		return nil, err
// 	}

// 	return rs, err
// }

// func (us *MainService) UpdateObjectTrackStatusByID(
// 	ctx context.Context,
// 	id string,
// 	ObjectTrackStatus pb.ObjectTrackStatus,
// 	ObjectTrackDataStatus pb.ObjectTrackDataStatus,
// 	ObjectTrackTileStatus pb.ObjectTrackTileStatus,
// 	ObjectTrackRoutingStatus pb.ObjectTrackRoutingStatus,
// 	eventAPI bool,
// ) error {
// 	originData := &pb.ObjectTrack{}
// 	err := objectTrackColl.Find(ctx, bson.M{"_id": id}).One(originData)
// 	if err != nil {
// 		config.PrintErrorLog(ctx, err, "Failed to find objectTrack  by id: %s", id)

// 		return err
// 	}

// 	updatedData := &pb.ObjectTrack{}
// 	err = copier.Copy(updatedData, originData)
// 	if err != nil {
// 		config.PrintErrorLog(ctx, err, "Failed to create objectTrack  updated")

// 		return err
// 	}

// 	if ObjectTrackStatus != pb.ObjectTrackStatus_MSS_UNKNOWN {
// 		updatedData.Status = ObjectTrackStatus
// 	}

// 	if ObjectTrackDataStatus != pb.ObjectTrackDataStatus_MSDS_UNKNOWN {
// 		updatedData.DataStatus = ObjectTrackDataStatus
// 	}

// 	if ObjectTrackTileStatus != pb.ObjectTrackTileStatus_MSTS_UNKNOWN {
// 		updatedData.TileStatus = ObjectTrackTileStatus
// 	}

// 	if ObjectTrackRoutingStatus != pb.ObjectTrackRoutingStatus_MSRS_UNKNOWN {
// 		updatedData.RoutingStatus = ObjectTrackRoutingStatus
// 	}

// 	_, err = objectTrackColl.UpsertId(ctx, id, updatedData)
// 	if err != nil {
// 		config.PrintErrorLog(ctx, err, "Failed to upsert objectTrack  status by id: %s", id)

// 		return err
// 	}

// 	us.publishEvent(
// 		ctx,
// 		util.CreatePublishEventData(originData, updatedData),
// 		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_OBJECT_TRACK, config.ACT_PATCH, eventAPI, id),
// 	)

// 	return err
// }
