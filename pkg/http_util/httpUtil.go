package http_util

import (
	"bytes"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/tristan-club/bot-wizard/pkg/util"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func Get(url string, headers map[string]string) ([]byte, error) {

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func GetJSON(url string, headers map[string]string, obj interface{}) error {
	b, err := Get(url, headers)

	if err != nil {
		return err
	}

	//var d interface{}
	err = json.Unmarshal(b, &obj)
	if err != nil {
		return err
	}
	return nil
}

func Post(url string, params interface{}, headers map[string]string) ([]byte, error) {

	bytesData, err := json.Marshal(params)

	if err != nil {
		return nil, err
	}
	client := http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bytesData))
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("type", "API").
		Int("status", res.StatusCode).
		Str("url", url).
		Fields(map[string]interface{}{
			"post":    params,
			"headers": headers,
			"result":  content,
		}).
		Send()

	return content, nil
}

func PostJSON(url string, params interface{}, headers map[string]string, ret interface{}) error {

	if headers == nil {
		headers = make(map[string]string, 0)
	}
	headers["Content-Type"] = "application/json"
	b, err := Post(url, params, headers)

	if err != nil {
		log.Error().Err(err).Send()
		return err
	}

	if !util.IsNil(ret) {
		err = json.Unmarshal(b, ret)
	}

	return err
}

func PostForm(thisUrl string, params map[string]string, headers map[string]string, result interface{}) error {

	headers["Content-Type"] = "application/x-www-form-urlencoded"
	values := url.Values{}
	a := []string{}
	for k, v := range params {
		a = append(a, v)
		values[k] = a
		a = []string{}
	}

	res, err := http.PostForm(thisUrl, values)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}
	log.Info().Fields(map[string]interface{}{
		"action":   "post form",
		"url":      thisUrl,
		"response": result,
	}).Send()

	return err
}
