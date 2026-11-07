package googlecontactsfeed

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/tekkamanendless/httperror"
)

// Client is a client for the legacy Google Contacts API.
type Client struct {
	config ClientConfig
}

// NewClient creates a new Client.
func NewClient(opts ...ClientOption) *Client {
	config := ClientConfig{
		HTTPClient: &http.Client{},
	}
	for _, opt := range opts {
		opt(&config)
	}
	return &Client{config: config}
}

// ClientConfig is the configuration for the Client.
type ClientConfig struct {
	Domain      string
	HTTPClient  *http.Client
	AccessToken string
}

// ClientOption is a function that configures the Client.
type ClientOption func(*ClientConfig)

// WithAccessToken sets the access token for the Client.
func WithAccessToken(accessToken string) ClientOption {
	return func(c *ClientConfig) {
		c.AccessToken = accessToken
	}
}

// WithDomain sets the domain for the Client.
func WithDomain(domain string) ClientOption {
	return func(c *ClientConfig) {
		c.Domain = domain
	}
}

// WithHTTPClient sets the HTTP client for the Client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *ClientConfig) {
		c.HTTPClient = httpClient
	}
}

// GetContactsOptions is the options for the GetContacts method.
//
// See: https://developers.google.com/workspace/admin/domain-shared-contacts/contacts-query-parameters
type GetContactsOptions struct {
	MaxResults  uint      `query:"max-results"`
	StartIndex  uint      `query:"start-index"`
	UpdatedMin  time.Time `query:"updated-min"`
	OrderBy     string    `query:"orderby"`
	ShowDeleted bool      `query:"showdeleted"`
	SortOrder   string    `query:"sortorder"`
}

// ErrorResponse is the error response from the Google Contacts API.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// GetContacts gets the contacts from the Google Contacts API.
//
// This handles pagination automatically.
func (c *Client) GetContacts(ctx context.Context, GetContactsOptions GetContactsOptions) ([]GoogleContact, error) {
	var contacts []GoogleContact

	values := url.Values{}
	if GetContactsOptions.MaxResults > 0 {
		values.Add("max-results", strconv.Itoa(int(GetContactsOptions.MaxResults)))
	}
	if GetContactsOptions.StartIndex > 0 {
		values.Add("start-index", strconv.Itoa(int(GetContactsOptions.StartIndex)))
	}
	if !GetContactsOptions.UpdatedMin.IsZero() {
		values.Add("updated-min", GetContactsOptions.UpdatedMin.Format(time.RFC3339))
	}
	if GetContactsOptions.OrderBy != "" {
		values.Add("orderby", GetContactsOptions.OrderBy)
	}
	if GetContactsOptions.ShowDeleted {
		values.Add("showdeleted", "true")
	}
	if GetContactsOptions.SortOrder != "" {
		values.Add("sortorder", GetContactsOptions.SortOrder)
	}

	nextPageURL := "https://www.google.com/m8/feeds/contacts/" + c.config.Domain + "/full"
	if len(values) > 0 {
		nextPageURL += "?" + values.Encode()
	}
	for page := 0; nextPageURL != ""; page++ {
		err := func() error {
			slog.DebugContext(ctx, fmt.Sprintf("Getting Google contacts: %s", nextPageURL))
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, nextPageURL, nil)
			if err != nil {
				return fmt.Errorf("could not create request: %w", err)
			}

			request.Header.Set("GData-Version", "3.0")
			if c.config.AccessToken != "" {
				request.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
			}

			response, err := c.config.HTTPClient.Do(request)
			if err != nil {
				return fmt.Errorf("could not do request: %w", err)
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				return fmt.Errorf("could not read body: %w", err)
			}

			slog.DebugContext(ctx, fmt.Sprintf("Body: %s", string(body)))

			if response.StatusCode >= 400 {
				var errorResponse ErrorResponse
				err = json.Unmarshal(body, &errorResponse)
				if err != nil {
					return httperror.ErrorFromStatus(response.StatusCode)
				}
				return fmt.Errorf("%w: %s: %s", httperror.ErrorFromStatus(response.StatusCode), errorResponse.Status, errorResponse.Message)
			}

			var feed ContactFeed
			err = xml.Unmarshal(body, &feed)
			if err != nil {
				return fmt.Errorf("could not unmarshal feed: %w", err)
			}
			slog.DebugContext(ctx, fmt.Sprintf("Feed: %+v", feed))
			//spew.Dump(feed)

			contacts = append(contacts, feed.Entries...)

			nextPageURL = ""
			for _, link := range feed.Link {
				if link.Rel == "next" {
					nextPageURL = link.Href
					break
				}
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return contacts, nil
}
