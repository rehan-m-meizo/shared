package constants

import (
	"errors"
	"fmt"
	"shared/config"
)

var (
	AllowedTypes = []string{"application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"}
)

var (
	// Postgres
	PostgresUser     = config.Config("SERVICE_POSTGRES_USER")
	PostgresPassword = config.Config("SERVICE_POSTGRES_PASSWORD")
	PostgresHost     = config.Config("SERVICE_POSTGRES_HOST")
	PostgresPort     = config.Config("SERVICE_POSTGRES_PORT")
	PostgresSSLMode  = config.Config("SERVICE_POSTGRES_SSL_MODE")
	PostgresDatabase = config.Config("SERVICE_POSTGRES_DATABASE")
	EncryptionKey    = config.Config("ENCRYPTION_KEY")

	// Mongo
	MongoUser     = config.Config("SERVICE_MONGODB_USER")
	MongoPassword = config.Config("SERVICE_MONGODB_PASSWORD")
	MongoHost     = config.Config("SERVICE_MONGODB_HOST")
	MongoPort     = config.Config("SERVICE_MONGODB_PORT")

	LedgerDB         = config.Config("LEDGER_DB")
	LedgerCollection = config.Config("LEDGER")

	CompanyServiceDatabase      = config.Config("COMPANY_SERVICE_DATABASE")
	ProductServiceDatabase      = config.Config("PRODUCT_SERVICE_DATABASE")
	LicenseServiceDatabase      = config.Config("LICENSE_SERVICE_DATABASE")
	NotificationServiceDatabase = config.Config("NOTIFICATION_SERVICE_DATABASE")
	UserServiceDatabase         = config.Config("USER_SERVICE_DATABASE")
	FormServiceDatabase         = config.Config("FORM_SERVICE_DATABASE")

	// Redis
	RedisHost     = config.Config("SERVICE_REDIS_HOST")
	RedisPort     = config.Config("SERVICE_REDIS_PORT")
	RedisUsername = config.Config("REDIS_USERNAME")
	RedisPassword = config.Config("REDIS_PASSWORD")

	// KeyCloak

	KeyCloakBaseURL           = config.Config("KEYCLOAK_BASE_URL")
	KeyCloakRealm             = config.Config("KEYCLOAK_REALM")
	KeyCloakLoginUrl          = config.Config("KEYCLOAK_LOGIN_URL")
	KeyCloakLoginGrantType    = config.Config("KEYCLOAK_LOGIN_GRANT_TYPE")
	KeyCloakClientId          = config.Config("KEYCLOAK_CLIENT_ID")
	KeyCloakLoginScope        = config.Config("KEYCLOAK_LOGIN_SCOPE")
	KeyCloakRefreshGrantType  = config.Config("KEYCLOAK_REFRESH_GRANT_TYPE")
	KeyCloakClientSecret      = config.Config("KEYCLOAK_CLIENT_SECRET")
	KeyCloakClientSecretScope = config.Config("KEYCLOAK_CLIENT_SECRET_SCOPE")

	// Opa
	OpaBaseURL = config.Config("OPA_BASE_URL")

	// Rabbit MQ
	RabbitMqUrl      = config.Config("RABBIT_MQ_URL")
	RabbitMqExchange = config.Config("RABBIT_MQ_EXCHANGE")
	RabbitMqDLX      = config.Config("RABBIT_MQ_DLX")
	RabbitMqQueue    = config.Config("RABBIT_MQ_QUEUE")

	// Webhook
	CompanyURL        = config.Config("COMPANY_URL")
	LicenseURL        = config.Config("LICENSE_URL")
	AuthenticationURL = config.Config("AUTHENTICATION_URL")
	NotificationURL   = config.Config("NOTIFICATION_URL")

	// License
	LicenseIssued      = config.Config("ROUTING_LICENSE_ISSUED")
	LicenseRevoked     = config.Config("ROUTING_LICENSE_REVOKED")
	LicenseExtended    = config.Config("ROUTING_LICENSE_EXTENDED")
	LicenseReactivated = config.Config("ROUTING_LICENSE_REACTIVATED")

	// User
	UserCreated = config.Config("ROUTING_USER_CREATED")

	// Miscellanouse
	WelcomEmail = config.Config("ROUTING_WELCOME_EMAIL")

	// Form
	FormPublished = config.Config("ROUTING_FORM_PUBLISHED")

	// Submission
	FormSubmitted         = config.Config("ROUTING_FORM_SUBMITTED")
	FormSubmissionUpdated = config.Config("ROUTING_FORM_SUBMISSION_UPDATED")
	FormSubmissionDeleted = config.Config("ROUTING_FORM_SUBMISSION_DELETED")
	FormBulkSubmission    = config.Config("ROUTING_FORM_BULK_SUBMISSION")

	MasterImport = config.Config("ROUTING_MASTER_IMPORT")

	// Routing Keys
	RoutingKeys = []string{
		config.Config("ROUTING_LICENSE_ISSUED"),
		config.Config("ROUTING_LICENSE_REVOKED"),
		config.Config("ROUTING_LICENSE_EXTENDED"),
		config.Config("ROUTING_LICENSE_REACTIVATED"),

		config.Config("ROUTING_USER_CREATED"),

		// Miscellanouse
		config.Config("ROUTING_WELCOME_EMAIL"),

		config.Config("ROUTING_FORM_PUBLISHED"),
	}

	// SMTP
	SmtpHost  = config.Config("SMTP_HOST")
	SmtpPort  = config.Config("SMTP_PORT")
	SmtpUser  = config.Config("SMTP_USERNAME")
	SmtpPass  = config.Config("SMTP_PASSWORD")
	FromEmail = config.Config("FROM_EMAIL")

	// Kafka
	KafkaConsumerGroup = config.Config("KAFKA_CONSUMER_GROUP")
	KafkaBrokers       = config.Config("KAFKA_BROKERS")
	KafkaUser          = config.Config("KAFKA_USER")
	KafkaPass          = config.Config("KAFKA_PASS")

	// SeaweedFS
	SeaweedFSHost = config.Config("SEAWEEDFS_HOST")
	SeaweedFSPort = config.Config("SEAWEEDFS_PORT")

	// DATABASE
	FormCollection     = config.Config("FORM_COLLECTION")
	FormLiveCollection = config.Config("FORM_LIVE_COLLECTION")
)

var (
	ErrKeyCloaUserNotFound = errors.New("user not found on keycloak")
)

var (
	MeizoDomain = config.Config("DOMAIN")
)

const (
	ContentTypeJSON              = "application/json"
	ContentTypeFormURLEncoded    = "application/x-www-form-urlencoded"
	ContentTypeMultipartFormData = "multipart/form-data"
	ContentTypeXML               = "application/xml"
	ContentTypeTextPlain         = "text/plain"

	// File-specific
	ContentTypeOctetStream = "application/octet-stream"
	ContentTypePDF         = "application/pdf"
	ContentTypeZIP         = "application/zip"

	// Images
	ContentTypeJPEG = "image/jpeg"
	ContentTypePNG  = "image/png"
	ContentTypeGIF  = "image/gif"
	ContentTypeSVG  = "image/svg+xml"
	ContentTypeWEBP = "image/webp"

	// Audio
	ContentTypeMP3 = "audio/mpeg"
	ContentTypeWAV = "audio/wav"
	ContentTypeOGG = "audio/ogg"

	// Video
	ContentTypeMP4  = "video/mp4"
	ContentTypeWEBM = "video/webm"
	ContentTypeAVI  = "video/x-msvideo"
	KeyClaims       = "key-claims"
)

var (
	PgUrl = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		PostgresUser,
		PostgresPassword,
		PostgresHost,
		PostgresPort,
		PostgresDatabase,
		PostgresSSLMode)
)

const (
	TopicEmployeeCreated  = "employee.created"
	TopicLicenseIssued    = "license.issued"
	TopicDocumentImported = "document.imported"

	TopicAuthCreateUser         = "auth.create-user"
	TopicAuthCreateUserFailed   = "auth.create-user-failed"
	TopicAuthCreateAccess       = "auth.create-access"
	TopicAuthCreateAccessFailed = "auth.create-access-failed"

	TopicGenesisUser       = "auth.genesis-user"
	TopicGenericUserFailed = "auth.generic-user-failed"

	TopicDocumentParsed      = "document.parsed"
	TopicDocumentParseFailed = "document.parse-failed"
)
