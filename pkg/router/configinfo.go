/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package router

import (
	"net/http"

	"github.com/fidelity/theliv/pkg/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func ConfigInfo(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		config := getUIConfig()
		if empty := processEmpty(w, r, config); !empty {
			render.Respond(w, r, config)
		}
	})
}

func getUIConfig() *ConfigData {
	thelivcfg := config.GetThelivConfig()
	return &ConfigData{
		EmailAddr:       thelivcfg.EmailAddr,
		DevelopedByTeam: thelivcfg.DevelopedByTeam,
		VideoLink:       thelivcfg.VideoLink,
	}
}

type ConfigData struct {
	EmailAddr       string `json:"emailAddr,omitempty"`
	DevelopedByTeam string `json:"developedByTeam,omitempty"`
	VideoLink       string `json:"videoLink,omitempty"`
}
