package instanceUtils

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"golang.org/x/oauth2"
	"goli-cli/types"
	"goli-cli/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func generateOauthClientToken(credentials types.ClientCredentials, subdomain string) (string, error) {
	if subdomain != "current" {
		startIndex := strings.Index(credentials.URL, "://") + 3
		endIndex := strings.Index(credentials.URL, ".")
		if endIndex == -1 || startIndex == -1 {
			return "", errors.New("invalid url")
		}
		credentials.URL = credentials.URL[:startIndex] + subdomain + credentials.URL[endIndex:]
	}
	urlStr := credentials.URL + "/oauth/token"

	v := url.Values{}
	v.Set("grant_type", "client_credentials")
	v.Set("response_type", "token")
	v.Set("client_id", credentials.Clientid)
	v.Set("client_secret", credentials.Clientsecret)

	hc := oauth2.NewClient(context.Background(), nil)
	resp, err := hc.PostForm(urlStr, v)
	if err != nil {
		return "", err
	} else if resp.StatusCode > 300 {
		return "", errors.New("error getting token: " + resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var tokenRes struct {
		AccessToken string `json:"access_token"`
		//TokenType   string `json:"token_type"`
		//IDToken     string `json:"id_token"`
		//ExpiresIn   int64  `json:"expires_in"` // relative seconds from now
	}

	err = json.Unmarshal(body, &tokenRes)
	if err != nil {
		return "", err
	}

	return tokenRes.AccessToken, nil

}

func generateX509ClientToken(credentials types.ClientCredentials, subdomain string) (string, error) {
	if subdomain != "current" {
		startIndex := strings.Index(credentials.CertUrl, "://") + 3
		endIndex := strings.Index(credentials.CertUrl, ".")
		if endIndex == -1 || startIndex == -1 {
			return "", errors.New("invalid url")
		}
		credentials.CertUrl = credentials.CertUrl[:startIndex] + subdomain + credentials.CertUrl[endIndex:]
	}
	urlStr := credentials.CertUrl + "/oauth/token"

	v := url.Values{}
	v.Set("grant_type", "client_credentials")
	v.Set("response_type", "token")
	v.Set("client_id", credentials.Clientid)

	cert, _ := tls.X509KeyPair([]byte(strings.ReplaceAll(credentials.Certificate, "\\n", "\n")), []byte(strings.ReplaceAll(credentials.Key, "\\n", "\n")))

	clientD := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	resp, err := clientD.PostForm(urlStr, v)
	if err != nil {
		return "", err
	} else if resp.StatusCode > 300 {
		return "", errors.New("error getting token: " + resp.Status)

	}

	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var tokenRes struct {
		AccessToken string `json:"access_token"`
		//TokenType   string `json:"token_type"`
		//IDToken     string `json:"id_token"`
		//ExpiresIn   int64  `json:"expires_in"` // relative seconds from now
	}

	err = json.Unmarshal(body, &tokenRes)
	if err != nil {
		return "", err
	}

	return tokenRes.AccessToken, nil
}

func UnBindService(cf *client.Client, appGUID string, instanceGUID string) error {
	credBindRes, err := cf.ServiceCredentialBindings.Single(context.Background(), &client.ServiceCredentialBindingListOptions{
		AppGUIDs:             client.Filter{Values: []string{appGUID}},
		ServiceInstanceGUIDs: client.Filter{Values: []string{instanceGUID}},
	})
	if err != nil {
		return err
	}
	jobId, err := cf.ServiceCredentialBindings.Delete(context.Background(), credBindRes.GUID)
	if err != nil {
		return err
	}
	fmt.Println("job created: ", jobId)
	err = cf.Jobs.PollComplete(context.Background(), jobId, nil)
	if err != nil {
		return err
	}
	fmt.Println("service unbound!")
	return nil
}

func CreateKeyForCred(cf *client.Client, instanceGUID string, x509 bool) (map[string]interface{}, string, error) {
	fmt.Println("Creating key for credentials...")
	keyName := "Cli - " + time.Now().String()
	err := CreateKey(cf, keyName, x509, instanceGUID)
	if err != nil {
		return nil, "", err
	}
	key, keyGUID, err := GetKey(cf, keyName, instanceGUID)
	fmt.Println("Key created successfully - ", keyGUID)
	return key.Credentials, keyGUID, err
}

func CreateKey(cf *client.Client, keyName string, x509Key bool, instanceGUID string) error {
	data := []byte("{}")
	if x509Key {
		data = []byte(`{"certificate":true,"xsuaa":{"credential-type":"x509","x509":{"key-length":2048,"validity":365,"validity-type":"DAYS"}}}`)
	}
	jobId, _, err := cf.ServiceCredentialBindings.Create(context.Background(), &resource.ServiceCredentialBindingCreate{
		Name:       &keyName,
		Parameters: (*json.RawMessage)(&data),
		Relationships: resource.ServiceCredentialBindingRelationships{
			ServiceInstance: &resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: instanceGUID,
				},
			},
		},
		Type: "key",
	})
	if err != nil {
		return err
	}
	err = cf.Jobs.PollComplete(context.Background(), jobId, nil)
	if err != nil {
		return err
	}
	return nil
}

func GetKey(cf *client.Client, keyName string, instanceGUID string) (*resource.ServiceCredentialBindingDetails, string, error) {
	keyRaw, err := cf.ServiceCredentialBindings.Single(context.Background(), &client.ServiceCredentialBindingListOptions{
		Type:                 client.Filter{Values: []string{"key"}},
		Names:                client.Filter{Values: []string{keyName}},
		ServiceInstanceGUIDs: client.Filter{Values: []string{instanceGUID}},
	})
	if err != nil {
		return nil, "", err
	}
	key, err := cf.ServiceCredentialBindings.GetDetails(context.Background(), keyRaw.GUID)
	if err != nil {
		return nil, "", err
	}

	return key, keyRaw.GUID, err
}

func DeleteKey(cf *client.Client, keyGUID string) error {
	jobId, err := cf.ServiceCredentialBindings.Delete(context.Background(), keyGUID)
	if err != nil {
		return err
	}
	err = cf.Jobs.PollComplete(context.Background(), jobId, nil)
	if err != nil {
		return err
	}
	return nil
}

func GetBoundApps(cf *client.Client, instanceGUID string) ([]types.CFAppData, error) {
	opts := &client.ServiceCredentialBindingListOptions{
		ServiceInstanceGUIDs: client.Filter{Values: []string{instanceGUID}},
		ListOptions:          client.NewListOptions(),
	}
	var apps []types.CFAppData
	for {
		_, appsObj, pager, err := cf.ServiceCredentialBindings.ListIncludeApps(context.Background(), opts)
		if err != nil {
			return nil, err
		}
		for _, app := range appsObj {
			apps = append(apps, types.CFAppData{
				Name:  app.Name,
				GUID:  app.GUID,
				State: app.State,
			})
		}
		if !pager.HasNextPage() {
			if len(apps) == 0 {
				return nil, errors.New("no bound apps")
			}
			return apps, err
		}
		pager.NextPage(opts)
	}
}

func GenerateClientToken(cred map[string]interface{}, subdomain string) (string, error) {
	var token string
	var err error
	serviceCred := getCredentialsForToken(cred)
	if serviceCred.Clientid == "" {
		return "", errors.New("cannot generate token from this credentials")
	}
	if serviceCred.Clientsecret != "" {
		token, err = generateOauthClientToken(serviceCred, subdomain)
	} else {
		token, err = generateX509ClientToken(serviceCred, subdomain)
	}
	return token, err
}

func getCredentialsForToken(cred map[string]interface{}) types.ClientCredentials {
	tokenCredentials := types.ClientCredentials{
		Clientid:     utils.GetStringFromMap(cred, "clientid"),
		Clientsecret: utils.GetStringFromMap(cred, "clientsecret"),
		URL:          utils.GetStringFromMap(cred, "url"),
		CertUrl:      utils.GetStringFromMap(cred, "certurl"),
		Certificate:  utils.GetStringFromMap(cred, "certificate"),
		Key:          utils.GetStringFromMap(cred, "key"),
	}
	return tokenCredentials
}

func GetCredFromBinding(cf *client.Client, guid string) (map[string]interface{}, error) {
	boundApp := GetFirstStartedApp(cf, guid, false)
	if boundApp == nil {
		return nil, errors.New("no app with started state is exist")
	}

	bindInfo, err := cf.ServiceCredentialBindings.Single(context.Background(), &client.ServiceCredentialBindingListOptions{
		ServiceInstanceGUIDs: client.Filter{Values: []string{guid}},
		AppGUIDs:             client.Filter{Values: []string{boundApp.GUID}},
	})
	if err != nil {
		return nil, err
	}
	BindDetails, err := cf.ServiceCredentialBindings.GetDetails(context.Background(), bindInfo.Resource.GUID)
	if err != nil {
		return nil, err
	}
	if BindDetails.Credentials == nil {
		return make(map[string]interface{}), nil
	}
	return BindDetails.Credentials, err
}

func GetFirstStartedApp(cf *client.Client, guid string, withSSHEnabled bool) *types.CFAppData {
	boundApps, err := GetBoundApps(cf, guid)
	if err != nil {
		return nil
	}
	var startedApp *types.CFAppData
	for _, app := range boundApps {
		if app.State == "STARTED" {
			startedApp = &app
			if withSSHEnabled {
				sshState, _ := cf.Applications.SSHEnabled(context.Background(), app.GUID)
				if sshState != nil && sshState.Enabled {
					return &app
				}
			} else {
				break
			}
		}
	}
	return startedApp // Return nil if no app is found
}
