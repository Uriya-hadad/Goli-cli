package instancesTypesUtils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetQueues(uri string, token string) (queues []map[string]interface{}, err error) {
	var res map[string]interface{}
	res, err = sendRequest(fmt.Sprintf("%s/v1/management/queues?count=25", uri), token, "GET")
	if err != nil {
		return nil, err
	}
	for _, item := range res["results"].([]interface{}) {
		m := item.(map[string]interface{})
		queues = append(queues, m)
	}
	return queues, err
}

func GetQueueFullDetails(uri string, queueName string, token string) (queues map[string]interface{}, err error) {
	var res map[string]interface{}
	res, err = sendRequest(fmt.Sprintf("%s/v1/management/queues/%s?messageStatistics=true", uri, queueName), token, "GET")
	if err != nil {
		return nil, err
	}
	return res, err
}

func DeleteQueue(uri string, token string, queueName string) error {
	fmt.Println("Sending request to message-queue service...")
	resp, err := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/management/queues/%s", uri, queueName), nil)
	resp.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(resp)
	defer res.Body.Close()
	if res.StatusCode > 300 {
		return fmt.Errorf("error deleting queue: %s", res.Status)
	}
	return err

}

func GetSubscribers(uri string, token string, queueName string) (subscribers []string, err error) {
	var res map[string]interface{}
	uri = fmt.Sprintf("%s/v1/management/queues/%s/subscriptions/topics?count=25", uri, queueName)
	for uri != "" {
		res, err = sendRequest(uri, token, "GET")
		if err != nil {
			return nil, err
		}
		for _, item := range res["results"].([]interface{}) {
			m := item.(map[string]interface{})
			subscribers = append(subscribers, m["topic"].(string))
		}
		if res["nextPageUri"] != nil {
			uri = res["nextPageUri"].(string)
		} else {
			uri = ""
		}
	}
	return subscribers, err
}

func sendRequest(uri string, token string, method string) (map[string]interface{}, error) {
	fmt.Println("Sending request to message-queue service...")
	resp, err := http.NewRequest(method, uri, nil)
	resp.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		return nil, err
	}
	resRaw, err := http.DefaultClient.Do(resp)
	if err != nil {
		return nil, err
	} else if resRaw.StatusCode > 300 {
		return nil, fmt.Errorf("error sending request to message-queue service: %s", resRaw.Status)
	}
	body, err := io.ReadAll(resRaw.Body)
	if err != nil {
		return nil, err
	}
	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	return res, err
}
