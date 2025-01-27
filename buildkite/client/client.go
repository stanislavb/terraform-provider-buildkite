package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	defaultBaseURL             = "https://api.buildkite.com/"
	defaultGraphQLUrl          = "https://graphql.buildkite.com/v1"
	applicationJsonContentType = "application/json"
)

type Client struct {
	client   *http.Client
	graphQl  *graphql.Client
	baseURL  *url.URL
	orgSlug  string
	apiToken string
}

func NewClient(orgSlug string, apiToken string, userAgent string) *Client {
	var authTransport http.RoundTripper = NewAuthTransport(apiToken, userAgent, nil)
	baseURL, _ := url.Parse(defaultBaseURL)

	graphQlClient := graphql.NewClient(defaultGraphQLUrl, graphql.WithHTTPClient(&http.Client{
		Transport: authTransport,
	}))

	graphQlClient.Log = func(responseBody string) {
		log.Printf("[TRACE] Response body:\n%s", responseBody)
	}

	return &Client{
		client: &http.Client{
			Transport: authTransport,
		},
		graphQl:  graphQlClient,
		baseURL:  baseURL,
		orgSlug:  orgSlug,
		apiToken: apiToken,
	}
}

func (c *Client) graphQLRequest(req *graphql.Request, result interface{}) error {
	jsonBytes, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("[TRACE] GraphQL request %s", string(jsonBytes))

	err := c.graphQl.Run(context.Background(), req, &result)
	if err != nil {
		log.Printf("[TRACE] GraphQL error %v", err)
		return err
	}

	jsonBytes, _ = json.MarshalIndent(result, "", "  ")
	log.Printf("[TRACE] GraphQL response %s", string(jsonBytes))
	return nil
}

func (c *Client) createOrgSlug(slug string) string {
	return fmt.Sprintf("%s/%s", c.orgSlug, slug)
}

func marshalBody(body interface{}) (*bytes.Buffer, error) {
	if body == nil {
		return nil, nil
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal body")
	}
	log.Printf("[TRACE] Buildkite Request body %s\n", string(bodyBytes))

	return bytes.NewBuffer(bodyBytes), nil
}

func unmarshalResponse(body io.Reader, result interface{}) error {
	responseBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return errors.Wrap(err, "could not read response body")
	}
	log.Printf("[TRACE] Buildkite Response body %s\n", string(responseBytes))

	err = json.Unmarshal(responseBytes, result)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal response body")
	}

	return nil
}
