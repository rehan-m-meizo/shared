package validations

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

// initValidator initializes the validator singleton.
func initValidator() {
	once.Do(func() {
		validate = validator.New()
		_ = validate.RegisterValidation("password", validatePassword)
		_ = validate.RegisterValidation("comma-separated-valid", validateCommaSeparated)
		_ = validate.RegisterValidation("phone", validatePhone)
		_ = validate.RegisterValidation("ascii", validateASCII)
		_ = validate.RegisterValidation("uuid", validateUUID)
		_ = validate.RegisterValidation("ip", validateIP)
		_ = validate.RegisterValidation("domain", validateDomain)
		_ = validate.RegisterValidation("hexcolor", validateHexColor)
		_ = validate.RegisterValidation("base64", validateBase64)
		_ = validate.RegisterValidation("json", validateJSON)
		_ = validate.RegisterValidation("creditcard", validateCreditCard)
		_ = validate.RegisterValidation("latitude", validateLatitude)
		_ = validate.RegisterValidation("longitude", validateLongitude)
		_ = validate.RegisterValidation("username", validateUsername)
		_ = validate.RegisterValidation("slug", validateSlug)
		_ = validate.RegisterValidation("date", validateDate)
		_ = validate.RegisterValidation("time", validateTime)
		_ = validate.RegisterValidation("datetime", validateDateTime)
		_ = validate.RegisterValidation("zipcode", validateZipcode)
		_ = validate.RegisterValidation("macaddress", validateMacAddress)
		_ = validate.RegisterValidation("email", validateEmail)
		_ = validate.RegisterValidation("current_date", validateCurrentDate)
		_ = validate.RegisterValidation("telephone", validateTel)
		_ = validate.RegisterValidation("pan", validatePAN)
		_ = validate.RegisterValidation("tan", validateTAN)
		_ = validate.RegisterValidation("aadhar", validateAadhar)
	})
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 6 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, c := range password {
		if unicode.IsUpper(c) {
			hasUpper = true
		} else if unicode.IsLower(c) {
			hasLower = true
		} else if unicode.IsDigit(c) {
			hasDigit = true
		} else if unicode.IsPunct(c) || unicode.IsSymbol(c) {
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// validateCommaSeparated checks if a string is a valid comma-separated list of words.
func validateCommaSeparated(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^([a-zA-Z0-9]+,)*[a-zA-Z0-9]+$`).MatchString(fl.Field().String())
}

// validatePhone checks if a string is a valid phone number.
func validatePhone(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^\+?[1-9]\d{1,14}$`).MatchString(fl.Field().String())
}

// validateASCII checks if a string contains only ASCII characters.
func validateASCII(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^[\x00-\x7F]+$`).MatchString(fl.Field().String())
}

// validateUUID checks if a string is a valid UUID.
func validateUUID(fl validator.FieldLevel) bool {
	return regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`).MatchString(fl.Field().String())
}

func validateIP(fl validator.FieldLevel) bool {
	ipRegex := `^(\d{1,3}\.){3}\d{1,3}$|^([a-fA-F0-9:]+:+)+[a-fA-F0-9]+$`
	return regexp.MustCompile(ipRegex).MatchString(fl.Field().String())
}

func validateDomain(fl validator.FieldLevel) bool {
	domainRegex := `^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,6}$`
	return regexp.MustCompile(domainRegex).MatchString(fl.Field().String())
}

func validateHexColor(fl validator.FieldLevel) bool {
	hexRegex := `^#(?:[0-9a-fA-F]{3}){1,2}$`
	return regexp.MustCompile(hexRegex).MatchString(fl.Field().String())
}

func validateBase64(fl validator.FieldLevel) bool {
	base64Regex := `^(?:[A-Za-z0-9+\/]{4})*(?:[A-Za-z0-9+\/]{2}==|[A-Za-z0-9+\/]{3}=)?$`
	return regexp.MustCompile(base64Regex).MatchString(fl.Field().String())
}

func validateJSON(fl validator.FieldLevel) bool {
	jsonRegex := `^\{.*\}$`
	return regexp.MustCompile(jsonRegex).MatchString(fl.Field().String())
}

func validateCreditCard(fl validator.FieldLevel) bool {
	ccRegex := `^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|6(?:011|5[0-9]{2})[0-9]{12}|(?:2131|1800|35\d{3})\d{11})$`
	return regexp.MustCompile(ccRegex).MatchString(fl.Field().String())
}

func validateLatitude(fl validator.FieldLevel) bool {
	latRegex := `^-?(90(\.0+)?|[1-8]?\d(\.\d+)?)$`
	return regexp.MustCompile(latRegex).MatchString(fl.Field().String())
}

func validateTel(fl validator.FieldLevel) bool {
	telRegex := `^\+?[0-9\-]{7,15}$`
	return regexp.MustCompile(telRegex).MatchString(fl.Field().String())
}

func validateLongitude(fl validator.FieldLevel) bool {
	lonRegex := `^-?(180(\.0+)?|((1[0-7]\d)|([1-9]?\d))(\.\d+)?)$`
	return regexp.MustCompile(lonRegex).MatchString(fl.Field().String())
}

func validateUsername(fl validator.FieldLevel) bool {
	usernameRegex := `^[a-zA-Z0-9_]{3,20}$`
	return regexp.MustCompile(usernameRegex).MatchString(fl.Field().String())
}

func validateSlug(fl validator.FieldLevel) bool {
	slugRegex := `^[a-z0-9]+(?:-[a-z0-9]+)*$`
	return regexp.MustCompile(slugRegex).MatchString(fl.Field().String())
}

func validateTime(fl validator.FieldLevel) bool {
	timeRegex := `^(?:[01]\d|2[0-3]):[0-5]\d:[0-5]\d$`
	return regexp.MustCompile(timeRegex).MatchString(fl.Field().String())
}

func validateDateTime(fl validator.FieldLevel) bool {
	dateTimeRegex := `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`
	return regexp.MustCompile(dateTimeRegex).MatchString(fl.Field().String())
}

func validateZipcode(fl validator.FieldLevel) bool {
	zipcodeRegex := `^\d{5}(-\d{4})?$`
	return regexp.MustCompile(zipcodeRegex).MatchString(fl.Field().String())
}

func validateMacAddress(fl validator.FieldLevel) bool {
	macRegex := `^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`
	return regexp.MustCompile(macRegex).MatchString(fl.Field().String())
}

func validateCurrentDate(fl validator.FieldLevel) bool {
	currentDate := time.Now().Format("2006-01-02")
	return fl.Field().String() == currentDate
}

func validateDate(fl validator.FieldLevel) bool {
	_, err := time.Parse("2006-01-02", fl.Field().String())

	if err != nil {
		return false
	}

	return true
}

func validateEmail(fl validator.FieldLevel) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(emailRegex).MatchString(fl.Field().String())
}

// validatePAN validates Indian PAN (Permanent Account Number) format
// Format: AAAAA9999A (5 uppercase letters, 4 digits, 1 uppercase letter)
func validatePAN(fl validator.FieldLevel) bool {
	panRegex := `^[A-Z]{5}[0-9]{4}[A-Z]{1}$`
	return regexp.MustCompile(panRegex).MatchString(fl.Field().String())
}

// validateTAN validates Indian TAN (Tax Deduction and Collection Account Number) format
// Format: AAAA99999A (4 uppercase letters, 5 digits, 1 uppercase letter)
func validateTAN(fl validator.FieldLevel) bool {
	tanRegex := `^[A-Z]{4}[0-9]{5}[A-Z]{1}$`
	return regexp.MustCompile(tanRegex).MatchString(fl.Field().String())
}

// validateAadhar validates Indian Aadhar number format
// Format: 12 digits
func validateAadhar(fl validator.FieldLevel) bool {
	aadharRegex := `^[0-9]{12}$`
	return regexp.MustCompile(aadharRegex).MatchString(fl.Field().String())
}

// RegisterValidation allows registering custom validation rules.
func RegisterValidation(tag string, fn validator.Func) error {
	initValidator()
	return validate.RegisterValidation(tag, fn)
}

// handleValidationErrors processes validation errors into a readable format.
func handleValidationErrors(err error) error {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		var errMsg string
		for _, err := range validationErrors {
			switch err.Tag() {
			case "required":
				errMsg += fmt.Sprintf("%s is required. ", err.Field())
			case "max":
				errMsg += fmt.Sprintf("%s must be less than or equal to %s. ", err.Field(), err.Param())
			case "min":
				errMsg += fmt.Sprintf("%s must be greater than or equal to %s. ", err.Field(), err.Param())
			case "gte":
				errMsg += fmt.Sprintf("%s must be greater than %s. ", err.Field(), err.Param())
			case "lte":
				errMsg += fmt.Sprintf("%s must be less than %s. ", err.Field(), err.Param())
			case "len":
				errMsg += fmt.Sprintf("%s must be %s characters long. ", err.Field(), err.Param())
			case "email":
				errMsg += fmt.Sprintf("%s must be a valid email address. ", err.Field())
			case "numeric":
				errMsg += fmt.Sprintf("%s must be a number. ", err.Field())
			case "alpha":
				errMsg += fmt.Sprintf("%s must contain only letters. ", err.Field())
			case "alphanum":
				errMsg += fmt.Sprintf("%s must contain only letters and numbers. ", err.Field())
			case "alphaunicode":
				errMsg += fmt.Sprintf("%s must be a valid Unicode string. ", err.Field())
			case "comma-separated-valid":
				errMsg += fmt.Sprintf("%s contains invalid values. ", err.Field())
			case "phone":
				errMsg += fmt.Sprintf("%s must be a valid phone number. ", err.Field())
			case "ascii":
				errMsg += fmt.Sprintf("%s must contain only ASCII characters. ", err.Field())
			case "uuid":
				errMsg += fmt.Sprintf("%s must be a valid UUID. ", err.Field())
			case "ip":
				errMsg += fmt.Sprintf("%s must be a valid IP address. ", err.Field())
			case "domain":
				errMsg += fmt.Sprintf("%s must be a valid domain. ", err.Field())
			case "hexcolor":
				errMsg += fmt.Sprintf("%s must be a valid hex color. ", err.Field())
			case "base64":
				errMsg += fmt.Sprintf("%s must be a valid base64 string. ", err.Field())
			case "json":
				errMsg += fmt.Sprintf("%s must be a valid JSON string. ", err.Field())
			case "creditcard":
				errMsg += fmt.Sprintf("%s must be a valid credit card number. ", err.Field())
			case "latitude":
				errMsg += fmt.Sprintf("%s must be a valid latitude. ", err.Field())
			case "longitude":
				errMsg += fmt.Sprintf("%s must be a valid longitude. ", err.Field())
			case "username":
				errMsg += fmt.Sprintf("%s must be a valid username. ", err.Field())
			case "slug":
				errMsg += fmt.Sprintf("%s must be a valid slug. ", err.Field())
			case "date":
				errMsg += fmt.Sprintf("%s must be a valid date. ", err.Field())
			case "time":
				errMsg += fmt.Sprintf("%s must be a valid time. ", err.Field())
			case "datetime":
				errMsg += fmt.Sprintf("%s must be a valid datetime. ", err.Field())
			case "zipcode":
				errMsg += fmt.Sprintf("%s must be a valid zipcode. ", err.Field())
			case "macaddress":
				errMsg += fmt.Sprintf("%s must be a valid MAC address. ", err.Field())
			case "password":
				errMsg += fmt.Sprintf("%s must be a valid password. ", err.Field())
			case "current_date":
				errMsg += fmt.Sprintf("%s must be today's date. ", err.Field())
			case "valid_date":
				errMsg += fmt.Sprintf("%s must be a valid date. ", err.Field())
			case "latlong":
				errMsg += fmt.Sprintf("%s must be a valid lat long format.", err.Field())
			case "telephone":
				errMsg += fmt.Sprintf("%s must be a valid telephone number. ", err.Field())
			case "url":
				errMsg += fmt.Sprintf("%s must be a valid url. ", err.Field())
			case "pan":
				errMsg += fmt.Sprintf("%s must be a valid PAN number (format: AAAAA9999A). ", err.Field())
			case "tan":
				errMsg += fmt.Sprintf("%s must be a valid TAN number (format: AAAA99999A). ", err.Field())
			case "aadhar":
				errMsg += fmt.Sprintf("%s must be a valid Aadhar number (12 digits). ", err.Field())
			default:
				errMsg += fmt.Sprintf("%s failed '%s' validation. ", err.Field(), err.Tag())
			}
		}
		return errors.New(errMsg)
	}
	return nil
}

// validateStruct validates any struct.
func validateStruct(payload interface{}) error {
	initValidator()

	// Add panic recovery to catch "Bad field name provided" errors
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Errorf("validation panic: %v - this usually indicates a struct field name issue", r))
		}
	}()

	if err := validate.Struct(payload); err != nil {
		return handleValidationErrors(err)
	}
	return nil
}

func ValidatePayload(c *gin.Context, out interface{}) error {
	if out == nil {
		return errors.New("request body is empty")
	}

	if err := c.ShouldBindBodyWithJSON(out); err != nil {
		return err
	}

	if err := validateStruct(out); err != nil {
		return err
	}

	return nil
}

func ValidateHeaders(c *gin.Context, out interface{}) error {
	if err := c.ShouldBindHeader(out); err != nil {
		return err
	}

	if out == nil {
		return errors.New("requests body is empty")
	}

	if err := validateStruct(out); err != nil {
		return err
	}

	return nil
}

// Deprecated: use ValidateParam instead
func ValidateQuery(c *gin.Context, out interface{}) error {
	return ValidateParam(c, out)
}

func ValidateParam(c *gin.Context, out interface{}) error {
	if err := c.ShouldBindQuery(out); err != nil {
		return err
	}

	if out == nil {
		return errors.New("params are required")
	}

	if err := validateStruct(out); err != nil {
		return err
	}

	return nil
}

// Deprecated: use ValidateURI instead
func ValidateParams(c *gin.Context, out interface{}) error {
	return ValidateURI(c, out)
}

func ValidateURI(c *gin.Context, out interface{}) error {
	if err := c.ShouldBindUri(out); err != nil {
		return err
	}

	if out == nil {
		return errors.New("uris are empty")
	}

	if err := validateStruct(out); err != nil {
		return err
	}

	return nil
}

func ValidateStruct(out interface{}) error {
	if err := validateStruct(out); err != nil {
		return err
	}
	return nil
}
