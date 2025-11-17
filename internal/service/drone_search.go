package service

import (
	"context"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	util "172.21.5.249/air-trans/at-drone/internal/service/util"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"

	mongobuilder "go.jtlabs.io/mongo"
	queryoptions "go.jtlabs.io/query"

	"go.mongodb.org/mongo-driver/bson"
)

var droneSchemaBuilder = mongobuilder.NewQueryBuilder(
	DRONE,
	bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"properties": bson.M{
				"_id": bson.M{
					"bsonType": "string",
					"required": true,
				},
				"rtsp_link": bson.M{
					"bsonType": "string",
					"required": true,
				},
				"gcs_id": bson.M{
					"bsonType": "string",
					"required": true,
				},
				"drone_status": bson.M{
					"bsonType": "float",
					"required": true,
				},
				"deport_id": bson.M{
					"bsonType": "string",
					"required": true,
				},
				"mass": bson.M{
					"bsonType": "float",
					"required": true,
				},
				"max_payload": bson.M{
					"bsonType": "float",
					"required": true,
				},
				"max_battery": bson.M{
					"bsonType": "float",
					"required": true,
				},
				"max_power": bson.M{
					"bsonType": "float",
					"required": true,
				},
				"max_speed": bson.M{
					"bsonType": "float",
					"required": true,
				},
				"size": bson.M{
					"bsonType": "Array",
					"required": true,
				},
				"fixed_locker_ids": bson.M{
					"bsonType": "Array",
					"required": true,
				},
				"validate_status": bson.M{
					"bsonType": "float",
					"required": true,
				},
				"created_at": bson.M{
					"bsonType": "float",
					"required": true,
				},
				"updated_at": bson.M{
					"bsonType": "float",
					"required": true,
				},
				"created_by": bson.M{
					"bsonType": "string",
					"required": true,
				},
				"updated_by": bson.M{
					"bsonType": "string",
					"required": true,
				},
			},
		},
	},
	true,
)

func (us *MainService) SearchDrone(ctx context.Context, queryOpts queryoptions.Options) (*[]pb.Drone, int64) {
	filter, sorts, skip, limit, projection, err := util.ParseQueryOptions(droneSchemaBuilder, queryOpts)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to parse query option")

		return nil, 0
	}

	query := droneColl.Find(ctx, filter)

	count, err := query.Count()
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to count")

		return nil, 0
	}

	result := []pb.Drone{}
	err = query.Skip(skip).Limit(limit).Sort(sorts...).Select(projection).All(&result)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to query")

		return nil, 0
	}

	return &result, count
}
