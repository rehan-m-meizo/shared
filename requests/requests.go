package requests

import (
	"shared/pkgs/validations"

	"github.com/gin-gonic/gin"
)

type DocumentParsedRequest struct {
	EntityId    interface{} `json:"entity_id"`
	CompanyCode interface{} `json:"company_code"`
	Type        interface{} `json:"type"`
	ModuleData  interface{} `json:"module_data"`
	ModuleCount interface{} `json:"module_count"`
}

type KeycloakUserCreateRequest struct {
	Username      string                   `json:"username" validate:"required"`
	Email         string                   `json:"email" validate:"required,email"`
	Enabled       bool                     `json:"enabled"`
	EmailVerified bool                     `json:"emailVerified"`
	FirstName     string                   `json:"firstName"`
	Credentials   []KeycloakUserCredential `json:"credentials" validate:"required,min=1"`
	Attributes    map[string][]string      `json:"attributes"`
}

type KeycloakUserCredential struct {
	Type      string `json:"type" validate:"required,oneof=password"`
	Value     string `json:"value" validate:"required"`
	Temporary bool   `json:"temporary"`
}

type SendEmailRequest struct {
	To        string `json:"to" validate:"required,email"`
	Subject   string `json:"subject" validate:"required,max=100"`
	HtmlBody  string `json:"body" validate:"required"`
	FileBytes []byte `json:"file_bytes,omitempty"`
}

type LicenseProvisionRequest struct {
	GroupCode           string `json:"group_code"`
	EmployeeCode        string `json:"employee_code"`
	CompanyCode         string `json:"company_code"`
	LicenseCode         string `json:"license_code"`
	LicenseKey          string `json:"license_key"`
	ProductCode         string `json:"product_code"`
	Password            string `json:"password"`
	CompanyName         string `json:"company_name"`
	ProductName         string `json:"product_name"`
	ContactPerson       string `json:"contact_person"`
	ContactPersonEmail  string `json:"contact_person_email"`
	ContactPersonMobile string `json:"contact_person_mobile"`
}

type KeyClockCreateUserRequest struct {
	Email        string `json:"email"`
	Username     string `json:"username"`
	FullName     string `json:"fullName"`
	CompanyCode  string `json:"companyCode"`
	MobileNumber string `json:"mobileNumber"`
	IsActive     bool   `json:"isActive"`
}

func NewKeycloakUserCreateRequest() *KeycloakUserCreateRequest {
	return &KeycloakUserCreateRequest{}
}

func NewLicenseProvisionRequest() *LicenseProvisionRequest {
	return &LicenseProvisionRequest{}
}

func NewSendEmailRequest() *SendEmailRequest {
	return &SendEmailRequest{}
}

func NewKeyClockCreateUserRequest() *KeyClockCreateUserRequest {
	return &KeyClockCreateUserRequest{}
}

func NewDocumentParsedRequest() *DocumentParsedRequest {
	return &DocumentParsedRequest{}
}

func (r *LicenseProvisionRequest) Validate(c *gin.Context) error {
	if err := validations.ValidatePayload(c, r); err != nil {
		return err
	}
	return nil
}

func (r *KeycloakUserCreateRequest) Bind(l *LicenseProvisionRequest) error {
	r.Username = l.EmployeeCode
	r.Email = l.ContactPersonEmail
	r.FirstName = l.ContactPerson
	r.Enabled = true
	r.EmailVerified = true
	r.Credentials = []KeycloakUserCredential{
		{
			Type:      "password",
			Value:     l.Password,
			Temporary: false,
		},
	}
	r.Attributes = map[string][]string{
		"group_code":           {l.GroupCode},
		"product_code":         {l.ProductCode},
		"contact_person_email": {l.ContactPersonEmail},
		"role":                 {"admin"},
		"password_reset": {
			"true",
		},
	}
	return nil
}

func (r *KeycloakUserCreateRequest) Validate() error {

	if err := validations.ValidateStruct(r); err != nil {
		return err
	}

	return nil
}

func (r *SendEmailRequest) Validate(c *gin.Context) error {
	if err := validations.ValidatePayload(c, r); err != nil {
		return err
	}
	return nil
}

func (r *KeyClockCreateUserRequest) Validate(c *gin.Context) error {
	if err := validations.ValidatePayload(c, r); err != nil {
		return err
	}
	return nil
}
