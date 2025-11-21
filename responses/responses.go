package responses

import "encoding/json"

type PaginatedResponse struct {
	PaginatedData interface{} `json:"items"`
	TotalItems    int64       `json:"total_items"`
	TotalPages    int64       `json:"total_pages"`
	CurrentPage   int64       `json:"current_page"`
}

type KeyClockTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	Scope            string `json:"scope"`
}

type KeycloakUser struct {
	ID                   string              `json:"id"`
	CreatedTimestamp     int64               `json:"createdTimestamp"`
	Username             string              `json:"username"`
	Enabled              bool                `json:"enabled"`
	TOTP                 bool                `json:"totp"`
	EmailVerified        bool                `json:"emailVerified"`
	FirstName            string              `json:"firstName"`
	Email                string              `json:"email"`
	Attributes           map[string][]string `json:"attributes"` // e.g., "company_code": ["KT001"]
	DisableableCredTypes []string            `json:"disableableCredentialTypes"`
	RequiredActions      []string            `json:"requiredActions"`
	NotBefore            int                 `json:"notBefore"`
	Access               KeycloakAccess      `json:"access"`
}

type KeycloakAccess struct {
	ManageGroupMembership bool `json:"manageGroupMembership"`
	View                  bool `json:"view"`
	MapRoles              bool `json:"mapRoles"`
	Impersonate           bool `json:"impersonate"`
	Manage                bool `json:"manage"`
}

func NewKeyClockTokenResponse() *KeyClockTokenResponse {
	return &KeyClockTokenResponse{}
}

func NewKeycloakUser() *KeycloakUser {
	return &KeycloakUser{}
}

func (r *KeyClockTokenResponse) Parse(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r *KeyClockTokenResponse) GetAccessToken() string {
	return r.AccessToken
}

func (r *KeycloakUser) GetCompanyCode() string {
	if val, ok := r.Attributes["company_code"]; ok {
		return val[0]
	}
	return ""
}

func (r *KeycloakUser) GetContactPerson() string {
	if val, ok := r.Attributes["contact_person"]; ok {
		return val[0]
	}
	return ""
}

func (r *KeycloakUser) GetContactPersonEmail() string {
	if val, ok := r.Attributes["contact_person_email"]; ok {
		return val[0]
	}
	return ""
}

func (r *KeycloakUser) GetContactPersonMobile() string {
	if val, ok := r.Attributes["contact_person_mobile"]; ok {
		return val[0]
	}
	return ""
}

func (r *KeycloakUser) GetOProductCodes() []string {
	if val, ok := r.Attributes["product_code"]; ok {
		return val
	}
	return []string{}
}
