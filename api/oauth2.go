package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const OAUTH2_URL = "https://auth.tidal.com/v1/oauth2/token"
const DEVICE_CODE_URL = "https://auth.tidal.com/v1/oauth2/device_authorization"

type TokenResp struct {
	ClientName   string `json:"clientName"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	User         struct {
		CountryCode string `json:"countryCode"`
	} `json:"user"`
}

type OAuth2Error struct {
	ErrorDescription string `json:"error_description"`
}

func (a *TidalApi) oauth2Call(grantType string, params url.Values, target interface{}) error {
	finalParams := url.Values{}
	finalParams.Set("client_id", a.ClientId)
	if a.ClientSecret != "" {
		finalParams.Set("client_secret", a.ClientSecret)
	}
	finalParams.Set("grant_type", grantType)
	for key, values := range params {
		finalParams.Set(key, values[0])
	}
	finalParams.Set("scope", "r_usr")

	req, err := http.NewRequest(http.MethodPost, OAUTH2_URL, strings.NewReader(finalParams.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		errorResp := OAuth2Error{}
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return err
		}
		return fmt.Errorf("OAuth2 error, status code %d, message: \"%s\"", resp.StatusCode, errorResp.ErrorDescription)
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

func (a *TidalApi) Refresh() (clientName string, err error) {
	params := url.Values{}
	params.Set("refresh_token", a.refreshToken)

	data := TokenResp{}
	err = a.oauth2Call("refresh_token", params, &data)
	if err != nil {
		return "", err
	}

	clientName = data.ClientName
	a.accessToken = data.AccessToken
	a.tokenExpiration = time.Now().UnixMicro() + int64(data.ExpiresIn)*1000
	a.countryCode = data.User.CountryCode

	err = a.saveToConfig()
	if err != nil {
		return "", err
	}

	return
}

type DeviceCodeResp struct {
	DeviceCode              string `json:"deviceCode"`
	UserCode                string `json:"userCode"`
	VerificationURI         string `json:"verificationUri"`
	VerificationURIComplete string `json:"verificationUriComplete"`
	ExpiresIn               int    `json:"expiresIn"`
	Interval                int    `json:"interval"`
}

func (a *TidalApi) getDeviceCode() (respData *DeviceCodeResp, err error) {
	params := url.Values{}
	params.Set("client_id", a.ClientId)
	if a.ClientSecret != "" {
		params.Set("client_secret", a.ClientSecret)
	}
	params.Set("scope", "r_usr")

	req, err := http.NewRequest(http.MethodPost, DEVICE_CODE_URL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		errorResp := OAuth2Error{}
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("device code request error, status code %d, message: \"%s\"", resp.StatusCode, errorResp.ErrorDescription)
	}

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, err
	}

	return
}

func (a *TidalApi) deviceCodeAuth(deviceCode string) (clientName string, err error) {
	params := url.Values{}
	params.Set("device_code", deviceCode)

	data := TokenResp{}
	err = a.oauth2Call("urn:ietf:params:oauth:grant-type:device_code", params, &data)
	if err != nil {
		return "", err
	}

	clientName = data.ClientName
	a.refreshToken = data.RefreshToken
	a.accessToken = data.AccessToken
	a.tokenExpiration = time.Now().UnixMicro() + int64(data.ExpiresIn)*1000
	a.countryCode = data.User.CountryCode

	err = a.saveToConfig()
	if err != nil {
		return "", err
	}

	return
}

func (a *TidalApi) authCodeAuth(oauthCode string, redirectUri string, codeVerifier string) (clientName string, err error) {
	params := url.Values{}
	params.Set("code", oauthCode)
	params.Set("redirect_uri", redirectUri)
	params.Set("code_verifier", codeVerifier)

	data := TokenResp{}
	err = a.oauth2Call("authorization_code", params, &data)
	if err != nil {
		return "", err
	}

	clientName = data.ClientName
	a.refreshToken = data.RefreshToken
	a.accessToken = data.AccessToken
	a.tokenExpiration = time.Now().UnixMicro() + int64(data.ExpiresIn)*1000
	a.countryCode = data.User.CountryCode

	err = a.saveToConfig()
	if err != nil {
		return "", err
	}

	return
}

func (a *TidalApi) SwitchClient(clientId string) (clientName string, err error) {
	params := url.Values{}
	params.Set("access_token", a.accessToken)

	a.ClientId = clientId

	data := TokenResp{}
	err = a.oauth2Call("switch_client", params, &data)
	if err != nil {
		return "", err
	}

	clientName = data.ClientName
	a.refreshToken = data.RefreshToken
	a.accessToken = data.AccessToken
	a.tokenExpiration = time.Now().UnixMicro() + int64(data.ExpiresIn)*1000
	a.countryCode = data.User.CountryCode

	err = a.saveToConfig()
	if err != nil {
		return "", err
	}

	return
}
