package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const BASE_URL = "https://api.tidal.com/v1/"

type TidalApi struct {
	ClientId        string
	ClientSecret    string
	refreshToken    string
	accessToken     string
	tokenExpiration int64
	countryCode     string
}

type ApiError struct {
	UserMessage string `json:"userMessage"`
}

func NewClientFromConfig() (api *TidalApi, err error) {
	api = &TidalApi{
		viper.GetString("client_id"),
		viper.GetString("client_secret"),
		viper.GetString("refresh_token"),
		viper.GetString("access_token"),
		viper.GetInt64("expiration_ts"),
		viper.GetString("country_code"),
	}
	if api.tokenExpiration <= time.Now().UnixMilli() {
		_, err = api.Refresh()
	}
	return
}

func (a *TidalApi) saveToConfig() error {
	viper.Set("client_id", a.ClientId)
	viper.Set("client_secret", a.ClientSecret)
	viper.Set("refresh_token", a.refreshToken)
	viper.Set("access_token", a.accessToken)
	viper.Set("expiration_ts", a.tokenExpiration)
	viper.Set("country_code", a.countryCode)
	return viper.WriteConfig()
}

func (a *TidalApi) call(method string, path string, params url.Values, target interface{}) error {
	var body io.Reader
	if method != http.MethodGet {
		body = strings.NewReader(params.Encode())
	}

	req, err := http.NewRequest(method, BASE_URL+path, body)
	if err != nil {
		return err
	}

	if method == http.MethodGet {
		req.URL.RawQuery = params.Encode()
	} else {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if a.tokenExpiration <= time.Now().UnixMilli() {
		_, err := a.Refresh()
		if err != nil {
			return err
		}
	}

	req.Header.Set("Authorization", "Bearer "+a.accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		errorResp := ApiError{}
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&errorResp)
		if err != nil {
			return err
		}
		return fmt.Errorf("API error, status code %d, message: \"%s\"", resp.StatusCode, errorResp.UserMessage)
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}
