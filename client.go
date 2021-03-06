package auctioneer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cfhttp "code.cloudfoundry.org/cfhttp/v2"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/tlsconfig"
	"github.com/tedsuo/rata"
)

//go:generate counterfeiter -o auctioneerfakes/fake_client.go . Client
type Client interface {
	RequestLRPAuctions(logger lager.Logger, lrpStart []*LRPStartRequest) error
	RequestTaskAuctions(logger lager.Logger, tasks []*TaskStartRequest) error
}

type auctioneerClient struct {
	httpClient         *http.Client
	insecureHTTPClient *http.Client
	url                string
	requireTLS         bool
}

func NewClient(auctioneerURL string, requestTimeout time.Duration) Client {
	return &auctioneerClient{
		httpClient: cfhttp.NewClient(
			cfhttp.WithRequestTimeout(requestTimeout),
		),
		url: auctioneerURL,
	}
}

func NewSecureClient(auctioneerURL, caFile, certFile, keyFile string, requireTLS bool, requestTimeout time.Duration) (Client, error) {
	insecureHTTPClient := cfhttp.NewClient(
		cfhttp.WithRequestTimeout(requestTimeout),
	)

	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithInternalServiceDefaults(),
		tlsconfig.WithIdentityFromFile(certFile, keyFile),
	).Client(tlsconfig.WithAuthorityFromFile(caFile))
	if err != nil {
		return nil, err
	}

	httpClient := cfhttp.NewClient(
		cfhttp.WithRequestTimeout(requestTimeout),
		cfhttp.WithTLSConfig(tlsConfig),
	)

	return &auctioneerClient{
		httpClient:         httpClient,
		insecureHTTPClient: insecureHTTPClient,
		url:                auctioneerURL,
		requireTLS:         requireTLS,
	}, nil
}

func (c *auctioneerClient) RequestLRPAuctions(logger lager.Logger, lrpStarts []*LRPStartRequest) error {
	logger = logger.Session("request-lrp-auctions")

	reqGen := rata.NewRequestGenerator(c.url, Routes)
	payload, err := json.Marshal(lrpStarts)
	if err != nil {
		return err
	}

	req, err := reqGen.CreateRequest(CreateLRPAuctionsRoute, rata.Params{}, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(logger, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("http error: status code %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

func (c *auctioneerClient) RequestTaskAuctions(logger lager.Logger, tasks []*TaskStartRequest) error {
	logger = logger.Session("request-task-auctions")

	reqGen := rata.NewRequestGenerator(c.url, Routes)
	payload, err := json.Marshal(tasks)
	if err != nil {
		return err
	}

	req, err := reqGen.CreateRequest(CreateTaskAuctionsRoute, rata.Params{}, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(logger, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("http error: status code %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}

func (c *auctioneerClient) doRequest(logger lager.Logger, req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Fall back to HTTP and try again if we do not require TLS
		if !c.requireTLS && c.insecureHTTPClient != nil {
			logger.Error("retrying-on-http", err)
			req.URL.Scheme = "http"
			return c.insecureHTTPClient.Do(req)
		}
	}
	return resp, err
}
