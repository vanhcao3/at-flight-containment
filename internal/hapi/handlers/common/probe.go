package common

import (
	"github.com/qiniu/qmgo"
)

func ProbeReadiness(dbClient *qmgo.Client) error {
	dbErr := dbClient.Ping(5)

	if dbErr != nil {
		return dbErr
	}
	return nil
}
