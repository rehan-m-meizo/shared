package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"shared/constants"
	"shared/requests"
	"shared/responses"
	"strings"
	"time"
)

// Fetches Keycloak admin token using client credentials
func getKeyClockTokenResponse() (*responses.KeyClockTokenResponse, error) {
	fmt.Println("🔐 Getting token from:", constants.KeyCloakLoginUrl)

	values := url.Values{
		"client_id":     {constants.KeyCloakClientId},
		"client_secret": {constants.KeyCloakClientSecret},
		"grant_type":    {"client_credentials"},
	}

	req, err := http.NewRequest(http.MethodPost, constants.KeyCloakLoginUrl, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token requests: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute token requests: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("keycloak token fetch failed (%d): %s", resp.StatusCode, string(body))
	}

	var response responses.KeyClockTokenResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &response, nil
}

// Sends a user creation requests to Keycloak
func AddUserToKeyCloak(ctx context.Context, request *requests.KeycloakUserCreateRequest) error {
	log.Println("👤 Creating user in Keycloak")

	tokenResponse, err := getKeyClockTokenResponse()
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal user payload: %w", err)
	}

	uri := fmt.Sprintf("%s/admin/realms/%s/users", constants.KeyCloakBaseURL, constants.KeyCloakRealm)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create user requests: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Keycloak: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		log.Println("✅ User created successfully in Keycloak")
	case http.StatusConflict:
		log.Println("⚠️ User or email already exists in Keycloak")
	case http.StatusBadRequest:
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Bad requests to Keycloak: %s", string(body))
	default:
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❗ Unexpected Keycloak response (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func GetKeyCloakUser(ctx context.Context, email string) (*responses.KeycloakUser, error) {
	log.Println("👤 Gettting user in Keycloak")

	tokenResponse, err := getKeyClockTokenResponse()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	uri := fmt.Sprintf("%s/admin/realms/%s/users?email=%s", constants.KeyCloakBaseURL, constants.KeyCloakRealm, url.QueryEscape(email))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user requests: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, constants.ErrKeyCloaUserNotFound
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user lookup failed (%d): %s", resp.StatusCode, string(body))
	}

	var users []responses.KeycloakUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, constants.ErrKeyCloaUserNotFound
	}

	if len(users) == 0 {
		return nil, constants.ErrKeyCloaUserNotFound
	}

	return &users[0], nil
}

func UpdateUserOnKeycloak(ctx context.Context, userId string, request map[string][]string) error {
	log.Println("👤 Updating user in Keycloak")

	tokenResponse, err := getKeyClockTokenResponse()
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal user payload: %w", err)
	}

	uri := fmt.Sprintf("%s/admin/realms/%s/users/%s", constants.KeyCloakBaseURL, constants.KeyCloakRealm, userId)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uri, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create user requests: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Keycloak: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		log.Println("✅ User updated successfully in Keycloak")
	case http.StatusNotFound:
		log.Println("❌ User not found in Keycloak")
	case http.StatusBadRequest:
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Bad requests to Keycloak: %s", string(body))
	default:
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❗ Unexpected Keycloak response (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

func ResetPasswordOnKeycloak(ctx context.Context, userId string, password string) error {
	log.Println("👤 Resetting password in Keycloak")
	tokenResponse, err := getKeyClockTokenResponse()

	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	credentials := requests.KeycloakUserCredential{
		Type:      "password",
		Value:     password,
		Temporary: false,
	}

	data, err := json.Marshal(credentials)
	if err != nil {
		return fmt.Errorf("failed to marshal user payload: %w", err)
	}

	uri := fmt.Sprintf("%s/admin/realms/%s/users/%s/reset-password", constants.KeyCloakBaseURL, constants.KeyCloakRealm, userId)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create user requests: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Keycloak: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		log.Println("✅ Password reset successfully in Keycloak")
	case http.StatusNotFound:
		log.Println("❌ User not found in Keycloak")
	case http.StatusBadRequest:
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Bad requests to Keycloak: %s", string(body))
	default:
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❗ Unexpected Keycloak response (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}
