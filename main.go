/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package main

import (
	"flag"
	"fmt"

	"net/http"
	"strings"

	"github.com/fidelity/theliv/pkg/auth/authmiddleware"
	"github.com/fidelity/theliv/pkg/auth/samlmethod"
	"github.com/fidelity/theliv/pkg/config"
	err "github.com/fidelity/theliv/pkg/err"
	log "github.com/fidelity/theliv/pkg/log"
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
		log.S().Infof("Will load config from etcd, %v\n", conf)
	}

	conf.LoadConfigs()
	samlmethod.Init()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)

	//set content type as json by default
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Use(authmiddleware.StartAuth)

	// Add panic handling middleware
	r.Use(err.PanicHandler)

	// api route
	r.Route("/theliv-api/v1", router.Route)

	// saml route
	r.Handle("/auth/saml/*", samlmethod.GetSP())

	theliv := config.GetThelivConfig()
	// init default Zap logger
	Logger := log.NewDefaultLogger(log.DefaultLogConfig(theliv.LogLevel))
	defer Logger.Sync()

	err := http.ListenAndServe(fmt.Sprintf(":%v", theliv.Port), r)
	if err != nil {
		log.S().Fatalf("Failed to start server, %v", err)
	}
}
