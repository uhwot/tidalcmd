package api

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/119.0"

func RefreshTokenAuth(clientId string, clientSecret string, refreshToken string) (api *TidalApi, clientName string, err error) {
	api = &TidalApi{ClientId: clientId, ClientSecret: clientSecret, refreshToken: refreshToken}
	clientName, err = api.Refresh()
	return
}

func DeviceAuth(clientId string, clientSecret string) (api *TidalApi, clientName string, err error) {
	api = &TidalApi{ClientId: clientId, ClientSecret: clientSecret}

	for {
		deviceCodeResp, err := api.getDeviceCode()
		if err != nil {
			return nil, "", err
		}

		deviceCodeExpiry := time.Now().Unix() + int64(deviceCodeResp.ExpiresIn)
		fmt.Printf("Auth URL: https://%s\n", deviceCodeResp.VerificationURIComplete)

		for deviceCodeExpiry > time.Now().Unix() {
			var authErr error
			clientName, authErr = api.deviceCodeAuth(deviceCodeResp.DeviceCode)
			if authErr == nil {
				break
			}
			fmt.Println(authErr)
			time.Sleep(time.Duration(deviceCodeResp.Interval) * time.Second)
		}

		if api.refreshToken != "" {
			break
		}
	}

	return
}

func getJson(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

// https://jvt.me/posts/2022/04/07/go-cookie-header/
func setCookieHeader(cookie string) []*http.Cookie {
	header := http.Header{}
	header.Add("Set-Cookie", cookie)
	req := http.Response{Header: header}
	return req.Cookies()
}

func getCsrfToken(jar http.CookieJar) (string, error) {
	for _, cookie := range jar.Cookies(&url.URL{Scheme: "https", Host: "login.tidal.com"}) {
		if cookie.Name == "_csrf-token" {
			return cookie.Value, nil
		}
	}
	return "", errors.New("CSRF token not present")
}

type ddResponse struct {
	Cookie string `json:"cookie"`
}

type emailResponse struct {
	IsValidEmail bool `json:"isValidEmail"`
	NewUser      bool `json:"newUser"`
}

// https://stackoverflow.com/a/37533144
//func arrayToString(a []byte, delim string) string {
//	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
//}

func WebAuth(email string, password string, clientId string, clientSecret string, redirectUri string, appMode string) (*TidalApi, string, error) {
	//rawUniqueKey := make([]byte, 8)
	//rand.Read(rawUniqueKey)
	//uniqueKey := hex.EncodeToString(rawUniqueKey)
	rawCodeVerifier := make([]byte, 32)
	rand.Read(rawCodeVerifier)
	codeVerifier := base64.RawURLEncoding.EncodeToString(rawCodeVerifier)
	rawCodeChallenge := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(rawCodeChallenge[:])

	//rawState := make([]byte, 32)
	//rand.Read(rawState)
	//stateStr := fmt.Sprintf("TIDAL_%d_%s", time.Now().UnixMilli(), base64.RawURLEncoding.EncodeToString([]byte(arrayToString(rawState, ","))))
	//fmt.Println(stateStr)

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, "", err
	}

	// https://stackoverflow.com/a/38150816
	client := &http.Client{
		Jar: cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	params := url.Values{}
	if appMode != "" {
		params.Set("appMode", appMode)
	}
	params.Set("client_id", clientId)
	//params.Set("client_unique_key", uniqueKey)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")
	params.Set("lang", "en")
	params.Set("redirect_uri", redirectUri)
	params.Set("response_type", "code")

	queryString := params.Encode()

	loginUrl := "https://login.tidal.com/authorize?" + queryString

	// csrf token
	req, err := http.NewRequest("GET", loginUrl, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Add("accept-language", "en-US")
	req.Header.Add("user-agent", USER_AGENT)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	switch resp.StatusCode {
	case 200:
		break
	case 400:
		return nil, "", errors.New("invalid client")
	case 403:
		return nil, "", errors.New("bot protection triggered")
	default:
		return nil, "", fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	payload := url.Values{}
	payload.Add("jsData", fmt.Sprintf(`{"opts":"endpoint,ajaxListenerPath","ua":"%s"}`, USER_AGENT))
	payload.Add("ddk", "1F633CDD8EF22541BD6D9B1B8EF13A")
	payload.Add("Referer", url.QueryEscape(loginUrl))
	payload.Add("responsePage", "origin")
	payload.Add("ddv", "4.15.0")

	body := strings.NewReader(payload.Encode())
	req, err = http.NewRequest("POST", "https://dd.tidal.com/js/", body)
	if err != nil {
		return nil, "", err
	}
	req.Header.Add("accept", "*/*")
	req.Header.Add("accept-language", "en-US")
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	req.Header.Add("user-agent", USER_AGENT)

	resp, err = client.Do(req)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode != 200 {
		return nil, "", errors.New("bot protection triggered on DataDome request")
	}

	ddData := ddResponse{}
	err = getJson(resp, &ddData)
	if err != nil {
		return nil, "", err
	}

	if ddData.Cookie == "" {
		return nil, "", errors.New("bot protection triggered on DataDome request")
	}

	client.Jar.SetCookies(&url.URL{Scheme: "https", Host: "login.tidal.com"}, setCookieHeader(ddData.Cookie))

	csrfToken, err := getCsrfToken(client.Jar)
	if err != nil {
		return nil, "", err
	}

	jsonPayload, err := json.Marshal(map[string]string{"email": email})
	if err != nil {
		return nil, "", err
	}
	req, err = http.NewRequest("POST", "https://login.tidal.com/api/email?"+queryString, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, "", err
	}

	req.Header.Add("accept", "application/json, text/plain, */*")
	req.Header.Add("accept-language", "en-US")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("user-agent", USER_AGENT)
	req.Header.Add("x-csrf-token", csrfToken)

	resp, err = client.Do(req)
	if err != nil {
		return nil, "", err
	}

	switch resp.StatusCode {
	case 200:
		break
	case 403:
		return nil, "", errors.New("bot protection triggered")
	default:
		return nil, "", fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	emailData := emailResponse{}
	err = getJson(resp, &emailData)
	if err != nil {
		return nil, "", err
	}

	if !emailData.IsValidEmail {
		return nil, "", errors.New("invalid email")
	}
	if emailData.NewUser {
		return nil, "", errors.New("email doesn't exist")
	}

	csrfToken, err = getCsrfToken(client.Jar)
	if err != nil {
		return nil, "", err
	}

	jsonPayload, err = json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		return nil, "", err
	}
	req, err = http.NewRequest("POST", "https://login.tidal.com/api/email/user/existing?"+queryString, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, "", err
	}

	req.Header.Add("accept", "application/json, text/plain, */*")
	req.Header.Add("accept-language", "en-US")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("user-agent", USER_AGENT)
	req.Header.Add("x-csrf-token", csrfToken)

	resp, err = client.Do(req)
	if err != nil {
		return nil, "", err
	}

	switch resp.StatusCode {
	case 200:
		break
	case 403:
		return nil, "", errors.New("bot protection triggered")
	default:
		return nil, "", fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	req, err = http.NewRequest("GET", "https://login.tidal.com/success", nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Add("accept-language", "en-US")
	req.Header.Add("user-agent", USER_AGENT)

	resp, err = client.Do(req)
	if err != nil {
		return nil, "", err
	}

	oauthCodeUrlStr := ""

	switch resp.StatusCode {
	case 302:
		oauthCodeUrlStr = resp.Header.Get("location")
	case 200: // when the redirect url isnt http/https
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		re := regexp.MustCompile(`successRedirectUrl:("[\w\\\-\?&=:.]+")`)
		match := re.FindStringSubmatch(string(body))
		if match == nil {
			return nil, "", errors.New("redirect URL match not found")
		}
		oauthCodeUrlStr, err = strconv.Unquote(match[1])
		if err != nil {
			return nil, "", err
		}
	case 401:
		return nil, "", errors.New("incorrect password")
	default:
		return nil, "", fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	oauthCodeUrl, err := url.Parse(oauthCodeUrlStr)
	if err != nil {
		return nil, "", err
	}

	oauthCode := oauthCodeUrl.Query().Get("code")

	api := &TidalApi{ClientId: clientId, ClientSecret: clientSecret}
	clientName, err := api.authCodeAuth(oauthCode, redirectUri, codeVerifier)
	if err != nil {
		return nil, "", err
	}

	return api, clientName, nil
}
