package workflowmgt

type WorkflowId string

const (
	WFS_ENABLED string = "ENABLED"

	// user needs to wait for the approval
	WFS_PENDING string = "PENDING"

	// user need to submit new one
	WFS_NOT_FOUND string = "NOT_FOUND"
	WFS_REJECTED  string = "REJECTED"
	WFS_TIMEOUT   string = "TIMEOUT"
	WFS_CANCELLED string = "CANCELLED"

	// good to proceed
	WFS_DISABLED string = "DISABLED"
	WFS_APPROVED string = "APPROVED"

	ENV_PROMOTION_WORKFLOW_ID     WorkflowId = "ENV_PROMOTION"
	URL_CUSTOMIZATION_WORKFLOW_ID WorkflowId = "URL_CUSTOMIZATION"
)

type PromotionRequestStatus struct {
	Status     string `json:"status"`
	InstanceId string `json:"wkfInstanceId"`
}
