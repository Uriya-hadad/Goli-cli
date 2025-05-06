package instancesTypes

import (
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/utils"
	"goli-cli/utils/instanceUtils"
	"goli-cli/utils/instancesTypesUtils"
	"goli-cli/utils/outputUtils"
	"sync"
)

type MessageQueueInstance struct {
	Name          string
	Plan          string
	GUID          string
	LastOperation struct {
		Type  string
		State string
	}
	keyGUID     string
	Token       map[string]string
	Credentials map[string]interface{}
	FullDetails map[string]interface{}
	getterMutex sync.Mutex
}

func (p *MessageQueueInstance) GetBoundDetails(cf *client.Client) (map[string]interface{}, error) {
	if p.FullDetails != nil {
		return p.FullDetails, nil
	}
	_, err := p.GetCredentials(cf)
	return p.FullDetails, err
}

func (p *MessageQueueInstance) GetCredentials(cf *client.Client) (map[string]interface{}, error) {
	defer p.getterMutex.Unlock()
	p.getterMutex.Lock()
	if p.Credentials != nil {
		return p.Credentials, nil
	}
	BindingDetails, err := instanceUtils.GetCredFromBinding(cf, p.GUID)
	if BindingDetails != nil {
		p.FullDetails = BindingDetails
		p.Credentials = BindingDetails["uaa"].(map[string]interface{})
		return p.Credentials, nil
	}
	BindingDetails, keyGUID, err := instanceUtils.CreateKeyForCred(cf, p.GUID, true)
	if err != nil {
		return nil, err
	}
	p.keyGUID = keyGUID
	p.FullDetails = BindingDetails
	p.Credentials = BindingDetails["uaa"].(map[string]interface{})
	return p.Credentials, nil
}

func (p *MessageQueueInstance) GetGUID() string {
	return p.GUID
}

func (p *MessageQueueInstance) GetName() string {
	return p.Name
}

func (p *MessageQueueInstance) SetToken(subdomain, token string) {
	p.Token[subdomain] = token
}

func (p *MessageQueueInstance) GetToken(cf *client.Client, subdomain string) (string, error) {
	if subdomain == "" {
		subdomain = "current"
	}
	value, ok := p.Token[subdomain]
	if ok {
		return value, nil
	}
	cred, err := p.GetCredentials(cf)
	if err != nil {
		return "", err
	}
	token, err := instanceUtils.GenerateClientToken(cred, subdomain)
	if err != nil {
		return "", err
	}
	p.Token[subdomain] = token
	return token, nil
}

func (p *MessageQueueInstance) ListOptions(cf *client.Client) {
	const GetQueue = "Get one queue"
	const GetQueues = "Get all queues"
	const DeleteQueue = "Delete a queue"
	const GetTopics = "Get all topics for a queue"
	const Back = "Back"

	options := []string{GetQueue, GetQueues, DeleteQueue, GetTopics, Back}
	for {
		option, _ := utils.ListAndSelectItem(options, "Please select an option", false)
		var err error

		switch option {
		case GetQueue:
			var queue map[string]interface{}
			queue, err = p.getQueue(cf)
			if err != nil {
				break
			}
			outputUtils.PrintColoredJSON(queue, nil, nil)
		case GetQueues:
			var queues []map[string]interface{}
			queues, err = p.getQueues(cf)
			if err != nil {
				break
			}
			outputUtils.PrintColoredJsons(queues)
		case DeleteQueue:
			err = p.deleteQueue(cf)
		case GetTopics:
			var topics []string
			topics, err = p.getSubscribers(cf)
			if err != nil {
				break
			}

			for _, topic := range topics {
				outputUtils.PrintItemsMessage(topic)
			}
		case Back:
			return
		}
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			err = nil
		}
	}
}

func (p *MessageQueueInstance) CleanUp(cf *client.Client) {
	if p.keyGUID == "" {
		return
	}
	err := instanceUtils.DeleteKey(cf, p.keyGUID)
	if err != nil {
		outputUtils.PrintWarningMessage("an error occurred while deleting the key:", err.Error())
	}
}

func (p *MessageQueueInstance) getQueues(cf *client.Client) ([]map[string]interface{}, error) {
	token, err := p.GetToken(cf, "")
	if err != nil {
		return nil, err
	}
	queues, err := instancesTypesUtils.GetQueues(p.FullDetails["management"].(map[string]interface{})["url"].(string), token)
	return queues, err
}

func (p *MessageQueueInstance) deleteQueue(cf *client.Client) error {
	var queues []map[string]interface{}
	token, err := p.GetToken(cf, "")
	if err != nil {
		return err
	}
	queues, err = instancesTypesUtils.GetQueues(p.FullDetails["management"].(map[string]interface{})["url"].(string), token)
	if err != nil {
		return err
	}
	queue, _ := utils.ListAndSelectItemMap(queues, "Select a queue to delete:", true, "name")
	err = instancesTypesUtils.DeleteQueue(p.FullDetails["management"].(map[string]interface{})["url"].(string), token, queue["name"].(string))
	if err != nil {
		return err
	}
	fmt.Printf("Queue %s deleted successfully\n", queue["name"])
	return nil
}

func (p *MessageQueueInstance) getQueue(cf *client.Client) (map[string]interface{}, error) {
	token, err := p.GetToken(cf, "")
	if err != nil {
		return nil, err
	}
	queues, err := p.getQueues(cf)
	if err != nil {
		return nil, err
	}
	queue, _ := utils.ListAndSelectItemMap(queues, "Select a queue to get details for:", true, "name")
	queueFull, err := instancesTypesUtils.GetQueueFullDetails(p.FullDetails["management"].(map[string]interface{})["url"].(string), queue["name"].(string), token)

	return queueFull, nil
}

func (p *MessageQueueInstance) getSubscribers(cf *client.Client) ([]string, error) {
	var subscribers []string
	token, err := p.GetToken(cf, "")
	if err != nil {
		return nil, err
	}
	queues, err := p.getQueues(cf)
	if err != nil {
		return nil, err
	}
	queue, _ := utils.ListAndSelectItemMap(queues, "Select a queue to get subscribers for:", true, "name")
	subscribers, err = instancesTypesUtils.GetSubscribers(p.FullDetails["management"].(map[string]interface{})["url"].(string), token, queue["name"].(string))
	if err != nil {
		return nil, err
	}
	return subscribers, nil
}
