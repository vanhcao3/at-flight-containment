package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"time"

	mongobuilder "go.jtlabs.io/mongo"
	queryoptions "go.jtlabs.io/query"
	"google.golang.org/protobuf/proto"

	"172.21.5.249/air-trans/at-drone/internal/config"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/***************************************************************************************************************/

/* Create HTTP search options */
func CreateSearchOptions(filter map[string][]string, page int32, pageSize int32) *pb.SearchOptions {
	gFilter := map[string]*pb.StringList{}
	for k, v := range filter {
		gFilter[k] = &pb.StringList{
			Str: v,
		}
	}

	return &pb.SearchOptions{
		Page: map[string]int32{
			"page": page,
			"size": pageSize,
		},
		Filter: gFilter,
		Fields: []string{},
		Sorts:  []string{},
	}
}

type PatchOption struct {
	Op    string      `json:"op,omitempty"`
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

/* Create HTTP patch options */
func CreatePatchOptions(patchOptions []PatchOption) []byte {
	res, _ := json.Marshal(patchOptions)

	return res
}

/* Parse HTTP query options */
func ParseQueryOptions(schemaBuilder *mongobuilder.QueryBuilder, queryOpts queryoptions.Options) (
	primitive.M,
	[]string,
	int64,
	int64,
	interface{},
	error,
) {
	filter, err := schemaBuilder.Filter(queryOpts, mongobuilder.And)
	if err != nil {
		return nil, nil, 0, 0, nil, err
	}

	opts, err := schemaBuilder.FindOptions(queryOpts)
	if err != nil {
		return nil, nil, 0, 0, nil, err
	}
	var projection interface{}
	if opts.Projection != nil {
		projection = opts.Projection
	}
	var skip int64 = 0
	var limit int64 = 999999999999

	if opts.Skip != nil {
		skip = *opts.Skip
	}

	if opts.Limit != nil {
		limit = *opts.Limit
	}

	return filter, queryOpts.Sort, skip, limit, projection, nil
}

/* Create publish event data */
func CreatePublishEventData[T any, U any](requestData T, responseData U) []byte {
	request, _ := json.Marshal(requestData)
	response, _ := json.Marshal(responseData)

	data, _ := json.Marshal(
		map[string][]byte{
			"request":  request,
			"response": response,
		},
	)

	return data
}

/***************************************************************************************************************/

/* Replace escape HTML string */
func ReplaceEscapeHTMLString(espHTMLString string) string {
	for old, new := range map[string]string{
		`\u003c`: "<",
		`\u003e`: ">",
		`\u0026`: "&",
		`\u0028`: "U+2028",
		`\u0029`: "U+2029",
	} {
		espHTMLString = strings.ReplaceAll(espHTMLString, old, new)
	}

	return espHTMLString
}

/* Retry do function */
func Retry(
	ctx context.Context,
	fn func(...interface{}) error,
	retries int,
	tryAfter int,
	args ...interface{},
) error {
	for attempt := 1; attempt <= retries; attempt++ {
		err := fn(args...)
		if err == nil {
			return nil
		}

		remainingAttempts := retries - attempt
		if remainingAttempts > 0 {
			PrintErrorLog(ctx, err, "%d retries remaining - After %d ms...", remainingAttempts, tryAfter)

			time.Sleep(time.Duration(tryAfter) * time.Millisecond)
		} else {
			return err
		}
	}

	return nil
}

/* Run external command */
func RunCommand(ctx context.Context, options CommandOption) (string, string, error) {
	// Prepare command
	cmd := exec.CommandContext(ctx, options.Command, options.Options...)

	// Redirect stdin
	if options.Input != "" {
		// Log input
		if options.Logging {
			PrintDebugLog(ctx, "Exec input: %s", options.Input)
		}

		cmd.Stdin = strings.NewReader(options.Input)
	}

	// Redirect stdout, stderr
	stdOut, stdErr := bytes.Buffer{}, bytes.Buffer{}
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	// Log command
	if options.Logging {
		PrintDebugLog(ctx, "Exec command: %s", cmd.String())
	}

	// Call callback interval
	if options.Interval > 0 && options.Callback != nil {
		ticker := time.NewTicker(time.Duration(options.Interval) * time.Millisecond)
		defer ticker.Stop()

		go func() {
			for {
				select {
				case <-ticker.C:
					{
						options.Callback(stdOut.String())
					}

				case <-ctx.Done():
					{
						ticker.Stop()
					}
				}
			}
		}()
	}

	// Run command
	err := cmd.Run()
	if err != nil {
		if len(stdErr.Bytes()) > 0 {
			err = fmt.Errorf(stdErr.String())
		}
	}

	// Last call callback
	if options.Interval > 0 && options.Callback != nil {
		options.Callback(stdOut.String())
	}

	// Log error/warning & output
	if options.Logging {
		if len(stdErr.Bytes()) > 0 {
			if err != nil {
				PrintErrorLog(ctx, err, "Exec error: %s", stdErr.String())
			} else {
				PrintWarningLog(ctx, "Exec warning: %s", stdErr.String())
			}
		}

		PrintDebugLog(ctx, "Exec output: %s", stdOut.String())
	}

	return cmd.String(), stdOut.String(), err
}

/***************************************************************************************************************/

func FindCollectionName(prefix string, unixMs uint64) string {
	t := time.UnixMilli(int64(unixMs))
	return fmt.Sprintf("%s_%s", prefix, t.Format("020106"))
}

func KeepNumbers(input string) string {
	re := regexp.MustCompile(`[^0-9]+`)
	return re.ReplaceAllString(input, "")
}

type TrackHistoryByteAndJson struct {
	LocationJson *pb.Location     `json:"location_json"`
	TrackHistory *pb.TrackHistory `json:"track_history"`
}

func ConvertToJSONResponse(ctx context.Context, u []*pb.TrackHistory) []TrackHistoryByteAndJson {
	res := []TrackHistoryByteAndJson{}
	for i := range u {
		var droneLocation pb.Location
		if len((u)[i].LocationByte) != 0 {
			if err := proto.Unmarshal(u[i].LocationByte, &droneLocation); err != nil {
				config.PrintErrorLog(ctx, err, "Failed unmarshall location byte")
			}

		}
		res = append(res, TrackHistoryByteAndJson{
			LocationJson: &droneLocation,
			TrackHistory: (u)[i],
		})
	}
	return res
}

func GetDateFromTimestamp(timestamp int64) time.Time {
	// Assuming timestamp is in milliseconds; adjust if needed
	return time.Unix(timestamp/1000, (timestamp%1000)*1e6).Add(7 * time.Hour)
	// If hardcoded, do not add 7 hours more
}

// FindInvalidFloats recursively finds all fields with Inf or NaN values
func FindInvalidFloats(v interface{}, path string) []string {
	var problems []string
	val := reflect.ValueOf(v)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return problems
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := typ.Field(i)

			// Skip unexported fields
			if !field.CanInterface() {
				continue
			}

			fieldPath := path + "." + fieldType.Name
			problems = append(problems, FindInvalidFloats(field.Interface(), fieldPath)...)
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			item := val.Index(i)
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			problems = append(problems, FindInvalidFloats(item.Interface(), itemPath)...)
		}

	case reflect.Map:
		iter := val.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			mapPath := fmt.Sprintf("%s[%v]", path, k.Interface())
			problems = append(problems, FindInvalidFloats(v.Interface(), mapPath)...)
		}

	case reflect.Float32, reflect.Float64:
		f := val.Float()
		if math.IsInf(f, 0) {
			problems = append(problems, fmt.Sprintf("%s = +Inf", path))
		} else if math.IsInf(f, -1) {
			problems = append(problems, fmt.Sprintf("%s = -Inf", path))
		} else if math.IsNaN(f) {
			problems = append(problems, fmt.Sprintf("%s = NaN", path))
		}
	}

	return problems
}
