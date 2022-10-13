package silence

import (
	"context"
	"fmt"
	"time"

	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/ekristen/alertmanager-controller/pkg/apis/alertmanager.ekristen.dev/v1"
	"github.com/ekristen/alertmanager-controller/pkg/condition"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type SilenceResponse struct {
	v1.SilenceSpec `json:",inline"`
	ID             string           `json:"id"`
	Status         v1.SilenceStatus `json:"status"`
}

type SilenceCreateResponse struct {
	SilenceID string `json:"silenceID"`
}

type ContextKey string

var clientKey ContextKey = "client"

func AttachClient(client kclient.Client) router.Middleware {
	return (func(h router.Handler) router.Handler {
		return router.HandlerFunc(func(req router.Request, resp router.Response) error {
			req.Ctx = context.WithValue(req.Ctx, clientKey, client)

			return h.Handle(req, resp)
		})
	})
}

func SkipExpired(h router.Handler) router.Handler {
	return router.HandlerFunc(func(req router.Request, resp router.Response) error {
		silence := req.Object.(*v1.Silence)

		if silence.Status.State == "expired" {
			logrus.Info("handle: expired")
			return nil
		}

		return h.Handle(req, resp)
	})
}

func SkipInvalidSpec(h router.Handler) router.Handler {
	return router.HandlerFunc(func(req router.Request, resp router.Response) error {
		silence := req.Object.(*v1.Silence)
		cond := condition.Setter(silence, resp, "managed")

		logrus.Info("handle: start")

		if silence.Spec.URL == "" {
			cond.Error(fmt.Errorf("invalid or missing alertmanager url"))
			resp.Objects(silence)
			resp.RetryAfter(time.Minute * 2)
			silence.Status.State = "invalid-missing-url"
			resp.Objects(silence)

			return nil
		}

		return h.Handle(req, resp)
	})
}

func ManageSilence(req router.Request, resp router.Response) error {
	silence := req.Object.(*v1.Silence)
	cond := condition.Setter(silence, resp, "managed")

	if silence.Status.ID == "" {
		time.Sleep(1 * time.Second)

		id := uuid.New()
		fmt.Println("called>>", id.String())
		silence.Status.ID = id.String()
		silence.Status.State = "working"
	}

	resp.Objects(silence)

	logrus.Info("handle: saving silence state")

	cond.Success()

	if silence.Status.State == "pending" || silence.Status.State == "active" {
		resp.RetryAfter(time.Minute * 5)
	}

	return nil
}

func RemoveSilence(req router.Request, resp router.Response) error {

	time.Sleep(1)

	return nil
}

func ManageSilence2(req router.Request, resp router.Response) error {
	return nil
}
