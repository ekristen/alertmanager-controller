package controller

import (
	"context"

	"github.com/acorn-io/baaah"
	"github.com/acorn-io/baaah/pkg/restconfig"
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/ekristen/alertmanager-controller/pkg/apis/alertmanager.ekristen.dev/v1"
	"github.com/ekristen/alertmanager-controller/pkg/scheme"
	"github.com/ibuildthecloud/baaah/pkg/crds"
	"github.com/rancher/wrangler/pkg/apply"
	"k8s.io/apimachinery/pkg/runtime"
)

type Controller struct {
	Router *router.Router
	Scheme *runtime.Scheme
	apply  apply.Apply
}

func New() (*Controller, error) {
	router, err := baaah.DefaultRouter(scheme.Scheme)
	if err != nil {
		return nil, err
	}

	cfg, err := restconfig.New(scheme.Scheme)
	if err != nil {
		return nil, err
	}

	apply, err := apply.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	routes(router)

	return &Controller{
		Router: router,
		Scheme: scheme.Scheme,
		apply:  apply.WithDynamicLookup(),
	}, nil
}

func (c *Controller) Start(ctx context.Context) error {
	if err := crds.Create(ctx, c.Scheme, v1.SchemeGroupVersion); err != nil {
		return err
	}
	return c.Router.Start(ctx)
}
