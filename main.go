package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/fidelity/theliv/pkg/auth/authmiddleware"
	"github.com/fidelity/theliv/pkg/auth/samlmethod"
	"github.com/fidelity/theliv/pkg/config"
	"github.com/fidelity/theliv/pkg/router"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	var thelivConfig, etcdca, etcdcert, etcdkey, etcdendpoints string
	flag.StringVar(&thelivConfig, "config", "", "theliv -config <full path of theliv.yaml>")
	flag.StringVar(&etcdca, "ca", "", "theliv -ca <full path of etcd ca cert file>")
	flag.StringVar(&etcdcert, "cert", "", "theliv -cert <full path of etcd cert file>")
	flag.StringVar(&etcdkey, "key", "", "theliv -key <full path of etcd key file>")
	flag.StringVar(&etcdendpoints, "endpoints", "", "theliv -endpoints <https://url1:port, https://url2:port, ...>")

	flag.Parse()

	// if thelivConfig == "" {
	// 	thelivConfig = "/etc/theliv-configs/theliv.yaml"
	// }

	// load config
	var conf config.ConfigLoader
	if thelivConfig != "" {
		conf = config.NewFileConfigLoader(thelivConfig)
	} else {
		conf = config.NewEtcdConfigLoader(etcdca, etcdcert, etcdkey, strings.Split(etcdendpoints, ",")...)
		log.Printf("Will load config from etcd, %v\n", conf)
	}

	conf.LoadConfigs()
	samlmethod.Init()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)

	//set content type as json by default
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Use(authmiddleware.StartAuth)

	r.Route("/theliv-api/v1/health", router.HealthCheck)

	// List cluster and namespaces
	r.Route("/theliv-api/v1/clusters", router.Cluster)

	// detector
	r.Route("/theliv-api/v1/detector", router.Detector)

	// userinfo
	r.Route("/theliv-api/v1/userinfo", router.Userinfo)

	// rbac
	r.Route("/theliv-api/v1/rbac", router.Rbac)

	// saml route
	r.Handle("/auth/saml/*", samlmethod.GetSP())

	theliv := config.GetThelivConfig()
	err := http.ListenAndServe(fmt.Sprintf(":%v", theliv.Port), r)
	if err != nil {
		log.Fatalf("Failed to start server, %v", err)
	}
}
