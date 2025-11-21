package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"shared/constants"
	"shared/middleware/claims"
	"shared/pkgs/httpresponses"
	"shared/pkgs/uuids"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

const MaxRequestSize = 10 * 1024 * 1024 // 10 MB
var RequestIDKey = "request_id"         // now string for Gin context

const ClaimsKey = "claims"

func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("%s %s\n", c.Request.Method, c.Request.URL.String())
		c.Next()
	}
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedExactOrigins := map[string]bool{
			"http://localhost:3000":    true,
			"https://localhost:3000":   true,
			"http://localhost:5173":    true,
			"https://lic.meizoerp.com": true,
		}

		originAllowed := false

		if origin == "" {
			originAllowed = true
		} else if allowedExactOrigins[origin] {
			originAllowed = true
		} else if isValidMeizoSubdomain(origin) {
			originAllowed = true
		}

		if originAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, PATCH, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Origin, X-Requested-With, X-CSRF-Token, Authorization, X-Forwarded-Proto")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")

			if c.Request.Method == http.MethodOptions {
				httpresponses.OK(c, nil)
				c.Abort()
				return
			}
		} else {
			httpresponses.Forbidden(c, errors.New("invalid origin: "+origin))
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxRequestSize)
		c.Next()
	}
}

// ✅ Check if origin is a valid HTTPS subdomain of meizohrm.com
func isValidMeizoSubdomain(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if parsed.Scheme != "https" {
		return false
	}

	host := strings.ToLower(parsed.Hostname())

	// allowed root domains
	allowedRoots := []string{"meizohrm.com", "meizoerp.com"}

	for _, root := range allowedRoots {
		if strings.HasSuffix(host, "."+root) {
			// ensure it's not exactly the root domain itself
			if host != root {
				return true
			}
		}
	}

	return false
}

func HstsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none';")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("Referrer-Policy", "no-referrer")
		c.Writer.Header().Set("X-DNS-Prefetch-Control", "off")
		c.Writer.Header().Set("SameSite", "Lax")
		c.Next()
	}
}

func RequestIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//id := uuid.NewSHA1(uuid.NameSpaceURL, []byte(c.Request.URL.String())).String()
		id, err := uuids.NewUUID5(c.Request.URL.String(), uuids.DnsNamespace)
		if err != nil {
			httpresponses.InternalServerError(
				c,
				err,
			)
			return
		}
		c.Writer.Header().Set("X-Request-Id", id)
		c.Set(RequestIDKey, id)
		c.Next()
	}
}

func RequestTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		log.Printf("%s %s %s\n", c.Request.Method, c.Request.URL.String(), duration)
	}
}

func PanicRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic: %v\n", r)
				// Log stack trace for debugging
				log.Printf("Panic stack trace:")
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				log.Printf("%s", buf[:n])
				httpresponses.InternalServerError(c, fmt.Errorf("internal server error"))
			}
		}()
		c.Next()
	}
}

func JWTMiddleware(config claims.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Only allow HS256 (Keycloak uses RS256 by default unless changed)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return config.SigningKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			return
		}

		extracted, err := extractClaims(claims, config.ClientID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set(ClaimsKey, extracted)
		c.Next()
	}
}

func extractClaims(clms jwt.MapClaims, clientID string) (*claims.ExtractedClaims, error) {
	username := clms["preferred_username"]
	email := clms["email"]

	var realmRoles []string
	if realm, ok := clms["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realm["roles"].([]interface{}); ok {
			for _, r := range roles {
				realmRoles = append(realmRoles, fmt.Sprintf("%v", r))
			}
		}
	}

	var clientRoles []string
	if res, ok := clms["resource_access"].(map[string]interface{}); ok {
		if client, ok := res[clientID].(map[string]interface{}); ok {
			if roles, ok := client["roles"].([]interface{}); ok {
				for _, r := range roles {
					clientRoles = append(clientRoles, fmt.Sprintf("%v", r))
				}
			}
		}
	}

	return &claims.ExtractedClaims{
		Username:    fmt.Sprintf("%v", username),
		Email:       fmt.Sprintf("%v", email),
		RealmRoles:  realmRoles,
		ClientRoles: clientRoles,
		Raw:         clms,
	}, nil
}

func BaseURLMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme := "http"
		if forwardedProto := c.GetHeader("X-Forwarded-Proto"); forwardedProto != "" {
			scheme = forwardedProto
		} else if c.Request.TLS != nil {
			scheme = "https"
		}

		host := c.GetHeader("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}

		baseURL := fmt.Sprintf("%s://%s", scheme, host)
		c.Set("BaseURL", baseURL)
		c.Set("BaseScheme", scheme)
		c.Next()
	}
}

// GetClaimsFromContext extracts claims from context
func GetClaimsFromContext(c *gin.Context) (*claims.ExtractedClaims, error) {
	if clm, exists := c.Get(ClaimsKey); exists {
		if casted, ok := clm.(*claims.ExtractedClaims); ok {
			return casted, nil
		}
	}
	return nil, errors.New("no claims in context")
}

func GetBaseURLFromContext(c *gin.Context) (string, string, error) {
	if baseURL, exists := c.Get("BaseURL"); exists {
		if castedBaseURL, ok := baseURL.(string); ok {
			if baseScheme, exists := c.Get("BaseScheme"); exists {
				if castedBaseScheme, ok := baseScheme.(string); ok {
					return castedBaseURL, castedBaseScheme, nil
				}
			}
		}
	}
	return "", "", errors.New("no base url in context")
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing bearer token"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, keycloakJWKS.Keyfunc)
		if err != nil || !token.Valid {
			fmt.Println(err, "ERROr")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token", "details": err.Error()})
			return
		}

		claimsMap, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			return
		}

		var kcClaims claims.KeycloakClaims
		if err := mapToStruct(claimsMap, &kcClaims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to decode claims", "details": err.Error()})
			return
		}

		c.Set("keycloak_claims", &kcClaims)
		c.Next()
	}
}

func checkOpa(claims *claims.KeycloakClaims) (bool, error) {

	input := map[string]interface{}{
		"input": map[string]interface{}{
			"role": claims.Role,
		},
	}

	body, err := json.Marshal(input)

	if err != nil {
		return false, err
	}

	fmt.Println("OPA requests body:", string(body))

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1/data/opa/authz", constants.OpaBaseURL), strings.NewReader(string(body)))

	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("OPA requests failed with status %d", resp.StatusCode)
	}

	var response struct {
		Result struct {
			Allow bool `json:"allow"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode OPA response: %w", err)
	}

	if !response.Result.Allow {
		return false, fmt.Errorf("access denied by OPA policy")
	}

	return true, nil

}
