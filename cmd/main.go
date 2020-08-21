package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dbsystel/grafana-config-controller/controller"
	"github.com/dbsystel/grafana-config-controller/grafana"
	"github.com/dbsystel/kube-controller-dbsystel-go-common/controller/configmap"
	"github.com/dbsystel/kube-controller-dbsystel-go-common/kubernetes"
	k8sflag "github.com/dbsystel/kube-controller-dbsystel-go-common/kubernetes/flag"
	opslog "github.com/dbsystel/kube-controller-dbsystel-go-common/log"
	logflag "github.com/dbsystel/kube-controller-dbsystel-go-common/log/flag"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New(filepath.Base(os.Args[0]), "Grafana Controller")
	//Here you can define more flags for your application
	grafanaUrl = app.Flag("grafana-url", "The url to issue requests to update dashboards to.").Required().String()
	id         = app.Flag("id", "The grafana id to issue requests to update dashboards to.").Default("0").Int()
	namespace  = app.Flag("namespace", "The namespace to watching.").Default("").String()
)

func main() {
	//Define config for logging
	var logcfg opslog.Config
	//Definie if controller runs outside of k8s
	var runOutsideCluster bool
	//Add two additional flags to application for logging and decision if inside or outside k8s
	logflag.AddFlags(app, &logcfg)
	k8sflag.AddFlags(app, &runOutsideCluster)
	//Parse all arguments
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		//Received error while parsing arguments from function app.Parse
		fmt.Fprintln(os.Stderr, "Catched the following error while parsing arguments: ", err)
		app.Usage(os.Args[1:])
		os.Exit(2)
	}

	//Initialize new logger from opslog
	logger, err := opslog.New(logcfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		app.Usage(os.Args[1:])
		os.Exit(2)
	}
	//First usage of initialized logger for testing
	level.Debug(logger).Log("msg", "Logging initiated...")
	//Initialize new k8s client from common k8s package
	k8sClient, err := kubernetes.NewClientSet(runOutsideCluster)
	if err != nil {
		level.Error(logger).Log("msg", err.Error())
		os.Exit(2)
	}

	gUrl, err := url.Parse(*grafanaUrl)
	if err != nil {
		level.Error(logger).Log("msg", "Grafana URL could not be parsed: "+*grafanaUrl)
		os.Exit(2)
	}

	if os.Getenv("GRAFANA_USER") != "" && os.Getenv("GRAFANA_PASSWORD") == "" {
		gUrl.User = url.User(os.Getenv("GRAFANA_USER"))
	}

	if os.Getenv("GRAFANA_USER") != "" && os.Getenv("GRAFANA_PASSWORD") != "" {
		gUrl.User = url.UserPassword(os.Getenv("GRAFANA_USER"), os.Getenv("GRAFANA_PASSWORD"))
	}

	g := grafana.New(gUrl, *id, logger)

	if os.Getenv("MONITORING_PASSWORD") != "" {
		createMonitoringUser(g, logger)
	}

	sigs := make(chan os.Signal, 1) // Create channel to receive OS signals
	stop := make(chan struct{})     // Create channel to receive stop signal

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT) // Register the sigs channel to receieve SIGTERM

	wg := &sync.WaitGroup{} // Goroutines can add themselves to this to be waited on so that they finish

	//Initialize new k8s configmap-controller from common k8s package
	configMapController := &configmap.ConfigMapController{}
	configMapController.Controller = controller.New(*g, logger)
	configMapController.Initialize(k8sClient, *namespace)
	//Run initiated configmap-controller as go routine
	go configMapController.Run(stop, wg)

	<-sigs // Wait for signals (this hangs until a signal arrives)

	level.Info(logger).Log("msg", "Shutting down...")

	close(stop) // Tell goroutines to stop themselves
	wg.Wait()   // Wait for all to be stopped
}

func createMonitoringUser(g *grafana.APIClient, logger log.Logger) {
	level.Info(logger).Log("msg", "Creating monitoring user...")
	user := "{\"name\": \"Monitoring\", \"login\":\"monitoring\", \"password\": \"" + os.Getenv("MONITORING_PASSWORD") + "\", \"role\": \"Viewer\"}"
	err := g.CreateUser(strings.NewReader(user))
	if err != nil {
		for strings.Contains(err.Error(), "timeout") {
			level.Error(logger).Log("err", err.Error())
			level.Info(logger).Log("msg", "Perhaps Grafana is not ready. Waiting for 8 seconds and retry again...")
			time.Sleep(8 * time.Second)
			err = g.CreateUser(strings.NewReader(user))
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		level.Info(logger).Log("msg", "Failed to create monitoring user")
		level.Error(logger).Log("err", err.Error())
	} else {
		level.Info(logger).Log("msg", "Succeeded: Created monitoring user")
	}
}
