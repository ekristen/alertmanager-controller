package controller

import (
	v1 "github.com/ekristen/prom-am-operator/pkg/apis/promam.ekristen.dev/v1"
	"github.com/ekristen/prom-am-operator/pkg/controller/silence"
	"github.com/ibuildthecloud/baaah/pkg/router"
)

func routes(router *router.Router) {
	router.HandleFunc(&v1.Silence{}, silence.ManageSilence)
	//router.Type(&v1.Silence{}).Middleware(appdefinition.RequireNamespace).HandlerFunc(appdefinition.PullAppImage)
}
