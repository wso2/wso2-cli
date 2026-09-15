package workflowmgt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wso2/integration-platform-tools/pkg/api"
)

type WorkflowMgtClient struct {
	workflowMgtUrl string
	client         *api.IPHTTPClient
}

func NewWorkflowMgtClient(workflowMgtUrl string, tokenStore api.ReadOnlyTokenStore) *WorkflowMgtClient {
	return &WorkflowMgtClient{
		workflowMgtUrl: workflowMgtUrl,
		client:         api.NewIPHTTPClient(tokenStore),
	}
}

func (wfc *WorkflowMgtClient) CheckWorkflowStatus(
	orgId, buildId, envId string) (status *PromotionRequestStatus, err error) {

	req, err := http.NewRequest("GET",
		fmt.Sprintf(
			"%s/workflow-instances/status?wkfDefinitionId=ENV_PROMOTION&resource=%s_%s",
			wfc.workflowMgtUrl,
			buildId,
			envId,
		),
		nil,
	)

	if err != nil {
		return
	}

	resp, err := wfc.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while fetching cloud data planes: %w", err)
	}

	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return
}

func (wfc *WorkflowMgtClient) RequestWorkflowApproval(orgId string, reqBody RequestWorkflowApproval) (err error) {
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/workflow-instances", wfc.workflowMgtUrl), bytes.NewBuffer(reqBodyBytes))

	if err != nil {
		return
	}

	resp, err := wfc.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while requesting workflow approval: %w", err)
	}

	defer resp.Body.Close()
	return
}

func (wfc *WorkflowMgtClient) CancelApprovalRequest(orgId, wkfInstanceId string) (err error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/workflow-instances/%s/cancellation", wfc.workflowMgtUrl, wkfInstanceId), nil)

	if err != nil {
		return
	}

	resp, err := wfc.client.Do(req, orgId)
	if err != nil {
		return fmt.Errorf("error while cancelling approval request: %w", err)
	}

	defer resp.Body.Close()
	return
}

func (wfc *WorkflowMgtClient) IsWorkflowEnabled(orgId string, workflowId WorkflowId) (enabled bool, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/workflow/configs", wfc.workflowMgtUrl), nil)
	if err != nil {
		return false, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := wfc.client.Do(req, orgId)
	if err != nil {
		return false, fmt.Errorf("error while executing request: %w", err)
	}

	defer resp.Body.Close()

	var response []WorkflowConfig
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("error while decoding response: %w", err)
	}

	for _, config := range response {
		if config.WorkflowDefinitionId == string(workflowId) {
			return config.Enabled, nil
		}
	}
	return false, nil
}

func (wfc *WorkflowMgtClient) GetEnvPromotionWfStatus(orgId, imageId, envToId string) (status *WorkflowInstanceStatus, err error) {
	resource := getEnvPromotionWfResourceId(imageId, envToId)

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/workflow-instances/status?wkfDefinitionId=%s&resource=%s", wfc.workflowMgtUrl, ENV_PROMOTION_WORKFLOW_ID, resource),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := wfc.client.Do(req, orgId)
	if err != nil {
		return nil, fmt.Errorf("error while executing request: %w", err)
	}

	defer resp.Body.Close()

	var response WorkflowInstanceStatus
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error while decoding response: %w", err)
	}

	return &response, nil
}

func (wfc *WorkflowMgtClient) CreateEnvPromotionRequest(
	orgId, orgHandler, projectId, projectName, requestComment, componentName, buildId, imageId, envToId, envToName, envFromId, envFromName string,
) (wfInstanceId string, err error) {
	resource := getEnvPromotionWfResourceId(imageId, envToId)

	reqBody := RequestWorkflowApproval{
		ProjectId:      projectId,
		RequestComment: requestComment,
		Context: RequestWorkflowApprovalContext{
			WorkflowDefinitionIdentifier: "ENV_PROMOTION",
			Resource:                     resource,
		},
		Data: RequestWorkflowApprovalData{
			Payload: RequestWorkflowApprovalDataPayload{
				OrgName:       orgHandler,
				ProjectName:   projectName,
				ComponentName: componentName,
				BuildId:       buildId,
				EnvFrom: RequestWorkflowApprovalEnv{
					Id:   envFromId,
					Name: envFromName,
				},
				EnvTo: RequestWorkflowApprovalEnv{
					Id:   envToId,
					Name: envToName,
				},
			},
			Metadata: make(map[string]string),
		},
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error while encoding request body: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/workflow-instances", wfc.workflowMgtUrl), bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("error while creating request: %w", err)
	}

	resp, err := wfc.client.Do(req, orgId)
	if err != nil {
		return "", fmt.Errorf("error while executing request: %w", err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error while reading response body: %w", err)
	}

	wfInstanceId = buf.String()
	return wfInstanceId, nil
}

func getEnvPromotionWfResourceId(imageId string, envId string) string {
	return fmt.Sprintf("%s_%s", imageId, envId)
}
