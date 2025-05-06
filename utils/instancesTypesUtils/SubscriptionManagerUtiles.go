package instancesTypesUtils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetAllSmsSubscriptions(uri string, token string) ([]map[string]interface{}, error) {
	var allSubscriptions []map[string]interface{}

	client := &http.Client{}
	nextPage := true
	page := 1

	for nextPage {
		fmt.Printf("Fetching page %d...\n", page)
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/subscription-manager/v1/subscriptions?page=%d", uri, page), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode > 300 {
			return nil, fmt.Errorf("error getting subscriptions: %s", res.Status)
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		var saasRegistryRes struct {
			Subscriptions []map[string]interface{} `json:"subscriptions"`
			MorePages     bool                     `json:"morePages"`
		}
		err = json.Unmarshal(body, &saasRegistryRes)
		if err != nil {
			return nil, err
		}

		allSubscriptions = append(allSubscriptions, saasRegistryRes.Subscriptions...)
		nextPage = saasRegistryRes.MorePages
		page++
	}

	return allSubscriptions, nil
}

func UpdateSmsSubscription(uri string, tenantId string, token string) error {
	resp, err := http.NewRequest("PATCH", fmt.Sprintf("%s/subscription-manager/v1/subscriptions/%s", uri, tenantId), nil)
	resp.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(resp)
	if err != nil {
		return err
	} else if res.StatusCode > 300 {
		return fmt.Errorf("error updating subscription: %s", res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode == 202 {
		return nil
	}
	return fmt.Errorf("error updating subscription: %s", body)
}

func GetSmsSubscription(uri, tenantId, token string) (map[string]interface{}, error) {
	saasRegistryRes := struct {
		Subscriptions []map[string]interface{} `json:"subscriptions"`
	}{}
	resp, err := http.NewRequest("GET", fmt.Sprintf("%s/subscription-manager/v1/subscriptions?app_tid=%s", uri, tenantId), nil)
	resp.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(resp)
	if err != nil {
		return nil, err
	} else if res.StatusCode > 300 {
		return nil, fmt.Errorf("error updating subscription: %s", res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &saasRegistryRes)
	if saasRegistryRes.Subscriptions == nil || len(saasRegistryRes.Subscriptions) == 0 {
		return nil, fmt.Errorf("subscription not found")
	}
	return saasRegistryRes.Subscriptions[0], err
}
