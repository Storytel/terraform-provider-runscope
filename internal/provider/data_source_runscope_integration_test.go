package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceRunscopeIntegration_Basic(t *testing.T) {
	teamID := os.Getenv("RUNSCOPE_TEAM_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccDataSourceRunscopeIntegrationConfig, teamID),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceRunscopeIntegration("data.runscope_integration.by_type"),
				),
			},
		},
	})
}

func testAccDataSourceRunscopeIntegration(dataSource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[dataSource]
		a := r.Primary.Attributes

		if a["id"] == "" {
			return fmt.Errorf("Expected to get an integration ID from runscope data resource")
		}

		if a["type"] != "slack" {
			return fmt.Errorf("Expected to get an integration type slack from runscope data resource")
		}

		if a["description"] == "" {
			return fmt.Errorf("Expected to get an integration description from runscope data resource")
		}

		return nil
	}
}

const testAccDataSourceRunscopeIntegrationConfig = `
data "runscope_integration" "by_type" {
	team_uuid = "%s"
	type      = "slack"
}
`

func TestAccDataSourceRunscopeIntegration_Filter(t *testing.T) {
	teamID := os.Getenv("RUNSCOPE_TEAM_ID")
	integrationDesc, ok := os.LookupEnv("RUNSCOPE_INTEGRATION_DESC")
	if !ok {
		t.Skip("RUNSCOPE_INTEGRATION_DESC should be set")
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccDataSourceRunscopeIntegrationFilterConfig, teamID, integrationDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceRunscopeIntegrationFilter("data.runscope_integration.by_type"),
				),
			},
		},
	})
}

func testAccDataSourceRunscopeIntegrationFilter(dataSource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		integrationDesc := os.Getenv("RUNSCOPE_INTEGRATION_DESC")

		r := s.RootModule().Resources[dataSource]
		if r == nil {
			return fmt.Errorf("expected integration description to be %s, actual nil", integrationDesc)
		}

		a := r.Primary.Attributes

		if a["description"] != integrationDesc {
			return fmt.Errorf("expected integration description %s to be %s", a["description"], integrationDesc)
		}

		return nil
	}
}

const testAccDataSourceRunscopeIntegrationFilterConfig = `
data "runscope_integration" "by_type" {
  team_uuid = "%s"
  type      = "slack"

  filter {
    name   = "type"
    values = ["slack"]
  }

  filter {
    name   = "description"
    values = ["%s", "other test description"]
  }
}
`
