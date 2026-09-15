package platformgit

type GithubRepository struct {
	Name string `json:"name"`
}

type GithubOrganization struct {
	OrgName      string             `json:"orgName"`
	OrgHandler   string             `json:"orgHandler"`
	Repositories []GithubRepository `json:"repositories"`
}

type CommitHistory struct {
	Message  string `json:"message"`
	Sha      string `json:"sha"`
	IsLatest bool   `json:"isLatest"`
	Author   Author `json:"author"`
}

type IsPublicRepoResponse struct {
	IsPublicRepo bool `json:"isPublicRepo"`
}

type Author struct {
	Name      string `json:"name"`
	Date      string `json:"date"`
	Email     string `json:"email"`
	AvatarUrl string `json:"avatarUrl"`
}
