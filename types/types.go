package types

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"time"
)

type LandscapeData map[string]OrgData
type OrgData map[string]SpaceData
type SpaceData map[string]*CliData
type CliData struct {
	Apps      map[string]AppData
	Instances *OfferData
}
type AppData struct {
	Name string `json:"name"`
	GUID string `json:"guid"`
}
type OfferData map[string][]InstanceData

type InstanceData struct {
	Name string `json:"name"`
	GUID string `json:"guid"`
	Plan string `json:"plan"`
}

type Landscape map[string][]*CfOrg

type CFAppData struct {
	Name  string `json:"Name"`
	GUID  string `json:"Guid"`
	State string `json:"State"`
}

type ConnectionInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hostname string `json:"hostname"`
	Dbname   string `json:"dbname"`
	Port     string `json:"port"`
}

type CfUser struct {
	Email string `json:"user_name"`
	Role  string `json:"role"`
}

type CfSpace struct {
	Name string
	GUID string
}

type CfOrg struct {
	Name   string
	GUID   string
	Spaces []*CfSpace
}

type InstancesResponse struct {
	Pagination struct {
		TotalResults int `json:"total_results"`
		TotalPages   int `json:"total_pages"`
		First        struct {
			Href string `json:"href"`
		} `json:"first"`
		Last struct {
			Href string `json:"href"`
		} `json:"last"`
		Next     interface{} `json:"next"`
		Previous interface{} `json:"previous"`
	} `json:"pagination"`
	Resources []struct {
		GUID          string        `json:"guid"`
		CreatedAt     time.Time     `json:"created_at"`
		UpdatedAt     time.Time     `json:"updated_at"`
		Name          string        `json:"name"`
		Tags          []interface{} `json:"tags"`
		LastOperation struct {
			Type        string    `json:"type"`
			State       string    `json:"state"`
			Description string    `json:"description"`
			UpdatedAt   time.Time `json:"updated_at"`
			CreatedAt   time.Time `json:"created_at"`
		} `json:"last_operation"`
		Type             string `json:"type"`
		UpgradeAvailable bool   `json:"upgrade_available"`
		DashboardURL     string `json:"dashboard_url"`
		Relationships    struct {
			Space struct {
				Data struct {
					GUID string `json:"guid"`
				} `json:"data"`
			} `json:"space"`
			ServicePlan struct {
				Data struct {
					GUID string `json:"guid"`
				} `json:"data"`
			} `json:"service_plan"`
		} `json:"relationships"`
	} `json:"resources"`
	Included struct {
		ServicePlans []struct {
			GUID          string `json:"guid"`
			Name          string `json:"name"`
			Relationships struct {
				ServiceOffering struct {
					Data struct {
						GUID string `json:"guid"`
					} `json:"data"`
				} `json:"service_offering"`
			} `json:"relationships"`
		} `json:"service_plans"`
		ServiceOfferings []struct {
			Name          string `json:"name"`
			GUID          string `json:"guid"`
			Relationships struct {
				ServiceBroker struct {
					Data struct {
						GUID string `json:"guid"`
					} `json:"data"`
				} `json:"service_broker"`
			} `json:"relationships"`
		} `json:"service_offerings"`
	} `json:"included"`
}

type ClientCredentials struct {
	Clientid     string
	Clientsecret string
	URL          string
	CertUrl      string
	Certificate  string
	Key          string
}

type LocalConfig struct {
	PresentByTeams bool   `json:"presentByTeams"`
	IUA            bool   `json:"iua"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	MigVar         string `json:"migVar"`
	DbCred         struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"dbCredentials"`
}
type UserInfo struct {
	Email string `json:"mail"`
	Role  string `json:"role"`
	Auth  bool   `json:"auth"`
}

type ManagedInstance interface {
	GetGUID() string
	GetName() string
	GetCredentials(cf *client.Client) (map[string]interface{}, error)
	GetBoundDetails(cf *client.Client) (map[string]interface{}, error)
	GetToken(cf *client.Client, subdomain string) (string, error)
	SetToken(subdomain, token string)
	ListOptions(cf *client.Client)
	CleanUp(cf *client.Client)
}
