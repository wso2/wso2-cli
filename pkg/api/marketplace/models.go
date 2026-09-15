package marketplace

type MarketplaceService struct {
	Component         *MarketplaceServiceComponent `json:"component,omitempty"`
	EndpointRefs      map[string]string            `json:"endpointRefs,omitempty"`
	ServiceID         string                       `json:"serviceId"`
	Status            string                       `json:"status"`
	ServiceType       string                       `json:"serviceType"`
	IsThirdParty      bool                         `json:"isThirdParty"`
	ConnectionSchemas []MarketplaceServiceScheme   `json:"connectionSchemas"`
	ResourceID        string                       `json:"resourceId"`
	ThumbnailURL      string                       `json:"thumbnailUrl"`
	AverageRating     float64                      `json:"averageRating"`
	TotalRatingCount  int                          `json:"totalRatingCount"`
	CreatedTime       string                       `json:"createdTime"`
	Name              string                       `json:"name"`
	Version           string                       `json:"version"`
	ResourceType      string                       `json:"resourceType"`
	OrganizationID    string                       `json:"organizationId"`
	ProjectID         string                       `json:"projectId"`
	Summary           string                       `json:"summary"`
	Description       string                       `json:"description"`
	Categories        []string                     `json:"categories"`
	Visibility        []string                     `json:"visibility"`
	Tags              []string                     `json:"tags"`
	TemplateType      string                       `json:"templateType,omitempty"`
}

type MarketplaceServiceComponent struct {
	ComponentID  string `json:"componentId"`
	EndpointID   string `json:"endpointId"`
	ApimID       string `json:"apimId"`
	EndpointName string `json:"endpointName,omitempty"`
}

type MarketplaceGetServicesReq struct {
	// default is 20
	Limit int64 `json:"limit,omitempty" default:"20" format:"int64"`
	// Offset of the results. By default 0.
	Offset int64 `json:"offset,omitempty" default:"0" format:"int64"`
	// Sort by `name`, `createdTime`. By default sorted by `name`.
	SortBy string `json:"sortBy,omitempty"`
	//  Whether to sort in ascending order. By default `true`.
	SortAscending bool `json:"sortAscending,omitempty" default:"true"`
	// Search within the content (description, summary and IDL) of the service. By default `false`.
	SearchContent bool `json:"searchContent,omitempty" default:"false"`
	// Filter services based on network visibility. Possible values are "project", "org", "public".
	NetworkVisibilityFilter string `json:"networkVisibilityFilter,omitempty" default:"org"`
	// Optionally filter services based on service name, description, summary and IDL.
	Query *string `json:"query,omitempty"`
	// Optionally filter services based on tags. Multiple tags can be provided as a comma separated list.
	Tags *string `json:"tags,omitempty"`
	// Optionally filter services based on categories. Multiple categories can be provided as a comma separated list.
	Categories *string `json:"categories,omitempty"`
	// Optionally filter services based on whether they are third party or not. By default null, meaning this filter is not effective.
	IsThirdParty *bool `json:"isThirdParty,omitempty"`
	// When networkVisibilityFilter is "project", this parameter can be used to filter services
	NetworkVisibilityprojectId *string `json:"networkVisibilityprojectId,omitempty"`
}

type MarketplaceServiceScheme struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	IsDefault   bool                     `json:"isDefault"`
	Entries     []MarketplaceSchemeEntry `json:"entries"`
}

type MarketplaceSchemeEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	IsSensitive bool   `json:"isSensitive"`
	IsOptional  bool   `json:"isOptional"`
}

type MarketplaceResp struct {
	Count      int `json:"count"`
	Pagination struct {
		Limit  int `json:"limit"`
		Total  int `json:"total"`
		Offset int `json:"offset"`
	} `json:"pagination"`
	Data []MarketplaceService `json:"data"`
}

type MarketplaceIdlResp struct {
	EnvironmentId string `json:"environmentId"`
	Content       string `json:"content"`
	IdlType       string `json:"idlType"`
}

type MarketplaceDatabasesResp struct {
	Count      int                           `json:"count"`
	Pagination MarketplacePagination         `json:"pagination"`
	Data       []MarketplaceDatabaseResource `json:"data"`
}

type MarketplacePagination struct {
	Limit  int `json:"limit"`
	Total  int `json:"total"`
	Offset int `json:"offset"`
}

type MarketplaceDatabaseResource struct {
	ResourceDetails   MarketplaceDatabaseDetails    `json:"resourceDetails"`
	ConnectionSchemas []MarketplaceConnectionSchema `json:"connectionSchemas"`
	ThumbnailUrl      string                        `json:"thumbnailUrl"`
	AverageRating     float64                       `json:"averageRating"`
	TotalRatingCount  int                           `json:"totalRatingCount"`
	CreatedTime       string                        `json:"createdTime"`
	Name              string                        `json:"name"`
	Version           string                        `json:"version"`
	ResourceType      string                        `json:"resourceType"`
	OrganizationId    string                        `json:"organizationId"`
	Summary           string                        `json:"summary"`
	Description       string                        `json:"description"`
	Tags              []string                      `json:"tags"`
	Categories        []string                      `json:"categories"`
	Visibility        []string                      `json:"visibility"`
	Properties        map[string]string             `json:"properties"`
	ResourceId        string                        `json:"resourceId"`
}

type MarketplaceDatabaseDetails struct {
	DatabaseServerId   string `json:"databaseServerId"`
	DatabaseServerName string `json:"databaseServerName"`
	DatabaseType       string `json:"databaseType"`
	Status             string `json:"status"`
	IsRestricted       bool   `json:"isRestricted"`
	CloudProvider      string `json:"cloudProvider"`
	CloudRegion        string `json:"cloudRegion"`
}

type MarketplaceConnectionSchema struct {
	Name        string                       `json:"name"`
	Id          string                       `json:"id"`
	Description string                       `json:"description"`
	IsDefault   bool                         `json:"isDefault"`
	Entries     []MarketplaceConnectionEntry `json:"entries"`
}

type MarketplaceConnectionEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	IsSensitive bool   `json:"isSensitive"`
	IsOptional  bool   `json:"isOptional"`
}

type ListDatabasesResponse struct {
	Count      int             `json:"count"`
	Pagination Pagination      `json:"pagination"`
	Data       []DatabaseEntry `json:"data"`
}

type Pagination struct {
	Limit  int `json:"limit"`
	Total  int `json:"total"`
	Offset int `json:"offset"`
}

type DatabaseEntry struct {
	ResourceDetails   ResourceDetails    `json:"resourceDetails"`
	ConnectionSchemas []ConnectionSchema `json:"connectionSchemas"`
	ThumbnailUrl      string             `json:"thumbnailUrl"`
	AverageRating     float64            `json:"averageRating"`
	TotalRatingCount  int                `json:"totalRatingCount"`
	CreatedTime       string             `json:"createdTime"`
	Name              string             `json:"name"`
	Version           string             `json:"version"`
	ResourceType      string             `json:"resourceType"`
	OrganizationId    string             `json:"organizationId"`
	Summary           string             `json:"summary"`
	Description       string             `json:"description"`
	Tags              []string           `json:"tags"`
	Categories        []string           `json:"categories"`
	Visibility        []string           `json:"visibility"`
	Properties        map[string]string  `json:"properties"`
	ResourceId        string             `json:"resourceId"`
}

type ResourceDetails struct {
	DatabaseServerId   string `json:"databaseServerId"`
	DatabaseServerName string `json:"databaseServerName"`
	DatabaseType       string `json:"databaseType"`
	Status             string `json:"status"`
	IsRestricted       bool   `json:"isRestricted"`
	CloudProvider      string `json:"cloudProvider"`
	CloudRegion        string `json:"cloudRegion"`
}

type ConnectionSchema struct {
	Name        string                  `json:"name"`
	ID          string                  `json:"id"`
	Description string                  `json:"description"`
	IsDefault   bool                    `json:"isDefault"`
	Entries     []ConnectionSchemaEntry `json:"entries"`
}

type ConnectionSchemaEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	IsSensitive bool   `json:"isSensitive"`
	IsOptional  bool   `json:"isOptional"`
}

// ConnectionGuideReq represents the query parameters for fetching connection guides
type ConnectionGuideReq struct {
	ConfigGroupId          string `json:"configGroupId,omitempty"`
	Audience               string `json:"audience,omitempty"`
	OrgId                  string `json:"orgId,omitempty"`
	ComponentType          string `json:"componentType,omitempty"`
	IsProjectLvlConnection bool   `json:"isProjectLvlConnection,omitempty"`
	ConfigFileType         string `json:"configFileType,omitempty"`
	ConnectionName         string `json:"connectionName,omitempty"`
	BuildpackType          string `json:"buildpackType,omitempty"`
}
