package main

import (
	"net/http"
	"net/url"

	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"crypto/hmac"
	"log"
)

type APIClient struct {
	Key    string
	Secret string
}

const (
	METHOD_GET  = "GET"
	METHOD_PUT  = "PUT"
	METHOD_POST = "POST"
)

const API_URL = "https://api.zadarma.com/"

func (api APIClient) CallMethod(methodName string, params url.Values, methodType string) interface{} {
	client := http.Client{}

	fullURL := BuildAPIUrl(methodName, params)
	sign := Sign(api, methodName, params)
	request, err := http.NewRequest(methodType, fullURL, nil)
	if err != nil {
		log.Printf("Error when creating newRequest for \"%s\": %s", fullURL, err.Error())
		return nil
	}

	request.Header.Set("Authorization", api.Key+":"+sign)

	response, requesting_err := client.Do(request)
	if requesting_err != nil {
		log.Printf("Error when requesting %s: %s", fullURL, requesting_err.Error())
		return nil
	}

	var result interface{}
	decoder := json.NewDecoder(response.Body)
	decoder.Decode(&result)
	return result
}

func (api APIClient) Callback(from, to, sip string, isPredicted bool) interface{} {
	params := make(url.Values)
	params.Add("from", from)
	params.Add("to", to)
	if sip != "" {
		params.Add("sip", sip)
	}
	if isPredicted {
		params.Add("isPredicted", "1")
	}

	return api.CallMethod("/v1/request/callback/", params, METHOD_GET)
}

// don't work :(
func (api APIClient) ChangeCallerID(sip, callerID string) interface{} {
	params := make(url.Values)
	params.Add("id", sip)
	params.Add("number", callerID)
	return api.CallMethod("/v1/sip/callerid/", params, METHOD_PUT)
}

func (api APIClient) DirectNumbers() interface{} {
	return api.CallMethod("/v1/direct_numbers/", nil, METHOD_GET)
}

func (api APIClient) SIMs() interface{} {
	return api.CallMethod("/v1/sim/", nil, METHOD_GET)
}

//func (api APIClient) SendSMS(number, message, callerID string) interface{} {
//	return nil
//}

func BuildAPIUrl(methodName string, params url.Values) string {
	var resultURL string

	if strings.HasPrefix(methodName, "/") && strings.HasSuffix(API_URL, "/") {
		resultURL = API_URL + strings.TrimLeft(methodName, "/") + "?"
	} else {
		resultURL = API_URL + methodName + "?"
	}

	for name, value := range params {
		resultURL += name + "=" + value[0] + "&"
	}

	resultURL = strings.TrimRight(resultURL, "&")
	return resultURL
}

func Sign(api APIClient, methodName string, params url.Values) string {
	var paramParts []string
	for name, value := range params {
		paramParts = append(paramParts, name+"="+value[0])
	}

	sort.Slice(paramParts, func(i, j int) bool {
		firstRune, _ := utf8.DecodeRuneInString(paramParts[i])
		secondRune, _ := utf8.DecodeRuneInString(paramParts[j])

		return firstRune < secondRune
	})

	paramsUrlStr := strings.Join(paramParts, "&")

	md5Params := fmt.Sprintf("%x", md5.Sum([]byte(paramsUrlStr)))

	sign := methodName + paramsUrlStr + md5Params
	hmacer := hmac.New(sha1.New, []byte(api.Secret))
	hmacer.Write([]byte(sign))
	sha1Result := hmacer.Sum(nil)
	sha1Hash := fmt.Sprintf("%x", sha1Result)

	base64Encoded := base64.StdEncoding.EncodeToString([]byte(sha1Hash))

	return base64Encoded
}
