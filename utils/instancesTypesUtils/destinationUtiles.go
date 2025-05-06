package instancesTypesUtils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetDestinations(uri string, token string) (destinations []map[string]interface{}, err error) {
	fmt.Println("Sending request to destination service...")
	resp, err := http.NewRequest("GET", fmt.Sprintf("%s/destination-configuration/v1/subaccountDestinations", uri), nil)
	resp.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(resp)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	} else if res.StatusCode > 300 {
		return nil, fmt.Errorf("error getting destinations: %s", res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &destinations)
	return destinations, err
}
