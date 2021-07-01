package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type whatsonchain struct{
	cli *http.Client
}

func NewWhatsOnChain(cli *http.Client	) *whatsonchain{
	return &whatsonchain{cli: cli}
}

func (w *whatsonchain) Height(ctx context.Context) (int, error){
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.whatsonchain.com/v1/bsv/main/chain/info",nil)
	if err != nil{
		return 0, errors.Wrap(err, "failed to build request when sending Header request to woc")
	}
	resp, err := w.cli.Do(req)
	if err != nil{
		return 0, errors.Wrap(err, "failed to send request when sending Header request to woc")
	}
	defer resp.Body.Close() // nolint
	if resp.StatusCode != http.StatusOK{
		body, _ := ioutil.ReadAll(resp.Body)
		return 0, errors.Wrapf(err,"unexpected error received from whatsonchain \n statuscode: %d \n body:%s", resp.StatusCode, body )
	}
	blockResp := struct{
		Headers int `json:"headers"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&blockResp); err != nil{
		return 0, errors.Wrapf(err, "failed to decode woc response")
	}
	return blockResp.Headers, nil
}
