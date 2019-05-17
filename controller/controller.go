package controller

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/dbsystel/grafana-config-controller/grafana"
	"k8s.io/api/core/v1"
)

type Controller struct {
	logger log.Logger
	g      grafana.APIClient
}

// d something when a configmap created
func (c *Controller) Create(obj interface{}) {
	configmapObj := obj.(*v1.ConfigMap)
	id, _ := configmapObj.Annotations["grafana.net/id"]
	dh, _ := configmapObj.Annotations["grafana.net/dashboard"]
	ds, _ := configmapObj.Annotations["grafana.net/datasource"]
	nc, _ := configmapObj.Annotations["grafana.net/notification-channel"]
	isGrafanaDashboards, _ := strconv.ParseBool(dh)
	isGrafanaDatasource, _ := strconv.ParseBool(ds)
	isGrafanaNotificationChannel, _ := strconv.ParseBool(nc)
	grafanaId,_ := strconv.Atoi(id)
	if  grafanaId == c.g.Id && (isGrafanaDashboards || isGrafanaDatasource || isGrafanaNotificationChannel) {
		var err error
		for k, v := range configmapObj.Data {
			if isGrafanaDatasource {
				level.Info(c.logger).Log("msg", "Creating datasource: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
				err = c.g.CreateDatasource(strings.NewReader(v))
			} else if isGrafanaDashboards {
				if !regexp.MustCompile("^{\\s*\"dashboard\":(\\s*.*)*").MatchString(v) {
					v = "{\n  \"dashboard\":\n    " + strings.TrimSpace(v) + ",\n  \"overwrite\": true\n}"
				}
				fd, _ := configmapObj.Annotations["grafana.net/folder"]
				v, _ = c.checkFolderId(fd, configmapObj, v)
				level.Info(c.logger).Log("msg", "Creating dashboard: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
				err = c.g.CreateDashboard(strings.NewReader(v))
			} else {
				level.Info(c.logger).Log("msg", "Creating notification-channel: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
				err = c.g.CreateNotificationChannel(strings.NewReader(v))
			}

			if err != nil {
				level.Info(c.logger).Log("msg", "Failed to create: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
                level.Error(c.logger).Log("err", err.Error())
			} else {
				level.Info(c.logger).Log("msg", "Succeeded: Created: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
			}
		}
	} else {
		level.Debug(c.logger).Log("msg", "Skipping configmap: " + configmapObj.Name)
	}
}

// do something when a configmap updated
func (c *Controller) Update(oldobj interface{}, newobj interface{}) {
	configmapObj := newobj.(*v1.ConfigMap)
	id, _ := configmapObj.Annotations["grafana.net/id"]
	ds, _ := configmapObj.Annotations["grafana.net/datasource"]
	nc, _ := configmapObj.Annotations["grafana.net/notification-channel"]
	grafanaId,_ := strconv.Atoi(id)
	isGrafanaDatasource, _ := strconv.ParseBool(ds)
	isGrafanaNotificationChannel, _ := strconv.ParseBool(nc)
	if noDifference(oldobj.(*v1.ConfigMap), configmapObj) {
		level.Debug(c.logger).Log("msg", "Skipping automatically updated configmap:" + configmapObj.Name)
		return
	}
	if grafanaId != c.g.Id {
		level.Debug(c.logger).Log("msg", "Skipping configmap:" + configmapObj.Name)
		return
	}
	if isGrafanaNotificationChannel {
		c.updateNotificationChannels(configmapObj)
	} else if isGrafanaDatasource {
		c.updateDatasource(configmapObj)
	} else {
		c.Delete(oldobj)
		c.Create(newobj)
	}
}

// do something when a configmap deleted
func (c *Controller) Delete(obj interface{}) {
	configmapObj := obj.(*v1.ConfigMap)
	id, _ := configmapObj.Annotations["grafana.net/id"]
	dh, _ := configmapObj.Annotations["grafana.net/dashboard"]
	ds, _ := configmapObj.Annotations["grafana.net/datasource"]
	nc, _ := configmapObj.Annotations["grafana.net/notification-channel"]
	isGrafanaDashboards, _ := strconv.ParseBool(dh)
	isGrafanaDatasource, _ := strconv.ParseBool(ds)
	isGrafanaNotificationChannel, _ := strconv.ParseBool(nc)
	grafanaId,_ := strconv.Atoi(id)

	if grafanaId == c.g.Id && (isGrafanaDashboards || isGrafanaDatasource || isGrafanaNotificationChannel){
		var err error
		for k, v := range configmapObj.Data {
			if isGrafanaDatasource {
				level.Info(c.logger).Log("msg", "Deleting datasource: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
				err = c.g.DeleteDatasource(strings.NewReader(v))
			} else if isGrafanaDashboards{
				if !regexp.MustCompile("\\{\\s*\"dashboard\":").MatchString(v) {
					v = "{\n  \"dashboard\":\n    " + strings.TrimSpace(v) + ",\n  \"overwrite\": true\n}"
				}
				level.Info(c.logger).Log("msg", "Deleting dashboard: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
				gd, _ := c.g.SearchDashboard()
				fd, _ := configmapObj.Annotations["grafana.net/folder"]
				v, _ := c.checkFolderId(fd, configmapObj, v)
				uid := c.lookUpUid(gd, strings.NewReader(v))
				level.Debug(c.logger).Log("uid", uid)
				err = c.g.DeleteDashboard(uid)
			} else {
				level.Info(c.logger).Log("msg", "Deleting notification channel: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
				ans, _ := c.g.SearchNotificationChannel()
				an := c.lookUpId(ans,strings.NewReader(v))
				if int(an["id"].(float64)) != -1 {
					err = c.g.DeleteNotificationChannel(int(an["id"].(float64)))
				} else {
					err = errors.New("notification channel not found")
				}
			}
			if err != nil {
				level.Info(c.logger).Log("msg", "Failed to delete: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
				level.Error(c.logger).Log("err", err.Error())
			} else {
				level.Info(c.logger).Log("msg", "Succeeded: Deleted: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
		    }
		}
	} else {
		level.Debug(c.logger).Log("msg", "Skipping configmap:" + configmapObj.Name)
	}
}

// create new Controller instance
func New(g grafana.APIClient, logger log.Logger) *Controller {
	controller := &Controller{}
	controller.logger = logger
	controller.g = g
	return controller
}

// if a dashboard has folder, search the folder in grafana and return the folder id or create a new folder and return the folder id
func (c *Controller) checkFolderId(fd string, configmapObj *v1.ConfigMap, v string) (string, int) {
	fid := 0
	if fd == "" {
		return v, fid
	}
	hasFolder, isString := strconv.ParseBool(fd)
	if (hasFolder && isString == nil) || (!hasFolder && isString != nil) {
		if hasFolder {
			fid = c.searchFolder(configmapObj.Namespace)
		} else {
			fid = c.searchFolder(fd)
		}

		m := make(map[string]interface{})

		err := json.Unmarshal([]byte(v), &m)
		if err != nil {
			level.Error(c.logger).Log("msg", "Format error in dashboard: " + v, "err", err.Error())
		}

		m["folderID"] = fid

		byte_v, err := json.Marshal(m)

		if err != nil {
			level.Error(c.logger).Log("msg", "Format error in dashboard: " + v, "err", err.Error())
		}

		v = string(byte_v)
	}
	return v, fid
}

// search folder id with title, return folder id
func (c *Controller) searchFolder(title string) int {
	fdJson := "{\"title\":\"" + title + "\"}"
	fid := getFolderId(c, fdJson)
	if fid == -1 {
		level.Info(c.logger).Log("msg", "Creating folder: " + title)
		err := c.g.CreateFolder(strings.NewReader(fdJson))
		if err != nil {
			level.Info(c.logger).Log("msg", "Failed to create folder: " + title)
			level.Error(c.logger).Log("err", err.Error())
		} else {
			level.Info(c.logger).Log("msg", "Created folder: " + title)
		}
		fid = getFolderId(c, fdJson)
	}
	return fid
}

// search folder id from a folder json
func getFolderId(c *Controller, fdJson string) int {
	fds, _ := c.g.SearchFolder()
	fd := c.lookUpId(fds, strings.NewReader(fdJson))
	return int(fd["id"].(float64))
}

// are two configmaps same
func noDifference(newConfigMap *v1.ConfigMap, oldConfigMap *v1.ConfigMap) bool {
	if len(newConfigMap.Data) != len(oldConfigMap.Data) {
		return false
	}
	for k, v := range newConfigMap.Data {
		if v != oldConfigMap.Data[k]{
			return false
		}
	}
	if len(newConfigMap.Annotations) != len(oldConfigMap.Annotations) {
		return false
	}
	for k, v := range newConfigMap.Annotations {
		if v != oldConfigMap.Annotations[k] {
			return false
		}
	}
	return true
}

// update notification channels
func (c *Controller) updateNotificationChannels(configmapObj *v1.ConfigMap) {
	var err error
	for k, v := range configmapObj.Data {
		level.Info(c.logger).Log("msg", "Updating notification channel: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
		an, _ := c.g.SearchNotificationChannel()
		newNC := c.lookUpId(an,strings.NewReader(v))
		if int(newNC["id"].(float64)) != -1 {
			nc, _ := json.Marshal(newNC)
			err = c.g.UpdateNotificationChannel(int(newNC["id"].(float64)), strings.NewReader(string(nc)))
		} else {
			err = errors.New("notification channel not found")
		}
		if err != nil {
			level.Info(c.logger).Log("msg", "Failed to update notification channel: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
			level.Error(c.logger).Log("err", err.Error())
			level.Debug(c.logger).Log("msg", "Trying to create notification channel: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
			c.Create(configmapObj)
		} else {
			level.Info(c.logger).Log("msg", "Succeeded: Updated: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
		}
	}
}

// update datesource
func (c *Controller) updateDatasource(configmapObj *v1.ConfigMap) {
	var err error
	for k, v := range configmapObj.Data {
		level.Info(c.logger).Log("msg", "Updating datasource: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
		dss, _ := c.g.SearchDatasource()
		newDS := c.lookUpId(dss,strings.NewReader(v))
		if int(newDS["id"].(float64)) != -1 {
			ds, _ := json.Marshal(newDS)
			err = c.g.UpdateDatasource(int(newDS["id"].(float64)), strings.NewReader(string(ds)))
		} else {
			err = errors.New("datasource not found")
		}
		if err != nil {
			level.Info(c.logger).Log("msg", "Failed to update datasource: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
			level.Error(c.logger).Log("err", err.Error())
			level.Debug(c.logger).Log("msg", "Trying to create datasource: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
			c.Create(configmapObj)
		} else {
			level.Info(c.logger).Log("msg", "Succeeded: Updated: " + k, "configmap", configmapObj.Name, "namespace", configmapObj.Namespace)
		}
	}
}

// search uid of a given dashboard from a list of dashboards
func (c *Controller) lookUpUid(dashboards []grafana.GrafanaDashboard, dashboardJSON io.Reader) string {
	var newDashboard grafana.GrafanaDashboardConfigmap
	err := json.NewDecoder(dashboardJSON).Decode(&newDashboard)
	if err != nil {
		level.Error(c.logger).Log("err", err.Error())
	}
	level.Debug(c.logger).Log("msg", "New dashboard name: " + newDashboard.Dashboard.Title)
	for _,dh := range dashboards {
		if dh.Type == "dash-db" && dh.Title == newDashboard.Dashboard.Title && newDashboard.FolderId == dh.FolderId{
			return dh.Uid
		}
	}
	return ""
}

// search id of a given notification-channel or folder from a list of them and add the id into the json
func (c *Controller)lookUpId(objs []map[string]interface{}, objJSON io.Reader) map[string]interface{} {
	newObj := make(map[string]interface{})
	err := json.NewDecoder(objJSON).Decode(&newObj)
	if err != nil {
		level.Error(c.logger).Log("err", err.Error())
	}
	if newObj["title"] != nil {
		for _,obj := range objs {
			if obj["title"] == nil {
				continue
			}
			if strings.ToUpper(obj["title"].(string)) == strings.ToUpper(newObj["title"].(string)) {
				newObj["id"] = obj["id"]
				return newObj
			}
		}
	} else {
		if newObj["name"] != nil && newObj["type"] != nil {
			for _, obj := range objs {
				if obj["name"] == nil || obj["type"] == nil {
					continue
				}
				if obj["name"].(string) == newObj["name"].(string) && obj["type"].(string) == newObj["type"].(string) {
					newObj["id"] = obj["id"]
					return newObj
				}
			}
		}
	}

	newObj["id"] = float64(-1)
	return newObj
}
