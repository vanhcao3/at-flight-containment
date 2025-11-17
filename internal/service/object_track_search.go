package service

import (
	"context"
	"time"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	util "172.21.5.249/air-trans/at-drone/internal/service/util"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"

	mongobuilder "go.jtlabs.io/mongo"
	queryoptions "go.jtlabs.io/query"
	"go.mongodb.org/mongo-driver/bson"
)

var objectTrackSchemaBuilder = mongobuilder.NewQueryBuilder(
	util.FindCollectionName(OBJECT_TRACK, uint64(time.Now().UnixMilli())),
	bson.M{
		"bsonType": "object",
		"properties": bson.M{
			"_id": bson.M{
				"bsonType": "string",
				"required": true,
			},
			"object_track_id": bson.M{
				"bsonType": "float",
				"required": true,
			},
			"object_id": bson.M{
				"bsonType": "string",
				"required": true,
			},
			"speed": bson.M{
				"bsonType": "Object",
				"structure": bson.M{
					"vx": bson.M{
						"bsonType": "float",
						"required": true,
					},
					"vy": bson.M{
						"bsonType": "float",
						"required": true,
					},
					"vz": bson.M{
						"bsonType": "float",
						"required": true,
					},
				},
				"required": true,
			},
			"heading": bson.M{
				"bsonType": "float",
				"required": true,
			},
			"battery": bson.M{
				"bsonType": "float",
				"required": true,
			},
			"datasource_latest_update": bson.M{
				"bsonType": "Object",
				"structure": bson.M{
					"additionalProp1": bson.M{
						"bsonType": "float",
						"required": true,
					},
					"additionalProp2": bson.M{
						"bsonType": "float",
						"required": true,
					},
					"additionalProp3": bson.M{
						"bsonType": "float",
						"required": true,
					},
				},
				"required": true,
			},
			"position": bson.M{
				"bsonType": "Object",
				"structure": bson.M{
					"latitude": bson.M{
						"bsonType": "float",
						"required": true,
					},
					"longitude": bson.M{
						"bsonType": "float",
						"required": true,
					},
					"altitude": bson.M{
						"bsonType": "float",
						"required": true,
					},
				},
				"required": true,
			},
			"drone_status": bson.M{
				"bsonType": "float",
				"required": true,
			},
			"remain_time": bson.M{
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

	true,
)

func (us *MainService) SearchObjectTrack(ctx context.Context, queryOpts queryoptions.Options) (*[]*pb.ObjectTrack, int64) {

	filter, sorts, skip, limit, projection, err := util.ParseQueryOptions(objectTrackSchemaBuilder, queryOpts)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to parse query option")

		return nil, 0
	}
	query := objectTrackColl.Find(ctx, filter)

	count, err := query.Count()
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to count")

		return nil, 0
	}
	result := []*pb.ObjectTrack{}
	err = query.Skip(skip).Limit(limit).Sort(sorts...).Select(projection).All(&result)

	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to query")

		return nil, 0
	}

	return &result, count
}
