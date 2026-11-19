package runscope

import (
	"context"
	"fmt"
	"net/http"

	"github.com/terraform-providers/terraform-provider-runscope/internal/runscope/schema"
)

func StepSubtestFromSchema(s *schema.StepSubtest) *StepSubtest {
	step := StepSubtest{
		ID: s.ID,
	}

	step.setFromSchema(s)
	return &step
}

type StepSubtest struct {
	ID                   string
	BucketKey            string
	TestUUID             string
	EnvironmentUUID      string
	UseParentEnvironment bool
	Variables            []StepVariable
	Assertions           []StepAssertion
}

func (ss *StepSubtest) setFromSchema(s *schema.StepSubtest) {
	ss.BucketKey = s.BucketKey
	ss.EnvironmentUUID = s.EnvironmentUUID
	ss.TestUUID = s.TestUUID
	ss.UseParentEnvironment = s.UseParentEnvironment

	ss.Variables = make([]StepVariable, len(s.Variables))
	ss.Assertions = make([]StepAssertion, len(s.Assertions))
	for i, v := range s.Variables {
		ss.Variables[i] = StepVariable{
			Name:     v.Name,
			Property: v.Property,
			Source:   v.Source,
		}
	}
	for i, a := range s.Assertions {
		ss.Assertions[i] = StepAssertion{
			Source:     a.Source,
			Property:   a.Property,
			Comparison: a.Comparison,
			Value:      a.Value.String(),
		}
	}
}

type StepSubtestOpts struct {
	TestUUID             string
	EnvironmentUUID      string
	BucketKey            string
	UseParentEnvironment bool
	Assertions           []StepAssertion
	Variables            []StepVariable
}

func (sso *StepSubtestOpts) setRequest(ss *schema.StepSubtest) {
	ss.BucketKey = sso.BucketKey
	ss.TestUUID = sso.TestUUID
	ss.EnvironmentUUID = sso.EnvironmentUUID
	ss.UseParentEnvironment = sso.UseParentEnvironment

	ss.Variables = make([]schema.StepVariable, len(sso.Variables))
	ss.Assertions = make([]schema.StepAssertion, len(sso.Assertions))
	for i, v := range sso.Variables {
		ss.Variables[i] = schema.StepVariable{
			Name:     v.Name,
			Property: v.Property,
			Source:   v.Source,
		}
	}
	for i, a := range sso.Assertions {
		ss.Assertions[i] = schema.StepAssertion{
			Source:     a.Source,
			Comparison: a.Comparison,
			Value:      schema.FlexString(a.Value),
			Property:   a.Property,
		}
	}
}

type StepCreateSubtestOpts struct {
	StepUriOpts
	StepSubtestOpts
}

func (c *StepClient) CreateSubtest(ctx context.Context, opts *StepCreateSubtestOpts) (*StepSubtest, error) {
	body := schema.StepCreateSubtestRequest{
		StepType: "subtest",
	}
	opts.StepSubtestOpts.setRequest(&body.StepSubtest)
	req, err := c.client.NewRequest(ctx, http.MethodPost, opts.StepUriOpts.URL(), &body)
	if err != nil {
		return nil, err
	}

	var resp schema.StepCreateSubtestResponse
	if err := c.client.Do(req, &resp); err != nil {
		return nil, err
	}

	if len(resp.Step) < 1 {
		return nil, fmt.Errorf("no steps returned after created")
	}

	return StepSubtestFromSchema(&resp.Step[len(resp.Step)-1]), nil
}

func (c *StepClient) GetSubtest(ctx context.Context, opts *StepGetRequestOpts) (*StepSubtest, error) {
	var resp schema.StepGetSubstepResponse
	req, err := c.client.NewRequest(ctx, http.MethodGet, opts.URL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create req: %w", err)
	}

	err = c.client.Do(req, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to call endpoint: %w", err)
	}

	return StepSubtestFromSchema(&resp.Step), nil
}

type StepUpdateSubtestOpts struct {
	StepGetRequestOpts
	StepSubtestOpts
}

func (c *StepClient) UpdateSubtest(ctx context.Context, opts *StepUpdateSubtestOpts) (*StepSubtest, error) {
	body := &schema.StepUpdateSubtestRequest{
		StepType: "subtest",
	}
	opts.setRequest(&body.StepSubtest)
	body.ID = opts.Id

	req, err := c.client.NewRequest(ctx, http.MethodPut, opts.URL(), &body)
	if err != nil {
		return nil, err
	}

	var resp schema.StepUpdateSubtestResponse
	err = c.client.Do(req, &resp)
	if err != nil {
		return nil, err
	}

	return StepSubtestFromSchema(&resp.Step), nil
}
