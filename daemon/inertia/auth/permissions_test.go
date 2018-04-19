package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/ubclaunchpad/inertia/common"

	"github.com/stretchr/testify/assert"
)

func getTestPermissionsHandler(dir string, p ...string) (*PermissionsHandler, error) {
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	var endpoint string
	if len(p) > 0 {
		endpoint = p[0]
	} else {
		endpoint = "/"
	}
	return NewPermissionsHandler(
		path.Join(dir, "users.db"),
		"127.0.0.1", endpoint, 3000,
		getFakeAPIKey,
	)
}

func TestServeHTTPPublicPath(t *testing.T) {
	dir := "./test_perm"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachPublicHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, err := http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServeHTTPPublicPathOnNestedHandler(t *testing.T) {
	// This test emulates the daemon's PermissionsHandler setup
	dir := "./test_perm"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	webPrefix := "/web/"
	ph, err := getTestPermissionsHandler(dir, webPrefix)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()

	// Daemon uses a nested handler
	mux := http.NewServeMux()
	mux.Handle(webPrefix, http.StripPrefix(webPrefix, ph))
	ts.Config.Handler = mux
	ph.AttachPublicHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, err := http.NewRequest("POST", ts.URL+"/web/test", nil)
	assert.Nil(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServeHTTPWithUserReject(t *testing.T) {
	dir := "./test_perm"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachUserRestrictedHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, err := http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestServeHTTPWithUserLoginAndLogout(t *testing.T) {
	dir := "./test_perm"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachUserRestrictedHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Register user
	err = ph.users.AddUser("bobheadxi", "wowgreat", false)
	assert.Nil(t, err)

	// Login in as user, use cookiejar to catch cookie
	user := &common.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
	body, err := json.Marshal(user)
	assert.Nil(t, err)
	req, err := http.NewRequest("POST", ts.URL+"/user/login", bytes.NewReader(body))
	assert.Nil(t, err)
	loginResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Check for cookies
	assert.True(t, len(loginResp.Cookies()) > 0)
	cookie := loginResp.Cookies()[0]
	assert.Equal(t, "ubclaunchpad-inertia", cookie.Name)

	// Attempt to validate
	req, err = http.NewRequest("POST", ts.URL+"/user/validate", nil)
	assert.Nil(t, err)
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Log out
	req, err = http.NewRequest("POST", ts.URL+"/user/logout", nil)
	assert.Nil(t, err)
	req.AddCookie(cookie)
	logoutResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer logoutResp.Body.Close()
	assert.Equal(t, http.StatusOK, logoutResp.StatusCode)

	// Check for cookies
	assert.True(t, len(logoutResp.Cookies()) > 0)
	cookie = logoutResp.Cookies()[0]
	assert.Equal(t, "ubclaunchpad-inertia", cookie.Name)
	assert.Equal(t, -1, cookie.MaxAge)
}

func TestServeHTTPWithUserLoginAndAccept(t *testing.T) {
	dir := "./test_perm"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachUserRestrictedHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Register user
	err = ph.users.AddUser("bobheadxi", "wowgreat", false)
	assert.Nil(t, err)

	// Login in as user, use cookiejar to catch cookie
	user := &common.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
	body, err := json.Marshal(user)
	assert.Nil(t, err)
	req, err := http.NewRequest("POST", ts.URL+"/user/login", bytes.NewReader(body))
	assert.Nil(t, err)
	loginResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Check for cookies
	assert.True(t, len(loginResp.Cookies()) > 0)
	cookie := loginResp.Cookies()[0]
	assert.Equal(t, "ubclaunchpad-inertia", cookie.Name)

	// Attempt to access restricted endpoint with cookie
	req, err = http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServeHTTPDenyNonAdmin(t *testing.T) {
	dir := "./test_perm"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachAdminRestrictedHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Register user
	err = ph.users.AddUser("bobheadxi", "wowgreat", false)
	assert.Nil(t, err)

	// Login in as user, use cookiejar to catch cookie
	user := &common.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
	body, err := json.Marshal(user)
	assert.Nil(t, err)
	req, err := http.NewRequest("POST", ts.URL+"/user/login", bytes.NewReader(body))
	assert.Nil(t, err)
	loginResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Check for cookies
	assert.True(t, len(loginResp.Cookies()) > 0)
	cookie := loginResp.Cookies()[0]
	assert.Equal(t, "ubclaunchpad-inertia", cookie.Name)

	// Attempt to access restricted endpoint with cookie
	req, err = http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestServeHTTPAllowAdmin(t *testing.T) {
	dir := "./test_perm"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph
	ph.AttachAdminRestrictedHandlerFunc("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Register user
	err = ph.users.AddUser("bobheadxi", "wowgreat", true)
	assert.Nil(t, err)

	// Login in as user, use cookiejar to catch cookie
	user := &common.UserRequest{Username: "bobheadxi", Password: "wowgreat"}
	body, err := json.Marshal(user)
	assert.Nil(t, err)
	req, err := http.NewRequest("POST", ts.URL+"/user/login", bytes.NewReader(body))
	assert.Nil(t, err)
	loginResp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer loginResp.Body.Close()
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	// Check for cookies
	assert.True(t, len(loginResp.Cookies()) > 0)
	cookie := loginResp.Cookies()[0]
	assert.Equal(t, "ubclaunchpad-inertia", cookie.Name)

	// Attempt to access restricted endpoint with cookie
	req, err = http.NewRequest("POST", ts.URL+"/test", nil)
	assert.Nil(t, err)
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUserControlHandlers(t *testing.T) {
	dir := "./test_perm"
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Set up permission handler
	ph, err := getTestPermissionsHandler(dir)
	defer os.RemoveAll(dir)
	assert.Nil(t, err)
	defer ph.Close()
	ts.Config.Handler = ph

	// Test handler uses the getFakeAPIToken keylookup, which
	// will match with the testToken
	bearerTokenString := fmt.Sprintf("Bearer %s", testToken)

	// Add a new user
	body, err := json.Marshal(&common.UserRequest{
		Username: "jimmyneutron",
		Password: "asfasdlfjk",
		Admin:    false,
	})
	assert.Nil(t, err)
	payload := bytes.NewReader(body)
	req, err := http.NewRequest("POST", ts.URL+"/user/adduser", payload)
	assert.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerTokenString)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Remove a user
	body, err = json.Marshal(&common.UserRequest{
		Username: "jimmyneutron",
	})
	assert.Nil(t, err)
	payload = bytes.NewReader(body)
	req, err = http.NewRequest("POST", ts.URL+"/user/removeuser", payload)
	assert.Nil(t, err)
	req.Header.Set("Authorization", bearerTokenString)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// List users
	req, err = http.NewRequest("POST", ts.URL+"/user/listusers", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", bearerTokenString)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Reset all users
	req, err = http.NewRequest("POST", ts.URL+"/user/resetusers", nil)
	assert.Nil(t, err)
	req.Header.Set("Authorization", bearerTokenString)
	resp, err = http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
