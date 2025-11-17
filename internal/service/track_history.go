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

func (us *MainService) CreateTrackHistory(ctx context.Context, model *pb.TrackHistory, eventAPI bool) (*pb.TrackHistory, error) {
	colName := util.FindCollectionName(HISTORY_TRACK_PREFIX, uint64(time.Now().UnixMilli()))
	trackHistoryColl = db.Collection(colName)
	model.CreatedAt = uint64(time.Now().Unix()) * 1000
	_, err := trackHistoryColl.InsertOne(ctx, model)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to create track_history: %+v", model)

		return nil, err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(model, model),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_TRACK_HISTORY, config.ACT_CREATE, eventAPI, model.ID),
	)

	return model, err
}

func (us *MainService) UpdateTrackHistoryByID(ctx context.Context, updatedData *pb.TrackHistory, id string, eventAPI bool) (*pb.TrackHistory, error) {
	colName := util.FindCollectionName(HISTORY_TRACK_PREFIX, uint64(time.Now().UnixMilli()))
	trackHistoryColl = db.Collection(colName)
	originData := &pb.TrackHistory{}
	err := trackHistoryColl.Find(ctx, bson.M{"_id": id}).One(originData)
	updatedData.ID = originData.ID

	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find track_history  by id: %s", id)

		return nil, err
	}
	updatedData.CreatedAt = uint64(time.Now().Unix()) * 1000

	_, err = trackHistoryColl.UpsertId(ctx, id, updatedData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to upsert track_history  by id: %s: %+v", id, updatedData)

		return nil, err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(originData, updatedData),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_TRACK_HISTORY, config.ACT_UPDATE, eventAPI, id),
	)

	return updatedData, err
}

func (us *MainService) DeleteTrackHistoryByID(ctx context.Context, id string, eventAPI bool) error {
	colName := util.FindCollectionName(HISTORY_TRACK_PREFIX, uint64(time.Now().UnixMilli()))
	trackHistoryColl = db.Collection(colName)
	data := &pb.TrackHistory{}
	err := trackHistoryColl.Find(ctx, bson.M{"_id": id}).One(data)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find track_history  by id: %s", id)

		return err
	}

	err = trackHistoryColl.Remove(ctx, bson.M{"_id": id})
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to delete track_history  by id: %s", id)

		return err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(data, data),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_TRACK_HISTORY, config.ACT_DELETE, eventAPI, id),
	)

	return err
}

func (us *MainService) PatchTrackHistoryByID(ctx context.Context, patch *jsonpatch.Patch, id string, eventAPI bool) (*pb.TrackHistory, error) {
	colName := util.FindCollectionName(HISTORY_TRACK_PREFIX, uint64(time.Now().UnixMilli()))
	trackHistoryColl = db.Collection(colName)
	originData := &pb.TrackHistory{}
	err := trackHistoryColl.Find(ctx, bson.M{"_id": id}).One(originData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find track_history  by id: %s", id)

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

	updatedData := &pb.TrackHistory{}
	err = json.Unmarshal(updatedDataJson, updatedData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to encode data")

		return nil, err
	}
	updatedData.CreatedAt = uint64(time.Now().Unix()) * 1000
	_, err = trackHistoryColl.UpsertId(ctx, id, updatedData)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to upsert track_history  by id: %s: %+v", id, updatedData)

		return nil, err
	}

	us.publishEvent(
		ctx,
		util.CreatePublishEventData(originData, updatedData),
		fmt.Sprintf("%s.%s.%s.%t.%s", config.SVC_DRONE, config.RSC_TRACK_HISTORY, config.ACT_PATCH, eventAPI, id),
	)

	return updatedData, err
}

func (us *MainService) FindTrackHistoryByID(ctx context.Context, id string) (*pb.TrackHistory, error) {
	colName := util.FindCollectionName(HISTORY_TRACK_PREFIX, uint64(time.Now().UnixMilli()))
	trackHistoryColl = db.Collection(colName)
	rs := pb.TrackHistory{}
	// fmt.Printf("xxxx %+v \n", colName)
	err := trackHistoryColl.Find(ctx, bson.M{"_id": id}).One(&rs)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find track_history  by id: %s", id)

		return nil, err
	}

	return &rs, err
}

func (us *MainService) FindTrackHistoryAll(ctx context.Context) ([]*pb.TrackHistory, error) {
	colName := util.FindCollectionName(HISTORY_TRACK_PREFIX, uint64(time.Now().UnixMilli()))
	trackHistoryColl = db.Collection(colName)
	rs := []*pb.TrackHistory{}
	err := trackHistoryColl.Find(ctx, bson.M{}).All(&rs)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to find track_history  all")
	}

	return rs, err
}
