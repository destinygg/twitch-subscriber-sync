package api

import (
	"io"
	"net/http"
	"io/ioutil"
	"time"
	"crypto/tls"
	"github.com/destinygg/twitch-subscriber-sync/internal/debug"
	"github.com/destinygg/twitch-subscriber-sync/internal/config"
	"golang.org/x/net/context"
)

type Api struct {
	cfg *config.AppConfig
	client http.Client
}

func Init(ctx context.Context) context.Context {
	return context.WithValue(ctx, "dggapi", &Api{
		cfg: config.FromContext(ctx),
		client: http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{},
				ResponseHeaderTimeout: 5 * time.Second,
			},
		},
	})
}

func FromContext(ctx context.Context) *Api {
	cfg, _ := ctx.Value("dggapi").(*Api)
	return cfg
}

func (a *Api) SendSubDataToApi(body io.Reader) error {
	_, err := a.call("POST", a.cfg.SubURL, body)
	return err
}

func (a *Api) call(method, url string, body io.Reader) ([]byte, error) {
	u := url + "?privatekey=" + a.cfg.Website.PrivateAPIKey
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		d.PF(2, "Could not create request: %#v", err)
		return nil, err
	}

	res, err := a.client.Do(req)
	if res == nil || res.Body == nil {
		return nil, nil
	}
	defer res.Body.Close()

	if err != nil || res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := ioutil.ReadAll(res.Body)
		d.PF(2, "Request failed: %#v, body was \n%v", err, string(data))
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		d.PF(2, "Could not read body: %#v", err)
		return nil, err
	}

	return data, nil
}