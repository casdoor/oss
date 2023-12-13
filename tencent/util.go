// Copyright 2023 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tencent

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

func sha(s string) string {
	sha := sha1.New()
	sha.Write([]byte(s))
	b := sha.Sum(nil)

	return hex.EncodeToString(b)
}

func hmacSha(k, s string) string {
	enc := hmac.New(sha1.New, []byte(k))
	enc.Write([]byte(s))
	b := enc.Sum(nil)

	return hex.EncodeToString(b)
}

func getSignTime() string {
	now := time.Now()
	expired := now.Add(time.Second * 1800)
	return fmt.Sprintf("%d;%d", now.Unix(), expired.Unix())
}

func getSignature(k string, req *http.Request, signTime string) string {
	httpString := fmt.Sprintf("%s\n%s\n%s\n%s\n", strings.ToLower(req.Method),
		req.URL.Path, getParamsStr(req.URL.RawQuery), getHeadStr(req.Header))

	httpString = sha(httpString)
	signKey := hmacSha(k, signTime)
	signStr := fmt.Sprintf("sha1\n%s\n%s\n", signTime, httpString)

	return hmacSha(signKey, signStr)
}
func getHeadKeys(headers http.Header) string {
	if headers == nil || len(headers) == 0 {
		return ""
	}

	tmp := []string{}
	for k := range headers {
		tmp = append(tmp, strings.ToLower(k))
	}
	sort.Strings(tmp)

	return strings.Join(tmp, ";")
}
func getParamsKeys(p string) string {
	if p == "" {
		return ""
	}
	uv, err := url.ParseQuery(p)
	if err != nil {
		return ""
	}
	tmp := []string{}
	for k := range uv {
		tmp = append(tmp, strings.ToLower(k))
	}
	sort.Strings(tmp)

	return strings.Join(tmp, ";")
}
func getHeadStr(headers http.Header) string {
	if headers == nil || len(headers) == 0 {
		return ""
	}

	tmp := []string{}
	for k, v := range headers {
		str := fmt.Sprintf("%s=%s", strings.ToLower(k), escape(v[0]))
		tmp = append(tmp, str)
	}
	sort.Strings(tmp)

	return strings.Join(tmp, "&")
}

func getParamsStr(p string) string {
	if p == "" {
		return ""
	}
	uv, err := url.ParseQuery(p)
	if err != nil {
		return ""
	}
	tmp := []string{}
	for k, v := range uv {
		str := fmt.Sprintf("%s=%s", strings.ToLower(k), escape(v[0]))
		tmp = append(tmp, str)
	}
	sort.Strings(tmp)

	return strings.Join(tmp, "&")
}

func escape(str string) string {
	//go语言中将空格编码为+，需要改为%20
	return strings.Replace(url.QueryEscape(str), "+", "%20", -1)
}
