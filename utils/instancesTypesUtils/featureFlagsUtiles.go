package instancesTypesUtils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type FeatureFlag struct {
	ID                    string   `json:"id"`
	Description           string   `json:"description"`
	DirectShipments       []any    `json:"directShipments"`
	WeightedChoices       []any    `json:"weightedChoices"`
	Variations            []string `json:"variations"`
	OffVariationIndex     int      `json:"offVariationIndex"`
	VariationType         string   `json:"variationType"`
	DefaultVariationIndex int      `json:"defaultVariationIndex"`
	Enabled               bool     `json:"enabled"`
}

func GetFeatureFlags(uri string, token string) (featuresFlags []FeatureFlag, err error) {
	resp, err := http.NewRequest("GET", uri, nil)
	resp.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(resp)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	} else if res.StatusCode > 300 {
		return nil, fmt.Errorf("error getting feature flags: %s", res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &featuresFlags)
	return featuresFlags, err
}
