package runscope

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-providers/terraform-provider-runscope/internal/runscope/schema"
)

func StepRequestFromSchema(s *schema.StepRequest) *StepRequest {
	step := &StepRequest{
		ID: s.ID,
	}
	step.setFromSchema(s)

	return step
}

type StepRequest struct {
	ID            string
	StepType      string
	Method        string
	StepURL       string
	Variables     []StepVariable
	Assertions    []StepAssertion
	Headers       map[string][]string
	Auth          StepAuth
	Body          string
	Form          map[string][]string
	Scripts       []string
	BeforeScripts []string
	Note          string
	Skipped       bool
}

func (sb *StepRequest) setFromSchema(s *schema.StepRequest) {
	sb.ID = s.ID
	sb.StepType = s.StepType
	sb.Method = s.Method
	sb.StepURL = s.URL
	sb.Variables = make([]StepVariable, len(s.Variables))
	sb.Assertions = make([]StepAssertion, len(s.Assertions))
	sb.Headers = map[string][]string{}
	sb.Auth = StepAuth{
		Username: s.Auth.Username,
		Password: s.Auth.Password,
		AuthType: s.Auth.AuthType,
	}
	sb.Body = s.Body
	sb.Form = map[string][]string{}
	sb.Scripts = make([]string, len(s.Scripts))
	sb.BeforeScripts = make([]string, len(s.BeforeScripts))
	sb.Note = s.Note
	sb.Skipped = s.Skipped

	for i, v := range s.Variables {
		sb.Variables[i] = StepVariable{
			Name:     v.Name,
			Property: v.Property,
			Source:   v.Source,
		}
	}
	for i, a := range s.Assertions {
		sb.Assertions[i] = StepAssertion{
			Source:     a.Source,
			Property:   a.Property,
			Comparison: a.Comparison,
			Value:      a.Value.String(),
		}
	}
	for header, values := range s.Headers {
		sb.Headers[header] = make([]string, len(values))
		copy(sb.Headers[header], values)
	}
	for name, values := range s.Form {
		sb.Form[name] = make([]string, len(values))
		copy(sb.Form[name], values)
	}
	copy(sb.Scripts, s.Scripts)
	copy(sb.BeforeScripts, s.BeforeScripts)
}

type StepRequestOpts struct {
	Method        string
	StepURL       string
	Variables     []StepVariable
	Assertions    []StepAssertion
	Headers       map[string][]string
	Auth          StepAuth
	Body          string
	Form          map[string][]string
	Scripts       []string
	BeforeScripts []string
	Note          string
	Skipped       bool
}

func (sbo *StepRequestOpts) setRequest(sb *schema.StepRequest) {
	sb.Method = sbo.Method
	sb.URL = sbo.StepURL
	sb.Variables = make([]schema.StepVariable, len(sbo.Variables))
	sb.Assertions = make([]schema.StepAssertion, len(sbo.Assertions))
	sb.Headers = map[string][]string{}
	sb.Auth = schema.StepAuth{
		Username: sbo.Auth.Username,
		Password: sbo.Auth.Password,
		AuthType: sbo.Auth.AuthType,
	}
	sb.Body = sbo.Body
	sb.Form = map[string][]string{}
	sb.Scripts = make([]string, len(sbo.Scripts))
	sb.BeforeScripts = make([]string, len(sbo.BeforeScripts))
	sb.Note = sbo.Note
	sb.Skipped = sbo.Skipped

	for i, v := range sbo.Variables {
		sb.Variables[i] = schema.StepVariable{
			Name:     v.Name,
			Property: v.Property,
			Source:   v.Source,
		}
	}
	for i, a := range sbo.Assertions {
		sb.Assertions[i] = schema.StepAssertion{
			Source:     a.Source,
			Comparison: a.Comparison,
			Value:      schema.FlexString(a.Value),
			Property:   a.Property,
		}
	}
	for header, values := range sbo.Headers {
		sb.Headers[header] = make([]string, len(values))
		copy(sb.Headers[header], values)
	}
	for name, values := range sbo.Form {
		sb.Form[name] = make([]string, len(values))
		copy(sb.Form[name], values)
	}
	copy(sb.Scripts, sbo.Scripts)
	copy(sb.BeforeScripts, sbo.BeforeScripts)
}

type StepCreateRequestOpts struct {
	StepUriOpts
	StepRequestOpts
}

func (c *StepClient) CreateRequest(ctx context.Context, opts *StepCreateRequestOpts) (*StepRequest, error) {
	body := &schema.StepCreateRequestRequest{
		StepType: "request",
	}
	opts.StepRequestOpts.setRequest(&body.StepRequest)

	req, err := c.client.NewRequest(ctx, http.MethodPost, opts.StepUriOpts.URL(), &body)
	if err != nil {
		return nil, err
	}

	var resp schema.StepCreateRequestResponse
	err = c.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}

	return StepRequestFromSchema(&resp.Step[len(resp.Step)-1]), nil
}

type StepGetRequestOpts struct {
	StepUriOpts
	Id string
}

func (opts *StepGetRequestOpts) URL() string {
	return fmt.Sprintf("%s/%s", opts.StepUriOpts.URL(), opts.Id)
}

func (c *StepClient) GetRequest(ctx context.Context, opts *StepGetRequestOpts) (*StepRequest, error) {
	var resp schema.StepGetRequestResponse
	req, err := c.client.NewRequest(ctx, http.MethodGet, opts.URL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create req: %w", err)
	}

	err = c.client.Do(req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to call endpoint: %w", err)
	}

	return StepRequestFromSchema(&resp.Step), nil
}

type StepUpdateRequestOpts struct {
	StepGetRequestOpts
	StepRequestOpts
}

func (c *StepClient) UpdateRequest(ctx context.Context, opts *StepUpdateRequestOpts) (*StepRequest, error) {
	body := &schema.StepUpdateRequestRequest{
		StepType: "request",
	}
	opts.setRequest(&body.StepRequest)
	body.ID = opts.Id

	tflog.Info(ctx, "calling with PUT", map[string]interface{}{
		"url":  opts.URL(),
		"body": body,
	})
	req, err := c.client.NewRequest(ctx, http.MethodPut, opts.URL(), &body)
	if err != nil {
		return nil, err
	}

	var resp schema.StepUpdateRequstResponse
	err = c.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}

	return StepRequestFromSchema(&resp.Step), nil
}
