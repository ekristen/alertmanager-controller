package silence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/ekristen/alertmanager-controller/pkg/apis/alertmanager.ekristen.dev/v1"
	"github.com/ekristen/alertmanager-controller/pkg/condition"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Response struct {
	v1.SilenceSpec `json:",inline"`
	ID             string           `json:"id"`
	Status         v1.SilenceStatus `json:"status"`
}

type CreateResponse struct {
	SilenceID string `json:"silenceID"`
}

type ContextKey string

var clientKey ContextKey = "client"

func AttachClient(client kclient.Client) router.Middleware {
	return func(h router.Handler) router.Handler {
		return router.HandlerFunc(func(req router.Request, resp router.Response) error {
			req.Ctx = context.WithValue(req.Ctx, clientKey, client)

			return h.Handle(req, resp)
		})
	}
}

func SetDefaults(h router.Handler) router.Handler {
	return router.HandlerFunc(func(req router.Request, resp router.Response) error {
		silence := req.Object.(*v1.Silence)

		isChanged := false

		if silence.Spec.Comment == "" {
			silence.Spec.Comment = "(no comment) - created by alertmanager-controller"
			isChanged = true
		}
		if silence.Spec.CreatedBy == "" {
			silence.Spec.CreatedBy = "alertmanager-controller"
			isChanged = true
		}

		if isChanged {
			resp.Objects(silence)
			return nil
		}

		return h.Handle(req, resp)
	})
}

func SkipExpired(gcexpired bool, gcdelay time.Duration) router.Middleware {
	logrus.Trace("initialized")
	return func(h router.Handler) router.Handler {
		logrus.Trace("added middleware")
		return router.HandlerFunc(func(req router.Request, resp router.Response) error {
			logrus.Trace("middleware called")
			silence := req.Object.(*v1.Silence)

			if silence.Status.State == "expired" {
				if gcexpired && len(silence.GetOwnerReferences()) == 0 {
					logrus.Info("garbage collect expired")
					if silence.Spec.EndsAt.Add(gcdelay).Before(time.Now().UTC()) {
						client := req.Ctx.Value(clientKey).(kclient.Client)
						if err := client.Delete(req.Ctx, silence, &kclient.DeleteOptions{}); err != nil {
							return err
						}
					} else {
						// Set retry for GC
						resp.RetryAfter(gcdelay)
						resp.Objects(silence)
					}
				}

				logrus.Info("handle: expired")

				return nil
			}

			return h.Handle(req, resp)
		})
	}
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

func SetExpired(h router.Handler) router.Handler {
	return router.HandlerFunc(func(req router.Request, resp router.Response) error {
		silence := req.Object.(*v1.Silence)

		now := metav1.Now()
		if silence.Spec.EndsAt.Before(&now) && silence.Status.State != "expired" {
			silence.Status.State = "expired"
			resp.Objects(silence)
			return nil
		}

		return h.Handle(req, resp)
	})
}

func ManageSilence(req router.Request, resp router.Response) error {
	silence := req.Object.(*v1.Silence)
	cond := condition.Setter(silence, resp, "managed")

	amURL := strings.TrimSuffix(silence.Spec.URL, "/")

	if silence.Status.ID == "" {
		client := req.Ctx.Value(clientKey).(kclient.Client)

		newSilence := silence.Spec.DeepCopy()
		if newSilence.Comment == "" {
			newSilence.Comment = "automated silence"
		}
		if newSilence.CreatedBy == "" {
			newSilence.CreatedBy = "alertmanager-controller"
		}

		logrus.Debug("handle: no id, progressing")
		jsonData, err := json.Marshal(newSilence)
		if err != nil {
			return err
		}

		logrus.Debug("handle: creating silence")
		amResp, err := http.Post(fmt.Sprintf("%s/api/v2/silences", amURL), "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			silence.Status.State = "error-http"
			cond.Error(err)
			resp.Objects(silence)
			resp.RetryAfter(time.Second * 30)
			return nil
		}

		if amResp.StatusCode > 399 {
			errorContent, err := io.ReadAll(amResp.Body)
			if err != nil {
				return err
			}

			silence.Status.State = fmt.Sprintf("error-status-%d", amResp.StatusCode)

			cond.Error(errors.New(string(errorContent)))

			resp.RetryAfter(time.Second * 30)
			resp.Objects(silence)

			return nil
		}

		var silenceResp CreateResponse

		if err := json.NewDecoder(amResp.Body).Decode(&silenceResp); err != nil {
			return err
		}

		logrus.Info("handle: saving silence response")

		silence.Status.ID = silenceResp.SilenceID

		if err := client.Status().Update(req.Ctx, silence, &kclient.UpdateOptions{}); err != nil {
			return nil
		}

		resp.Objects(silence)

		return nil
	}

	logrus.Info("handle: querying existing silence")
	var silenceResp Response

	amResp, err := http.Get(fmt.Sprintf("%s/api/v2/silence/%s", amURL, silence.Status.ID))
	if err != nil {
		cond.Error(err)
		silence.Status.State = "error"
		resp.RetryAfter(time.Minute * 2)
		resp.Objects(silence)

		return nil
	}

	if amResp.StatusCode > 399 {
		errorContent, err := io.ReadAll(amResp.Body)
		if err != nil {
			return err
		}

		cond.Error(errors.New(string(errorContent)))
		silence.Status.State = "error"
		resp.RetryAfter(time.Minute * 2)
		resp.Objects(silence)

		return nil
	}

	if err := json.NewDecoder(amResp.Body).Decode(&silenceResp); err != nil {
		return err
	}

	silence.Status.State = silenceResp.Status.State

	resp.Objects(silence)

	logrus.Info("handle: saving silence state")

	cond.Success()

	now := time.Now().UTC()
	metaNow := metav1.NewTime(now)
	var retryAfterDuration = time.Minute * 1
	if silence.Status.State == "pending" {
		if !silence.Spec.StartsAt.Before(&metaNow) {
			retryAfterDuration = silence.Spec.StartsAt.Sub(now)
		}
	} else if silence.Status.State == "active" {
		retryAfterDuration = silence.Spec.EndsAt.Sub(now)
	}

	resp.RetryAfter(retryAfterDuration)

	return nil
}

func RemoveSilence(req router.Request, resp router.Response) error {
	silence := req.Object.(*v1.Silence)

	today := time.Now()
	tomorrow := silence.Spec.EndsAt.Time

	if tomorrow.After(today) {
		amURL := strings.TrimSuffix(silence.Spec.URL, "/")

		amReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v2/silence/%s", amURL, silence.Status.ID), nil)
		if err != nil {
			silence.Status.State = "error"
			resp.RetryAfter(time.Minute * 2)
			resp.Objects(silence)

			return err
		}

		amResp, err := http.DefaultClient.Do(amReq)
		if err != nil {
			silence.Status.State = "error"
			resp.RetryAfter(time.Minute * 2)
			resp.Objects(silence)

			return err
		}

		if amResp.StatusCode > 399 {
			silence.Status.State = "error"
			resp.RetryAfter(time.Minute * 2)
			resp.Objects(silence)

			return err
		}
	}

	return nil
}
