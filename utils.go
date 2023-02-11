package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
)

func getSortedHeaders(r *http.Request, sep string) string {
	reqHeaderKeys := []string{"Host"}
	for k := range r.Header {
		reqHeaderKeys = append(reqHeaderKeys, k)
	}
	sort.Strings(reqHeaderKeys)
	reqHeaders := []string{}
	for _, k := range reqHeaderKeys {
		if k == "Host" {
			reqHeaders = append(reqHeaders, fmt.Sprintf("Host: %s", r.Host))
			continue
		}
		reqHeaders = append(reqHeaders, fmt.Sprintf("%s: %s", k, r.Header.Get(k)))
	}
	return strings.Join(reqHeaders, sep)
}

func getStringMD5(s string) string {
	return string(fmt.Sprintf("%x", md5.Sum([]byte(s))))
}

func getNodeId() string {
	host, _ := os.Hostname()
	return getStringMD5(host)[:7]
}
