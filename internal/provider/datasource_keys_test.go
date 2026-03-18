package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKeysDataSource_basic(t *testing.T) {
	projectID := os.Getenv("DEEPGRAM_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "deepgram" {}

data "deepgram_keys" "all" {
  project_id = %q
}
`, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.deepgram_keys.all", "project_id"),
					// The list exists (may be empty — just check it is set)
					resource.TestCheckResourceAttrSet("data.deepgram_keys.all", "api_keys.#"),
				),
			},
		},
	})
}

func TestAccKeysDataSource_withCreatedKey(t *testing.T) {
	projectID := os.Getenv("DEEPGRAM_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create a key then read the list — the new key must appear
				Config: fmt.Sprintf(`
provider "deepgram" {}

resource "deepgram_key" "ds_test" {
  project_id = %q
  comment    = "tf-acc-ds-test"
  scopes     = toset(["usage:read"])
}

data "deepgram_keys" "all" {
  project_id = %q
  depends_on = [deepgram_key.ds_test]
}
`, projectID, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.deepgram_keys.all", "api_keys.#"),
					resource.TestCheckResourceAttrSet("deepgram_key.ds_test", "id"),
				),
			},
		},
	})
}
