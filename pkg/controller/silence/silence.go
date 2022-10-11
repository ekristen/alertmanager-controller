package silence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/ekristen/alertmanager-controller/pkg/apis/alertmanager.ekristen.dev/v1"
	"github.com/ekristen/alertmanager-controller/pkg/condition"
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

	amURL := strings.TrimSuffix(silence.Spec.URL, "/")

	if silence.Status.ID == "" {
		client := req.Ctx.Value(clientKey).(kclient.Client)
		s2 := &v1.Silence{}
		if err := client.Get(req.Ctx, kclient.ObjectKey{Namespace: silence.GetNamespace(), Name: silence.GetName()}, s2); err != nil {
			return err
		}
		if s2.Status.ID == "" {
			logrus.Info("handle: no id, progressing")
			jsonData, err := json.Marshal(silence.Spec)
			if err != nil {
				return err
			}

			logrus.Info("handle: creating silence")
			amResp, err := http.Post(fmt.Sprintf("%s/api/v2/silences", amURL), "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				silence.Status.State = "error-http"
				cond.Error(err)
				resp.Objects(silence)
				resp.RetryAfter(time.Minute * 2)
				return nil
			}

			if amResp.StatusCode > 399 {
				errorContent, err := ioutil.ReadAll(amResp.Body)
				if err != nil {
					return err
				}

				silence.Status.State = fmt.Sprintf("error-status-%d", amResp.StatusCode)

				cond.Error(errors.New(string(errorContent)))

				resp.RetryAfter(time.Minute * 2)
				resp.Objects(silence)

				return nil
			}

			var silenceResp SilenceCreateResponse

			if err := json.NewDecoder(amResp.Body).Decode(&silenceResp); err != nil {
				return err
			}

			logrus.Info("handle: saving silence response")

			silence.Status.ID = silenceResp.SilenceID

			fmt.Println(silence.Status)
			resp.Objects(silence)

			return nil
		}
	}

	logrus.Info("handle: querying existing silence")
	var silenceResp SilenceResponse

	amResp, err := http.Get(fmt.Sprintf("%s/api/v2/silence/%s", amURL, silence.Status.ID))
	if err != nil {
		cond.Error(err)
		silence.Status.State = "error"
		resp.RetryAfter(time.Minute * 2)
		resp.Objects(silence)

		return nil
	}

	if amResp.StatusCode > 399 {
		errorContent, err := ioutil.ReadAll(amResp.Body)
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

	if silence.Status.State == "pending" || silence.Status.State == "active" {
		resp.RetryAfter(time.Minute * 5)
	}

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

			return nil
		}

		amResp, err := http.DefaultClient.Do(amReq)
		if err != nil {
			silence.Status.State = "error"
			resp.RetryAfter(time.Minute * 2)
			resp.Objects(silence)

			return nil
		}

		if amResp.StatusCode > 399 {
			silence.Status.State = "error"
			resp.RetryAfter(time.Minute * 2)
			resp.Objects(silence)

			return nil
		}
	}

	return nil
}
