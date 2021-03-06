package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-runscope/internal/runscope"
)

func resourceRunscopeStepSubtest() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStepSubtestCreate,
		ReadContext:   resourceStepSubtestRead,
		UpdateContext: resourceStepSubtestUpdate,
		DeleteContext: resourceStepDelete,
		Schema: map[string]*schema.Schema{
			"bucket_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The bucket of the test this step belong to.",
			},
			"test_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the test this step belongs to.",
			},
			"source_bucket_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the bucket where the to-be-invoked test resides.",
			},
			"source_test_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the test to invoke as a subtest.",
			},
			"source_environment_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the environment which the subtest should run under.",
			},
			"use_parent_environment": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "The ID of the environment which the subtest should run under.",
			},
			"variable": {
				Type: schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the extracted variable, which can be used to reference the value elsewhere.",
						},
						"property": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The property to extract.",
						},
						"source": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(stepSources, false),
							Description:  "The source of the property, e.g. `response_json`.",
						},
					},
				},
				Optional:    true,
				Description: "Variables to extract from the subtest.",
			},
			"assertion": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(stepSources, false),
							Description:  "The source of the property to assert.",
						},
						"property": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The property to assert on.",
						},
						"comparison": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(stepComparisons, false),
							Description:  "The comparison type, eg `equal` or `has_key`.",
						},
						"value": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The value to assert the source.property has.",
						},
					},
				},
				Optional:    true,
				Description: "Assertions to ensure the subtest ran successfully.",
			},
		},
	}
}

func resourceStepSubtestCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerConfig).client

	var opts runscope.StepCreateSubtestOpts
	expandStepUriOpts(d, &opts.StepUriOpts)
	expandStepSubtestOpts(d, &opts.StepSubtestOpts)

	step, err := client.Step.CreateSubtest(ctx, &opts)
	if err != nil {
		if isNotFound(err) {
			d.SetId("")
			return nil
		}

		return diag.Errorf("couldn't read step: %s", err)
	}

	d.SetId(step.ID)

	return resourceStepSubtestRead(ctx, d, meta)
}

func resourceStepSubtestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerConfig).client

	opts := &runscope.StepGetRequestOpts{
		StepUriOpts: runscope.StepUriOpts{
			BucketId: d.Get("bucket_id").(string),
			TestId:   d.Get("test_id").(string),
		},
		Id: d.Id(),
	}

	step, err := client.Step.GetSubtest(ctx, opts)
	if err != nil {
		if isNotFound(err) {
			d.SetId("")
			return nil
		}
		var runscopeErr runscope.Error
		is := errors.As(err, &runscopeErr)

		return diag.Errorf("couldn't (is=%t) read step: %s", is, err)
	}

	d.Set("source_bucket_id", step.BucketKey)
	d.Set("source_test_id", step.TestUUID)
	d.Set("source_environment_id", step.EnvironmentUUID)
	d.Set("use_parent_environment", step.UseParentEnvironment)
	d.Set("variable", flattenStepVariables(step.Variables))
	d.Set("assertion", flattenStepAssertions(step.Assertions))

	return nil
}

func resourceStepSubtestUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerConfig).client

	opts := &runscope.StepUpdateSubtestOpts{}
	expandStepGetOpts(d, &opts.StepGetRequestOpts)
	expandStepSubtestOpts(d, &opts.StepSubtestOpts)

	_, err := client.Step.UpdateSubtest(ctx, opts)
	if err != nil {
		return diag.Errorf("Couldn't create step: %s", err)
	}

	return resourceStepSubtestRead(ctx, d, meta)
}

func expandStepSubtestOpts(d *schema.ResourceData, opts *runscope.StepSubtestOpts) {
	if v, ok := d.GetOk("source_bucket_id"); ok {
		opts.BucketKey = v.(string)
	}
	if v, ok := d.GetOk("source_test_id"); ok {
		opts.TestUUID = v.(string)
	}
	if v, ok := d.GetOk("source_environment_id"); ok {
		opts.EnvironmentUUID = v.(string)
	}
	if v, ok := d.GetOk("use_parent_environment"); ok {
		opts.UseParentEnvironment = v.(bool)
	}
	if v, ok := d.GetOk("variable"); ok {
		opts.Variables = expandStepVariables(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("assertion"); ok {
		opts.Assertions = expandStepAssertions(v.([]interface{}))
	}
}
