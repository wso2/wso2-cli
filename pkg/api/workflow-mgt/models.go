package workflowmgt

type RequestWorkflowApproval struct {
	ProjectId      string                         `json:"projectId"`
	RequestComment string                         `json:"requestComment"`
	Context        RequestWorkflowApprovalContext `json:"context"`
	Data           RequestWorkflowApprovalData    `json:"data"`
}

type RequestWorkflowApprovalContext struct {
	WorkflowDefinitionIdentifier string `json:"workflowDefinitionIdentifier"`
	Resource                     string `json:"resource"`
}

type RequestUrlCustomizationMetadata struct {
	ProjectName     string `json:"projectName"`
	ComponentName   string `json:"componentName"`
	EnvironmentName string `json:"environmentName"`
	CustomUrl       string `json:"customUrl"`
	Type            string `json:"type"`
	ProjectId       string `json:"projectId"`
}

type RequestWorkflowApprovalEnv struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
type RequestWorkflowApprovalData struct {
	Payload  RequestWorkflowApprovalDataPayload `json:"payload"`
	Metadata map[string]string                  `json:"metadata"`
}

type RequestWorkflowApprovalDataPayload struct {
	OrgName       string                     `json:"orgName"`
	ProjectName   string                     `json:"projectName"`
	ComponentName string                     `json:"componentName"`
	BuildId       string                     `json:"buildId"`
	EnvFrom       RequestWorkflowApprovalEnv `json:"envFrom"`
	EnvTo         RequestWorkflowApprovalEnv `json:"envTo"`
}
type WorkflowConfig struct {
	Id                   string   `json:"id"`
	OrgId                string   `json:"orgId"`
	WorkflowDefinitionId string   `json:"workflowDefinitionId"`
	AssigneeRoles        []string `json:"assigneeRoles"`
	Assignees            []string `json:"assignees"`
	NotifyEmails         []string `json:"notifyEmails"`
	Enabled              bool     `json:"enabled"`
	FormatRequestData    bool     `json:"formatRequestData"`
}

type WorkflowInstanceStatus struct {
	WorkflowInstanceId string `json:"wkfInstanceId"`
	Status             string `json:"status"`
}
