package controller

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/ekristen/alertmanager-controller/pkg/apis/alertmanager.ekristen.dev/v1"
	"github.com/ekristen/alertmanager-controller/pkg/controller/silence"
)

func routes(router *router.Router) {
	router.Type(&v1.Silence{}).Middleware(silence.SkipExpired).Middleware(silence.SkipInvalidSpec).HandlerFunc(silence.ManageSilence)
}
