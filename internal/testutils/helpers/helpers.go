package helpers

import (
	"github.com/RedHatInsights/sources-api-go/internal/testutils/fixtures"
	"github.com/RedHatInsights/sources-api-go/internal/testutils/parser"
	"testing"
)

func GetSourcesCountWithAppType(appTypeId int64) int {
	var sourcesCount int
	var sourceIdList []int64

	for _, app := range fixtures.TestApplicationData {
		if app.ApplicationTypeID == appTypeId {
			sourceIdList = append(sourceIdList, app.SourceID)
		}
	}
	for _, sourceID := range sourceIdList {
		for _, src := range fixtures.TestSourceData {
			if src.ID == sourceID {
				sourcesCount++
			}
		}
	}
	return sourcesCount
}

func GetAppTypeCountWithSourceId(sourceId int64) int {
	var appTypeList []int64

	for _, app := range fixtures.TestApplicationData {
		if app.SourceID == sourceId {
			appTypeList = append(appTypeList, app.ApplicationTypeID)
		}
	}

	return len(appTypeList)
}

// SkipIfNotRunningIntegrationTests is a helper function which skips a test if the integration tests don't want to be
// run.
func SkipIfNotRunningIntegrationTests(t *testing.T) {
	if !parser.RunningIntegrationTests {
		t.Skip("Skipping integration test")
	}
}
