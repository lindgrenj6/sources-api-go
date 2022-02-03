package service

import (
	"encoding/json"
	"fmt"

	"github.com/RedHatInsights/sources-api-go/dao"
	m "github.com/RedHatInsights/sources-api-go/model"
	"github.com/redhatinsights/sources-superkey-worker/superkey"
	"gorm.io/datatypes"
)

// loads up the application as well as the associates we need for the superkey
// request
func loadApplication(application *m.Application) (*m.Application, error) {
	var appDao dao.ApplicationDao = &dao.ApplicationDaoImpl{TenantID: &application.TenantID}

	// re-pulling it from the db to make sure we have the full-version, as well
	// as preloading any relations necessary.
	app, err := appDao.GetByIdWithPreload(&application.ID, "Source", "Source.Tenant")
	if err != nil {
		return nil, err
	}

	return app, nil
}

// returns the superkey steps from the metadata table for the specific
// application type
func getApplicationSuperkeyMetaData(application *m.Application) ([]superkey.Step, error) {
	// fetch the metadata from the db (no tenancy required)
	var mDB dao.MetaDataDao = &dao.MetaDataDaoImpl{}
	metadata, err := mDB.GetSuperKeySteps(application.ApplicationTypeID)
	if err != nil {
		return nil, err
	}

	steps := make([]superkey.Step, len(metadata))

	// parse the data brought back from the db into the superkey "step" struct
	for i, step := range metadata {
		substitutions := make(map[string]string)
		err := json.Unmarshal(step.Substitutions, &substitutions)
		if err != nil {
			return nil, err
		}

		steps[i] = superkey.Step{
			Step:          step.Step,
			Name:          step.Name,
			Payload:       string(step.Payload),
			Substitutions: substitutions,
		}
	}

	return steps, nil
}

// returns any extra values for the superkey provider
func getExtraValues(application *m.Application, provider string) (map[string]string, error) {
	extra := make(map[string]string)

	switch provider {
	case "amazon":
		// fetch the account number for replacing in the iam payloads
		var mDB dao.MetaDataDao = &dao.MetaDataDaoImpl{}
		acct, err := mDB.GetSuperKeyAccountNumber(application.ApplicationTypeID)
		if err != nil {
			return nil, err
		}
		extra["account"] = acct

		// fetch the result_type for the application_type
		var atDB dao.ApplicationTypeDao = &dao.ApplicationTypeDaoImpl{}
		authType, err := atDB.GetSuperKeyResultType(application.ApplicationTypeID, provider)
		if err != nil {
			return nil, err
		}
		extra["result_type"] = authType

	default:
		return nil, fmt.Errorf("invalid provider for superkey %v", provider)
	}

	return extra, nil
}

// returns the "super key" e.g. the authentication used to communicate with the
// provider
func getSuperKeyAuthentication(application *m.Application) (*m.Authentication, error) {
	var authDao dao.AuthenticationDao = &dao.AuthenticationDaoImpl{TenantID: &application.TenantID}

	// fetch auths for this source
	auths, _, err := authDao.ListForSource(application.SourceID, 100, 0, nil)
	if err != nil {
		return nil, err
	}

	// loop through, finding the source that is "attached" to the application's
	// source and has the right authtype for superkey. This will need to be
	// updated if we ever do superkey for other cloud types/authtypes
	for i, auth := range auths {
		// TODO: parameterize this if we need superkey on something OTHER than amazon.
		if auth.ResourceID == application.SourceID && auth.AuthType == "access_key_secret_key" {
			return &auths[i], nil
		}
	}

	return nil, fmt.Errorf("superkey authentication not found")
}

type superKeyData struct {
	GUID           string
	Provider       string
	StepsCompleted map[string]map[string]string
}

func parseSuperKeyData(data datatypes.JSON) (*superKeyData, error) {
	superkeyData := make(map[string]interface{})
	err := json.Unmarshal(data, &superkeyData)
	if err != nil {
		return nil, err
	}

	guid, ok := superkeyData["guid"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid type for guid %v", superkeyData["guid"])
	}

	provider, ok := superkeyData["provider"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid type for provider %v", superkeyData["provider"])
	}

	stepsCompleted, ok := superkeyData["steps_completed"].(map[string]map[string]string)
	if !ok {
		return nil, fmt.Errorf("invalid type for steps_completed %v", superkeyData["steps_completed"])
	}

	return &superKeyData{
		GUID:           guid,
		Provider:       provider,
		StepsCompleted: stepsCompleted,
	}, nil
}
