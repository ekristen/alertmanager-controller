package controller

import (
	"context"
	"time"

	"github.com/acorn-io/baaah"
	"github.com/acorn-io/baaah/pkg/apply"
	"github.com/acorn-io/baaah/pkg/restconfig"
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/ekristen/alertmanager-controller/pkg/apis/alertmanager.ekristen.dev/v1"
	"github.com/ekristen/alertmanager-controller/pkg/crds"
	"github.com/ekristen/alertmanager-controller/pkg/scheme"
	"k8s.io/apimachinery/pkg/runtime"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Controller struct {
	Router *router.Router
	Scheme *runtime.Scheme
	apply  apply.Apply
}

type ControllerOpts struct {
	GCExpired      bool
	GCExpiredDelay time.Duration
}

func New(opts *ControllerOpts) (*Controller, error) {
	router, err := baaah.DefaultRouter(scheme.Scheme)
	if err != nil {
		return nil, err
	}

	cfg, err := restconfig.New(scheme.Scheme)
	if err != nil {
		return nil, err
	}

	client, err := kclient.New(cfg, kclient.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		return nil, err
	}

	apply := apply.New(client)

	routes(router, client, opts)

	return &Controller{
		Router: router,
		Scheme: scheme.Scheme,
		apply:  apply,
	}, nil
}

func (c *Controller) Start(ctx context.Context) error {
	if err := crds.Create(ctx, c.Scheme, v1.SchemeGroupVersion); err != nil {
		return err
	}
	return c.Router.Start(ctx)
}
