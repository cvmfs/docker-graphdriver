package lib

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

func getAuthParams(authHeader string) map[string]string {
	params := make(map[string]string)

	for _, v := range strings.Split(authHeader, ",") {
		s := strings.Split(v, "=")
		cutset := "\""

		if s[0] == "Bearer realm" {
			params["realm"] = strings.Trim(s[1], cutset)
		} else if s[0] == "service" {
			params["service"] = strings.Trim(s[1], cutset)
		} else if s[0] == "scope" {
			params["scope"] = strings.Trim(s[1], cutset)
		}
	}

	return params
}

func getToken(authParams map[string]string) string {
	tokenUrl := authParams["realm"] + "?service=" + authParams["service"] +
		"&scope=" + authParams["scope"]

	resp, _ := http.Get(tokenUrl)

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var m TokenMessage
	json.Unmarshal(body, &m)

	return m.Token
}
