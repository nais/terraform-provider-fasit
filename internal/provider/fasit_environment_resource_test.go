package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TODO: test not working yet (like the other tests), needs to be fixed
func TestAccEnvironmentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccExampleResourceEnvironmentConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fasit_environment.test", "name", "test-environment"),
					resource.TestCheckResourceAttr("fasit_environment.test", "tenant_id", "tenant-id"),
					resource.TestCheckResourceAttr("fasit_environment.test", "kind", "test-kind"),
					resource.TestCheckResourceAttr("fasit_environment.test", "labels.foo", "bar"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "fasit_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccExampleResourceEnvironmentConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fasit_enviroment.test", "labels", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccExampleResourceEnvironmentConfig() string {
	return `
	provider "fasit" {
		url = "asdf"
	}
	resource "fasit_environment" "test" {
	  name = "test-environment"
	  tenant_id = "tenant-id"
	  kind = "test-kind"
	  labels = {
         "foo": "bar" 
      }
	}`
}
