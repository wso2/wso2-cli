package customDomain

import (
	"encoding/json"

	workflowmgt "github.com/wso2/integration-platform-tools/pkg/api/workflow-mgt"
)

type CustomDomain struct {
	ID          string                  `json:"id"`
	OrgID       string                  `json:"orgId"`
	Name        string                  `json:"name"`
	Type        string                  `json:"type"`
	Visibility  string                  `json:"visibility"`
	Environment CustomDomainEnvironment `json:"environment"`
	TLSProvider struct {
		Name string `json:"name"`
	} `json:"tlsProvider"`
	IsAvailable     bool `json:"isAvailable"`
	EnableAutoApply bool `json:"enableAutoApply"`
}

type CustomDomainEnvironment struct {
	TemplateID string
}

func (e *CustomDomainEnvironment) UnmarshalJSON(data []byte) error {
	type envAlias struct {
		ID string `json:"id"`
	}
	var alias envAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	e.TemplateID = alias.ID
	return nil
}

func (e CustomDomainEnvironment) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TemplateID string `json:"template_id"`
	}{
		TemplateID: e.TemplateID,
	})
}

type CustomDomainValidateReq struct {
	DomainName    string `json:"domain"`
	EnvID         string `json:"environmentId"`
	ComponentType string `json:"type"`
}

type CustomDomainValidateResp struct {
	IsValid bool   `json:"isValid"`
	Message string `json:"message"`
}

type CustomDomainCreateReq struct {
	DomainName  string `json:"name"`
	Environment struct {
		ID string `json:"id"`
	} `json:"environment"`
	ComponentType string `json:"type"`
	TLSProvider   struct {
		Name string `json:"name"`
	} `json:"tlsProvider"`
	EnableAutoApply bool `json:"enableAutoApply"`
}

type CreateUrlMappingReq struct {
	DomainID      string                                       `json:"domainId"`
	ComponentID   string                                       `json:"componentId"`
	ApiID         string                                       `json:"apiId"`
	DefaultDomain string                                       `json:"defaultDomain"`
	CustomPath    string                                       `json:"customPath"`
	DefaultPath   string                                       `json:"defaultPath"`
	MetaData      *workflowmgt.RequestUrlCustomizationMetadata `json:"metadata,omitempty"`
}

type CreateUrlMappingRes struct {
	ID            string `json:"id"`
	DomainID      string `json:"domainId"`
	DefaultPath   string `json:"defaultPath"`
	CustomPath    string `json:"customPath"`
	DefaultDomain string `json:"defaultDomain"`
	ComponentID   string `json:"componentId"`
	ApiID         string `json:"apiId"`
	Status        string `json:"status"`
	CreatedBy     string `json:"createdBy"`
}
