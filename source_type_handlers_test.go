package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/RedHatInsights/sources-api-go/internal/testutils/fixtures"
	"github.com/RedHatInsights/sources-api-go/internal/testutils/helpers"
	"github.com/RedHatInsights/sources-api-go/internal/testutils/request"
	"github.com/RedHatInsights/sources-api-go/internal/testutils/templates"
	m "github.com/RedHatInsights/sources-api-go/model"
	"github.com/RedHatInsights/sources-api-go/util"
)

func TestSourceTypeList(t *testing.T) {
	c, rec := request.CreateTestContext(
		http.MethodGet,
		"/api/sources/v3.1/source_types",
		nil,
		map[string]interface{}{
			"limit":   100,
			"offset":  0,
			"filters": []util.Filter{},
		},
	)

	err := SourceTypeList(c)

	if err != nil {
		t.Error(err)
	}

	if rec.Code != 200 {
		t.Error("Did not return 200")
	}

	var out util.Collection

	err = json.Unmarshal(rec.Body.Bytes(), &out)
	if err != nil {
		t.Error("Failed unmarshaling output")
	}

	if out.Meta.Limit != 100 {
		t.Error("limit not set correctly")
	}

	if out.Meta.Offset != 0 {
		t.Error("offset not set correctly")
	}

	if len(out.Data) != len(fixtures.TestSourceTypeData) {
		t.Error("not enough objects passed back from DB")
	}

	for _, sourceType := range out.Data {
		s, ok := sourceType.(map[string]interface{})
		if !ok {
			t.Error("model did not deserialize as a application type response")
		}
		if s["id"] == 1 && s["name"] != "amazon" {
			t.Error("ghosts infected the return")
		}
	}

	helpers.AssertLinks(t, c.Request().RequestURI, out.Links, 100, 0)
}

func TestSourceTypeListBadRequestInvalidFilter(t *testing.T) {
	helpers.SkipIfNotRunningIntegrationTests(t)

	c, rec := request.CreateTestContext(
		http.MethodGet,
		"/api/sources/v3.1/source_types",
		nil,
		map[string]interface{}{
			"limit":  100,
			"offset": 0,
			"filters": []util.Filter{
				{Name: "wrongName", Value: []string{"wrongValue"}},
			},
		},
	)

	badRequestSourceTypeList := ErrorHandlingContext(SourceTypeList)
	err := badRequestSourceTypeList(c)
	if err != nil {
		t.Error(err)
	}

	templates.BadRequestTest(t, rec)
}

func TestSourceTypeListWithOffsetAndLimit(t *testing.T) {
	helpers.SkipIfNotRunningIntegrationTests(t)
	testData := templates.TestDataForOffsetLimitTest
	wantSourceTypesCount := len(fixtures.TestSourceTypeData)

	for _, i := range testData {
		c, rec := request.CreateTestContext(
			http.MethodGet,
			"/api/sources/v3.1/source_types",
			nil,
			map[string]interface{}{
				"limit":   i["limit"],
				"offset":  i["offset"],
				"filters": []util.Filter{},
			},
		)

		err := SourceTypeList(c)
		if err != nil {
			t.Error(err)
		}

		path := c.Request().RequestURI
		templates.WithOffsetAndLimitTest(t, path, rec, wantSourceTypesCount, i["limit"], i["offset"])
	}
}

func TestSourceTypeGet(t *testing.T) {
	c, rec := request.CreateTestContext(
		http.MethodGet,
		"/api/sources/v3.1/source_types/1",
		nil,
		map[string]interface{}{
			"tenantID": int64(1),
		})

	c.SetParamNames("id")
	c.SetParamValues("1")

	err := SourceTypeGet(c)
	if err != nil {
		t.Error(err)
	}

	if rec.Code != 200 {
		t.Error("Did not return 200")
	}

	var outSrc m.SourceResponse
	err = json.Unmarshal(rec.Body.Bytes(), &outSrc)
	if err != nil {
		t.Error("Failed unmarshaling output")
	}

	if *outSrc.Name != "amazon" {
		t.Error("ghosts infected the return")
	}
}

func TestSourceTypeGetNotFound(t *testing.T) {
	c, rec := request.CreateTestContext(
		http.MethodGet,
		"/api/sources/v3.1/source_types/3098539345",
		nil,
		map[string]interface{}{
			"tenantID": int64(1),
		},
	)

	c.SetParamNames("id")
	c.SetParamValues("3098539345")

	notFoundSourceTypeGet := ErrorHandlingContext(SourceTypeGet)
	err := notFoundSourceTypeGet(c)
	if err != nil {
		t.Error(err)
	}

	templates.NotFoundTest(t, rec)
}

func TestSourceTypeGetBadRequest(t *testing.T) {
	c, rec := request.CreateTestContext(
		http.MethodGet,
		"/api/sources/v3.1/source_types/xxx",
		nil,
		map[string]interface{}{
			"tenantID": int64(1),
		},
	)

	c.SetParamNames("id")
	c.SetParamValues("xxx")

	notFoundSourceTypeGet := ErrorHandlingContext(SourceTypeGet)
	err := notFoundSourceTypeGet(c)
	if err != nil {
		t.Error(err)
	}

	templates.BadRequestTest(t, rec)
}
