package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-runscope/internal/runscope"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRunscopeStepRequest() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStepRequestCreate,
		ReadContext:   resourceStepRequestRead,
		UpdateContext: resourceStepRequestUpdate,
		DeleteContext: resourceStepDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.Split(d.Id(), "/")

				bucketId := parts[0]
				d.Set("bucket_id", bucketId)

				if len(parts) == 3 {
					d.Set("test_id", parts[1])
					d.SetId(parts[2])
					return []*schema.ResourceData{d}, nil
				}

				if len(parts) != 2 {
					return nil, fmt.Errorf("step ID for import should be in format bucket_id/test_id/step_id " +
						"or bucket_id/test_id#step_position")
				}

				parts = strings.Split(parts[1], "#")
				if len(parts) != 2 {
					return nil, fmt.Errorf("step ID for import should be in format bucket_id/test_id/step_id " +
						"or bucket_id/test_id#step_position")
				}

				stepPos, err := strconv.Atoi(parts[1])
				if err != nil || stepPos < 1 {
					return nil, fmt.Errorf("step_position should be a positive integer number")
				}

				testId := parts[0]
				d.Set("test_id", testId)

				opts := runscope.TestGetOpts{
					BucketId: bucketId,
					Id:       testId,
				}

				client := meta.(*providerConfig).client

				test, err := client.Test.Get(ctx, opts)
				if err != nil {
					return nil, fmt.Errorf("couldn't read test: %s", err)
				}

				nSteps := len(test.Steps)
				if nSteps < stepPos {
					return nil, fmt.Errorf("test %s contains only %d steps", testId, nSteps)
				}

				d.SetId(test.Steps[stepPos-1].Id)

				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"test_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"method": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"variable": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"property": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"source": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(stepSources, false),
						},
					},
				},
				Optional: true,
			},
			"assertion": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(stepSources, false),
						},
						"property": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"comparison": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(stepComparisons, false),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Optional: true,
			},
			"header": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"header": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"auth": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
						"auth_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"body": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"form_parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"scripts": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"before_scripts": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"note": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"skipped": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceStepRequestCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerConfig).client

	opts := &runscope.StepCreateRequestOpts{}
	expandStepUriOpts(d, &opts.StepUriOpts)
	expandStepBaseOpts(d, &opts.StepRequestOpts)

	step, err := client.Step.CreateRequest(ctx, opts)
	if err != nil {
		return diag.Errorf("Couldn't create step: %s", err)
	}

	d.SetId(step.ID)

	return resourceStepRequestRead(ctx, d, meta)
}

func resourceStepRequestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerConfig).client

	opts := &runscope.StepGetRequestOpts{}
	opts.Id = d.Id()
	opts.TestId = d.Get("test_id").(string)
	opts.BucketId = d.Get("bucket_id").(string)

	step, err := client.Step.GetRequest(ctx, opts)
	if err != nil {
		if isNotFound(err) {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Couldn't read step: %s", err)
	}

	d.Set("method", step.Method)
	d.Set("url", step.StepURL)
	d.Set("variable", flattenStepVariables(step.Variables))
	d.Set("assertion", flattenStepAssertions(step.Assertions))
	d.Set("header", flattenStepHeaders(step.Headers))
	if !step.Auth.Empty() {
		d.Set("auth", flattenStepAuth(step.Auth))
	}
	d.Set("body", step.Body)
	d.Set("form_parameter", flattenFormParameters(step.Form))
	d.Set("scripts", step.Scripts)
	d.Set("before_scripts", step.BeforeScripts)
	d.Set("note", step.Note)
	d.Set("skipped", step.Skipped)

	return nil
}

func resourceStepRequestUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerConfig).client

	opts := &runscope.StepUpdateRequestOpts{}
	expandStepGetOpts(d, &opts.StepGetRequestOpts)
	expandStepBaseOpts(d, &opts.StepRequestOpts)

	_, err := client.Step.UpdateRequest(ctx, opts)
	if err != nil {
		return diag.Errorf("Couldn't create step: %s", err)
	}

	return resourceStepRequestRead(ctx, d, meta)
}

func expandStepGetOpts(d *schema.ResourceData, opts *runscope.StepGetRequestOpts) {
	opts.Id = d.Id()
	expandStepUriOpts(d, &opts.StepUriOpts)
}

func expandStepBaseOpts(d *schema.ResourceData, opts *runscope.StepRequestOpts) {
	if v, ok := d.GetOk("method"); ok {
		opts.Method = v.(string)
	}
	if v, ok := d.GetOk("url"); ok {
		opts.StepURL = v.(string)
	}
	if v, ok := d.GetOk("variable"); ok {
		opts.Variables = expandStepVariables(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("assertion"); ok {
		opts.Assertions = expandStepAssertions(v.([]interface{}))
	}
	if v, ok := d.GetOk("header"); ok {
		opts.Headers = expandStepHeaders(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("auth"); ok {
		opts.Auth = expandStepAuth(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("body"); ok {
		opts.Body = v.(string)
	}
	if v, ok := d.GetOk("form_parameter"); ok {
		opts.Form = expandStepForm(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("scripts"); ok {
		opts.Scripts = expandStringSlice(v.([]interface{}))
	}
	if v, ok := d.GetOk("before_scripts"); ok {
		opts.BeforeScripts = expandStringSlice(v.([]interface{}))
	}
	if v, ok := d.GetOk("note"); ok {
		opts.Note = v.(string)
	}
	if v, ok := d.GetOk("skipped"); ok {
		opts.Skipped = v.(bool)
	}
}
