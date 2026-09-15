package platformservices

type CreateDatabaseServerRequest struct {
	Name            string `json:"name"`
	CloudProvider   string `json:"cloud_provider"`
	Region          string `json:"cloud_region"`
	IsVectorEnabled bool   `json:"is_vector_enabled"`
	ServicePlanId   string `json:"service_plan_id"`
}

type DatabaseServer struct {
	ID                   string           `json:"id"`
	Type                 string           `json:"type"`
	Name                 string           `json:"name"`
	CloudProvider        string           `json:"cloud_provider"`
	Region               string           `json:"cloud_region"`
	DisplayOnMarketplace bool             `json:"display_on_marketplace"`
	IsVectorEnabled      bool             `json:"is_vector_enabled"`
	ServicePlanId        string           `json:"service_plan_id"`
	CreatedAt            string           `json:"created_at"`
	Status               string           `json:"status"`
	AllowedIPs           AllowedIPs       `json:"allowed_ips,omitempty"`
	ConnectionParams     ConnectionParams `json:"connection_params,omitempty"`
	Maintenance          Maintenance      `json:"maintenance,omitempty"`
	Nodes                []Node           `json:"nodes,omitempty"`
	ServiceVersion       string           `json:"service_version,omitempty"`
}

type CreateDatabaseRequest struct {
	Name string `json:"name"`
}

type Database struct {
	Name                 string `json:"name"`
	DisplayOnMarketplace bool   `json:"display_on_marketplace"`
	Status               string `json:"status"`
}

type CreateDBCredRequest struct {
	ApplicableEnvironments []string `json:"applicable_environments"`
	Database               string   `json:"database"`
	DisplayName            string   `json:"display_name"`
	IsSuperAdmin           bool     `json:"is_super_admin"`
}

type PublishDatabaseRequest struct {
	DisplayOnMarketplace bool   `json:"display_on_marketplace"`
	Name                 string `json:"name"`
	Status               string `json:"status"`
}

type PublishDatabaseResponse struct {
	DisplayOnMarketplace bool   `json:"display_on_marketplace"`
	Name                 string `json:"name"`
	Status               string `json:"status"`
}

type AllowedIPs struct {
	AllowList []AllowItem `json:"allow_list,omitempty"`
	Mode      string      `json:"mode"`
}

type AllowItem struct {
	Description string `json:"description"`
	Network     string `json:"network"`
}

type ConnectionParams struct {
	Database      string `json:"database"`
	Host          string `json:"host"`
	PasswordReset bool   `json:"password_reset"`
	Port          string `json:"port"`
	SSLRequired   bool   `json:"ssl_required"`
	User          string `json:"user"`
}

type Maintenance struct {
	Day  string `json:"day"`
	Time string `json:"time"`
}

type Node struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	State string `json:"state"`
}

type ServicePlan struct {
	BackupIntervalHours int      `json:"backup_interval_hours"`
	BackupRetentionDays int      `json:"backup_retention_days"`
	Description         string   `json:"description"`
	FreeTrialAvailable  bool     `json:"free_trial_available"`
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	NodeCount           int      `json:"node_count"`
	Type                string   `json:"type"`
	Regions             []Region `json:"regions"`
}

type Region struct {
	CloudProvider   string `json:"cloud_provider"`
	Region          string `json:"region"`
	HourlyPriceUSD  string `json:"hourly_price_usd"`
	MonthlyPriceUSD string `json:"monthly_price_usd"`
	NodeCPUCount    int    `json:"node_cpu_count"`
	NodeRamGB       int    `json:"node_ram_gb"`
	StorageGB       int    `json:"storage_gb"`
}

type DatabaseCredential struct {
	ApplicableEnvironments []string  `json:"applicable_environments"`
	CreatedAt              string    `json:"created_at"`
	DatabaseName           string    `json:"database_name"`
	DisplayName            string    `json:"display_name"`
	ID                     string    `json:"id"`
	IsSuperAdmin           bool      `json:"is_super_admin"`
	PrivilegeLevels        *[]string `json:"privilege_levels"` // nullable
	UpdatedAt              string    `json:"updated_at"`
}
