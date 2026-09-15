package subscription

type GetSubscriptionsResp struct {
	Component int                `json:"count"`
	List      []SubscriptionItem `json:"list"`
	CloudType string             `json:"cloudType"`
	EmailType string             `json:"emailType"`
}

type SubscriptionItem struct {
	SubscriptionId                    string  `json:"subscriptionId"`
	TierId                            *string `json:"tierId,omitempty"`
	SupportPlanId                     *string `json:"supportPlanId,omitempty"`
	CloudType                         *string `json:"cloudType,omitempty"`
	SubscriptionType                  *string `json:"subscriptionType,omitempty"`
	SubscriptionBillingProvider       *string `json:"subscriptionBillingProvider,omitempty"`
	SubscriptionBillingProviderStatus *string `json:"subscriptionBillingProviderStatus,omitempty"`
}
