package controller

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/ekristen/alertmanager-controller/pkg/apis/alertmanager.ekristen.dev/v1"
	"github.com/ekristen/alertmanager-controller/pkg/controller/silence"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func routes(router *router.Router, client kclient.Client, opts *ControllerOpts) {
	router.Type(&v1.Silence{}).
		Middleware(silence.AttachClient(client)).
		Middleware(silence.SetDefaults).
		Middleware(silence.SetExpired).
		Middleware(silence.SkipExpired(opts.GCExpired, opts.GCExpiredDelay)).
		Middleware(silence.SkipInvalidSpec).
		HandlerFunc(silence.ManageSilence)

	router.Type(&v1.Silence{}).
		FinalizeFunc("alertmanager.ekristen.dev/silence", silence.RemoveSilence)
}
