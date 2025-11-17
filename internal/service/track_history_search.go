package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	util "172.21.5.249/air-trans/at-drone/internal/service/util"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"

	mongobuilder "go.jtlabs.io/mongo"
	queryoptions "go.jtlabs.io/query"

	"go.mongodb.org/mongo-driver/bson"
)

var trackHistorySchemaBuilder = mongobuilder.NewQueryBuilder(
	util.FindCollectionName(HISTORY_TRACK_PREFIX, uint64(time.Now().UnixMilli())),
	bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"properties": bson.M{
				"_id": bson.M{
					"bsonType": "string",
					"required": true,
				},
				"drone_id": bson.M{
					"bsonType": "string",
					"required": true,
				},
				"order_id": bson.M{
					"bsonType": "string",
					"required": true,
				},
				"datasource": bson.M{
					"bsonType": "string",
					"required": true,
				},
				"track_id": bson.M{
					"bsonType": "long",
					"required": true,
				},
				"location_byte": bson.M{
					"bsonType": "Object",
					"structure": bson.M{
						"_bsonbsonType": bson.M{
							"bsonType": "string",
							"required": true,
						},
						"sub_bsonType": bson.M{
							"bsonType": "float",
							"required": true,
						},
						"position": bson.M{
							"bsonType": "float",
							"required": true,
						},
						"buffer": bson.M{
							"bsonType": "Uint8Array",
							"required": true,
						},
						"put": bson.M{
							"bsonType": "function",
							"required": true,
						},
						"write": bson.M{
							"bsonType": "function",
							"required": true,
						},
						"read": bson.M{
							"bsonType": "function",
							"required": true,
						},
						"value": bson.M{
							"bsonType": "function",
							"required": true,
						},
						"length": bson.M{
							"bsonType": "function",
							"required": true,
						},
						"toJSON": bson.M{
							"bsonType": "function",
							"required": true,
						},
					},
					"required": true,
				},
				"created_at": bson.M{
					"bsonType": "long",
					"required": true,
				},
			},
		},
	},
	true,
)

func (us *MainService) SearchTrackHistory(ctx context.Context, queryOpts queryoptions.Options) ([]*pb.TrackHistory, int64) {
	colName := util.FindCollectionName(HISTORY_TRACK_PREFIX, uint64(time.Now().UnixMilli()))
	trackHistoryColl = db.Collection(colName)
	var countAll int64
	var startTime uint64
	var endTime uint64
	endTime = uint64(time.Now().Unix()) * 1000
	startTime = endTime - 60*60*24*1000
	createdAtFilter := queryOpts.Filter["created_at"]
	// updatedAtFilter := queryOpts.Filter["updated_at"]
	allResults := []*pb.TrackHistory{}
	var trackSearchEnabled bool
	for key, _ := range queryOpts.Filter {
		if key == "track_id" {
			trackSearchEnabled = true

		}

	}
	if trackSearchEnabled {

		// hard-coded for search from now to 1 hour before
		endTime = uint64(time.Now().Unix()) * 1000
		startTime = endTime - 60*60*1000
		result := []*pb.TrackHistory{}
		options := []string{">=" + fmt.Sprint(startTime), "<=" + fmt.Sprint(endTime)}
		queryOpts.Filter["created_at"] = options
		filter, sorts, skip, limit, projection, err := util.ParseQueryOptions(trackHistorySchemaBuilder, queryOpts)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to parse query option")
			return nil, 0
		}
		query := trackHistoryColl.Find(ctx, filter)

		count, err := query.Count()
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to count")

			return nil, 0
		}

		err = query.Skip(skip).Limit(limit).Sort(sorts...).Select(projection).All(&result)

		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to query")

			return nil, 0
		}

		allResults = append(allResults, result...)
		countAll += count
		return allResults, countAll

	} else if len(createdAtFilter) == 2 {
		startTime, _ = strconv.ParseUint(util.KeepNumbers(createdAtFilter[0]), 10, 64)
		endTime, _ = strconv.ParseUint(util.KeepNumbers(createdAtFilter[1]), 10, 64)
	}
	// queryOpts.Filter["created_at"] = nil

	collections := make(map[string]bool)

	startDate := util.GetDateFromTimestamp(int64(startTime))
	endDate := util.GetDateFromTimestamp(int64(endTime))

	// Normalize to start of day in UTC
	startDayUTC := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDayUTC := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

	// Iterate through each day in the range
	for d := startDayUTC; !d.After(endDayUTC); d = d.AddDate(0, 0, 1) {
		// Reset to start of day for consistent collection naming
		expectedDays := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		collName := util.FindCollectionName(HISTORY_TRACK_PREFIX, uint64(expectedDays.UnixMilli()))
		collections[collName] = true
	}
	if queryOpts.Filter["created_at"] == nil {
		options := []string{">=" + fmt.Sprint(startTime), "<=" + fmt.Sprint(endTime)}
		queryOpts.Filter["created_at"] = options
	}
	filter, sorts, skip, limit, projection, err := util.ParseQueryOptions(trackHistorySchemaBuilder, queryOpts)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to parse query option")

		return nil, 0
	}
	// Add time range to filter
	// if filter == nil {
	// 	filter = bson.M{}
	// }
	// filter["create_at"] = bson.M{
	// 	"$gt": startTs,
	// 	"$lt": endTs,
	// }

	for collName, _ := range collections {
		result := []*pb.TrackHistory{}
		trackHistoryColl = db.Collection(collName)
		config.PrintDebugLog(ctx, "filter valuess %v , time value %v ---- %v", filter, startTime, endTime)
		query := trackHistoryColl.Find(ctx, filter)

		count, err := query.Count()
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to count")

			return nil, 0
		}

		err = query.Skip(skip).Limit(limit).Sort(sorts...).Select(projection).All(&result)
		allResults = append(allResults, result...)
		countAll += count
	}

	return allResults, countAll
}
