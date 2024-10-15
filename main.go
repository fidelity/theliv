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

	"github.com/fidelity/theliv/pkg/auth/oidcmethod"
	"github.com/fidelity/theliv/pkg/config"
	log "github.com/fidelity/theliv/pkg/log"
	"github.com/fidelity/theliv/pkg/router"
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
	oidcmethod.InitAuth()

	r := router.NewRouter()

	theliv := config.GetThelivConfig()
	// init default Zap logger
	Logger := log.NewDefaultLogger(log.DefaultLogConfig(theliv.LogLevel))
	defer Logger.Sync()

	err := http.ListenAndServe(fmt.Sprintf(":%v", theliv.Port), r)
	if err != nil {
		log.S().Fatalf("Failed to start server, %v", err)
	}
}
