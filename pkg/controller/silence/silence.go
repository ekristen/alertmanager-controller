package silence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	v1 "github.com/ekristen/prom-am-operator/pkg/apis/promam.ekristen.dev/v1"
	"github.com/ibuildthecloud/baaah/pkg/router"
)

type SilenceResponse struct {
	v1.SilenceSpec `json:",inline"`
	ID             string           `json:"id"`
	Status         v1.SilenceStatus `json:"status"`
}

type SilenceCreateResponse struct {
	SilenceID string `json:"silenceID"`
}

func ManageSilence(req router.Request, resp router.Response) error {
	silence := req.Object.(*v1.Silence)

	if silence.Spec.URL == "" {
		return errors.New("alertmanager url is missing")
	}

	amURL := strings.TrimSuffix(silence.Spec.URL, "/")

	if silence.Status.State == "expired" {
		return nil
	}

	if silence.Status.ID == "" {
		jsonData, err := json.Marshal(silence.Spec)
		if err != nil {
			return err
		}

		amResp, err := http.Post(fmt.Sprintf("%s/api/v2/silences", amURL), "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		if amResp.StatusCode > 399 {
			errorContent, err := ioutil.ReadAll(amResp.Body)
			if err != nil {
				return err
			}

			return errors.New(string(errorContent))
		}

		var silenceResp SilenceCreateResponse

		if err := json.NewDecoder(amResp.Body).Decode(&silenceResp); err != nil {
			return err
		}

		silence.Status.ID = silenceResp.SilenceID

		resp.Objects(silence)
		resp.RetryAfter(1 * time.Second)

		return nil
	}

	var silenceResp SilenceResponse

	amResp, err := http.Get(fmt.Sprintf("%s/api/v2/silence/%s", amURL, silence.Status.ID))
	if err != nil {
		return err
	}

	if amResp.StatusCode > 399 {
		errorContent, err := ioutil.ReadAll(amResp.Body)
		if err != nil {
			return err
		}

		return errors.New(string(errorContent))
	}

	if err := json.NewDecoder(amResp.Body).Decode(&silenceResp); err != nil {
		return err
	}

	silence.Status.State = silenceResp.Status.State

	resp.Objects(silence)

	if silence.Status.State == "pending" || silence.Status.State == "active" {
		resp.RetryAfter(time.Minute * 5)
	}

	return nil
}
