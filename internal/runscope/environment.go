package runscope

import (
	"context"
	"fmt"
	"github.com/terraform-providers/terraform-provider-runscope/internal/runscope/schema"
)

type EnvironmentBase struct {
	Name             string
	Script           string
	PreserveCookies  bool
	InitialVariables map[string]string
	Integrations     []string
	Regions          []string
	RemoteAgents     []RemoteAgent
	RetryOnFailure   bool
	VerifySSL        bool
	Webhooks         []string
	Emails           Emails
}

type Environment struct {
	EnvironmentBase
	Id string
}

type RemoteAgent struct {
	Name string
	UUID string
}

type Emails struct {
	NotifyAll       bool
	NotifyOn        string
	NotifyThreshold int
	Recipients      []Recipient
}

func (e Emails) IsDefault() bool {
	return !e.NotifyAll && e.NotifyOn == "" && e.NotifyThreshold == 0 && len(e.Recipients) == 0
}

type Recipient struct {
	Id    string
	Name  string
	Email string
}

type EnvironmentClient struct {
	client *Client
}

func EnvironmentFromSchema(s *schema.Environment) *Environment {
	env := &Environment{}
	env.Id = s.Id
	env.Name = s.Name
	env.Script = s.Script
	env.PreserveCookies = s.PreserveCookies
	env.InitialVariables = s.InitialVariables
	env.Regions = s.Regions
	env.RetryOnFailure = s.RetryOnFailure
	env.VerifySSL = s.VerifySSL
	env.Webhooks = s.Webhooks
	env.Emails = Emails{
		NotifyAll:       s.Emails.NotifyAll,
		NotifyOn:        s.Emails.NotifyOn,
		NotifyThreshold: s.Emails.NotifyThreshold,
	}

	for _, i := range s.Integrations {
		env.Integrations = append(env.Integrations, i.Id)
	}
	for _, ra := range s.RemoteAgents {
		env.RemoteAgents = append(env.RemoteAgents, RemoteAgent{
			Name: ra.Name,
			UUID: ra.UUID,
		})
	}
	for _, r := range s.Emails.Recipients {
		env.Emails.Recipients = append(env.Emails.Recipients, Recipient{
			Id:    r.Id,
			Name:  r.Name,
			Email: r.Email,
		})
	}
	return env
}

type EnvironmentUriOpts struct {
	BucketId string
	TestId   string
}

func (opts *EnvironmentUriOpts) BaseURL() string {
	if opts.TestId == "" {
		return fmt.Sprintf("/buckets/%s/environments", opts.BucketId)
	}
	return fmt.Sprintf("/buckets/%s/test/%s/environments", opts.BucketId, opts.TestId)
}

type EnvironmentCreateOpts struct {
	EnvironmentUriOpts
	EnvironmentBase
}

func (c *EnvironmentClient) Create(ctx context.Context, opts *EnvironmentCreateOpts) (*Environment, error) {
	body := &schema.EnvironmentCreateRequest{}
	body.Name = opts.Name
	body.Script = opts.Script
	body.PreserveCookies = opts.PreserveCookies
	body.InitialVariables = opts.InitialVariables
	body.Regions = opts.Regions
	body.RetryOnFailure = opts.RetryOnFailure
	body.VerifySSL = opts.VerifySSL
	body.Webhooks = opts.Webhooks
	body.Emails = schema.Emails{
		NotifyAll:       opts.Emails.NotifyAll,
		NotifyOn:        opts.Emails.NotifyOn,
		NotifyThreshold: opts.Emails.NotifyThreshold,
	}
	for _, id := range opts.Integrations {
		body.Integrations = append(body.Integrations, schema.EnvironmentIntegration{Id: id})
	}
	for _, agent := range opts.RemoteAgents {
		body.RemoteAgents = append(body.RemoteAgents, schema.RemoteAgent{
			Name: agent.Name,
			UUID: agent.UUID,
		})
	}
	for _, recipient := range opts.Emails.Recipients {
		body.Emails.Recipients = append(body.Emails.Recipients, schema.Recipient{
			Id:    recipient.Id,
			Name:  recipient.Name,
			Email: recipient.Email,
		})
	}

	req, err := c.client.NewRequest(ctx, "POST", opts.BaseURL(), &body)
	if err != nil {
		return nil, err
	}

	var resp schema.EnvironmentCreateResponse
	err = c.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}

	return EnvironmentFromSchema(&resp.Environment), err
}

type EnvironmentGetOpts struct {
	EnvironmentUriOpts
	Id string
}

func (opts *EnvironmentGetOpts) URL() string {
	return fmt.Sprintf("%s/%s", opts.BaseURL(), opts.Id)
}

func (c *EnvironmentClient) Get(ctx context.Context, opts *EnvironmentGetOpts) (*Environment, error) {
	var resp schema.EnvironmentGetResponse
	req, err := c.client.NewRequest(ctx, "GET", opts.URL(), nil)
	if err != nil {
		return nil, err
	}

	err = c.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}

	return EnvironmentFromSchema(&resp.Environment), err
}

type EnvironmentUpdateOpts struct {
	EnvironmentGetOpts
	EnvironmentBase
}

func (c *EnvironmentClient) Update(ctx context.Context, opts *EnvironmentUpdateOpts) (*Environment, error) {
	body := &schema.EnvironmentUpdateRequest{}
	body.Name = opts.Name
	body.Script = opts.Script
	body.PreserveCookies = opts.PreserveCookies
	body.InitialVariables = opts.InitialVariables
	body.Regions = opts.Regions
	body.RetryOnFailure = opts.RetryOnFailure
	body.VerifySSL = opts.VerifySSL
	body.Webhooks = opts.Webhooks
	body.Emails = schema.Emails{
		NotifyAll:       opts.Emails.NotifyAll,
		NotifyOn:        opts.Emails.NotifyOn,
		NotifyThreshold: opts.Emails.NotifyThreshold,
	}
	for _, id := range opts.Integrations {
		body.Integrations = append(body.Integrations, schema.EnvironmentIntegration{Id: id})
	}
	for _, agent := range opts.RemoteAgents {
		body.RemoteAgents = append(body.RemoteAgents, schema.RemoteAgent{
			Name: agent.Name,
			UUID: agent.UUID,
		})
	}
	for _, recipient := range opts.Emails.Recipients {
		body.Emails.Recipients = append(body.Emails.Recipients, schema.Recipient{
			Id:    recipient.Id,
			Name:  recipient.Name,
			Email: recipient.Email,
		})
	}
	req, err := c.client.NewRequest(ctx, "PUT", opts.URL(), &body)
	if err != nil {
		return nil, err
	}

	var resp schema.EnvironmentUpdateResponse
	err = c.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}

	return EnvironmentFromSchema(&resp.Environment), err
}

type EnvironmentDeleteOpts struct {
	EnvironmentGetOpts
}

func (c *EnvironmentClient) Delete(ctx context.Context, opts *EnvironmentDeleteOpts) error {
	req, err := c.client.NewRequest(ctx, "DELETE", opts.URL(), nil)
	if err != nil {
		return err
	}

	err = c.client.Do(req, nil)
	if err != nil {
		return err
	}

	return nil
}
