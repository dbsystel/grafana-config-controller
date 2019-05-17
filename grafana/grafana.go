// Copyright 2016 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grafana

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type APIClient struct {
	BaseUrl    *url.URL
	HTTPClient *http.Client
	Id         int
	logger     log.Logger
}

type GrafanaDashboard struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Uri      string `json:"uri"`
	Uid      string `json:"uid"`
	Type     string `json:"type"`
	FolderId int    `json:"folderId"`
}

type GrafanaDashboardConfigmap struct {
	Dashboard  struct{ Title string } `json:"dashboard"`
	FolderId int    `json:"folderId"`
}

// return a list of grafana dashboards
func (c *APIClient) SearchDashboard() ([]GrafanaDashboard, error) {
	searchUrl := makeUrl(c.BaseUrl, "/api/search")
	resp, err := c.HTTPClient.Get(searchUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	searchResult := make([]GrafanaDashboard, 0)
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		return nil, err
	}

	return searchResult, nil
}

// return a list of grafana datasources
func (c *APIClient) SearchDatasource() ([]map[string]interface{}, error) {
	searchUrl := makeUrl(c.BaseUrl, "/api/datasources")
	resp, err := c.HTTPClient.Get(searchUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	searchResult := make([]map[string]interface{}, 0)
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		level.Error(c.logger).Log("err", err.Error())
		return nil, err
	}

	return searchResult, nil
}

//return a list of notification channels
func (c *APIClient) SearchNotificationChannel() ([]map[string]interface{}, error) {
	searchUrl := makeUrl(c.BaseUrl, "/api/alert-notifications")
	resp, err := c.HTTPClient.Get(searchUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	searchResult := make([]map[string]interface{}, 0)
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		level.Error(c.logger).Log("err", err.Error())
		return nil, err
	}

	return searchResult, nil
}

// return a list of folders
func (c *APIClient) SearchFolder() ([]map[string]interface{}, error) {
	searchUrl := makeUrl(c.BaseUrl, "/api/folders")
	resp, err := c.HTTPClient.Get(searchUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	searchResult := make([]map[string]interface{}, 0)
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		level.Error(c.logger).Log("err", err.Error())
		return nil, err
	}

	return searchResult, nil
}

func (c *APIClient) DeleteDashboard(uid string) error {

	deleteUrl := makeUrl(c.BaseUrl, "/api/dashboards/uid/"+uid)
	req, err := http.NewRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req)
}

func (c *APIClient) DeleteDatasource(datasourceJSON io.Reader) error {
	datasource := make(map[string]interface{})
	err := json.NewDecoder(datasourceJSON).Decode(&datasource)
	if err != nil {
		level.Error(c.logger).Log("err", err.Error())
	}
	deleteUrl := makeUrl(c.BaseUrl, "/api/datasources/name/"+datasource["name"].(string))
	req, err := http.NewRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req)
}

func (c *APIClient) DeleteNotificationChannel(id int) error {

	deleteUrl := makeUrl(c.BaseUrl, "/api/alert-notifications/"+strconv.Itoa(id))
	req, err := http.NewRequest("DELETE", deleteUrl, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req)
}

func (c *APIClient) UpdateDatasource(id int, datasourceJson io.Reader) error {
	updateUrl := makeUrl(c.BaseUrl, "/api/datasources/"+strconv.Itoa(id))
	return c.doPut(updateUrl, datasourceJson)
}


func (c *APIClient) UpdateNotificationChannel(id int, notificationChannelJSON io.Reader) error {
	updateUrl := makeUrl(c.BaseUrl, "/api/alert-notifications/"+strconv.Itoa(id))
	return c.doPut(updateUrl, notificationChannelJSON)
}


func (c *APIClient) CreateDashboard(dashboardJSON io.Reader) error {
	return c.doPost(makeUrl(c.BaseUrl, "/api/dashboards/db"), dashboardJSON)
}

func (c *APIClient) CreateDatasource(datasourceJSON io.Reader) error {
	return c.doPost(makeUrl(c.BaseUrl, "/api/datasources"), datasourceJSON)
}

func (c *APIClient) CreateNotificationChannel(notificationChannelJSON io.Reader) error {
	return c.doPost(makeUrl(c.BaseUrl, "/api/alert-notifications"), notificationChannelJSON)
}

func (c *APIClient) CreateFolder(folderJSON io.Reader) error {
	return c.doPost(makeUrl(c.BaseUrl, "/api/folders"), folderJSON)
}

func (c *APIClient) CreateUser(userJSON io.Reader) error {
	return c.doPost(makeUrl(c.BaseUrl, "/api/admin/users"), userJSON)
}

func (c *APIClient) doPut(url string, dataJSON io.Reader) error {
	req, err := http.NewRequest("PUT", url, dataJSON)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	if os.Getenv("GRAFANA_BEARER_TOKEN") != "" {
		req.Header.Add("Authorization", "Bearer " + os.Getenv("GRAFANA_BEARER_TOKEN"))
	}
	return c.doRequest(req)
}

func (c *APIClient) doPost(url string, dataJSON io.Reader) error {
	req, err := http.NewRequest("POST", url, dataJSON)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	if os.Getenv("GRAFANA_BEARER_TOKEN") != "" {
		req.Header.Add("Authorization", "Bearer " + os.Getenv("GRAFANA_BEARER_TOKEN"))
	}

	return c.doRequest(req)
}

func (c *APIClient) doRequest(req *http.Request) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		for strings.Contains(err.Error(), "connection refused") {
			level.Error(c.logger).Log("err", err.Error())
			level.Info(c.logger).Log("msg", "Perhaps Grafana is not ready. Waiting for 8 seconds and retry again...")
			time.Sleep(8 * time.Second)
			resp, err = c.HTTPClient.Do(req)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status code returned from Grafana API (got: %d, expected: 200, msg:%s)", resp.StatusCode, string(response))
	}
	return nil
}

// return a new APIClient
func New(baseUrl *url.URL, id int, logger log.Logger) *APIClient {
	return &APIClient{
		BaseUrl:    baseUrl,
		HTTPClient: http.DefaultClient,
		Id:         id,
		logger:     logger,
	}
}

// build url with grafana url and api endpoint
func makeUrl(baseURL *url.URL, endpoint string) string {
	result := *baseURL

	result.Path = path.Join(result.Path, endpoint)

	return result.String()
}
