package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccKeyConfig(projectID, comment string, scopes []string, tags []string) string {
	scopeStr := `toset([`
	for i, s := range scopes {
		if i > 0 {
			scopeStr += ", "
		}
		scopeStr += fmt.Sprintf("%q", s)
	}
	scopeStr += `])`

	tagStr := `[`
	for i, t := range tags {
		if i > 0 {
			tagStr += ", "
		}
		tagStr += fmt.Sprintf("%q", t)
	}
	tagStr += `]`

	return fmt.Sprintf(`
provider "deepgram" {}

resource "deepgram_key" "test" {
  project_id = %q
  comment    = %q
  scopes     = %s
  tags       = %s
}
`, projectID, comment, scopeStr, tagStr)
}

func TestAccKey_basic(t *testing.T) {
	projectID := os.Getenv("DEEPGRAM_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccKeyConfig(projectID, "tf-acc-test-key", []string{"usage:read"}, []string{"terraform-test"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("deepgram_key.test", "comment", "tf-acc-test-key"),
					resource.TestCheckResourceAttr("deepgram_key.test", "project_id", projectID),
					resource.TestCheckResourceAttrSet("deepgram_key.test", "id"),
					resource.TestCheckResourceAttrSet("deepgram_key.test", "key"),
					resource.TestCheckTypeSetElemAttr("deepgram_key.test", "scopes.*", "usage:read"),
				),
			},
			// Import by project_id/key_id
			{
				ResourceName:            "deepgram_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key"}, // key is only returned on creation
				ImportStateIdFunc:       testAccKeyImportID("deepgram_key.test"),
			},
		},
	})
}

func TestAccKey_withExpiration(t *testing.T) {
	projectID := os.Getenv("DEEPGRAM_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "deepgram" {}

resource "deepgram_key" "expiring" {
  project_id      = %q
  comment         = "tf-acc-expiring-key"
  scopes          = toset(["usage:read"])
  expiration_date = "2027-01-01T00:00:00Z"
}
`, projectID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("deepgram_key.expiring", "comment", "tf-acc-expiring-key"),
					resource.TestCheckResourceAttrSet("deepgram_key.expiring", "id"),
					resource.TestCheckResourceAttrSet("deepgram_key.expiring", "key"),
				),
			},
		},
	})
}

func TestAccKey_replaceOnChange(t *testing.T) {
	projectID := os.Getenv("DEEPGRAM_PROJECT_ID")

	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig(projectID, "tf-acc-original", []string{"usage:read"}, []string{}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("deepgram_key.test", "comment", "tf-acc-original"),
					// capture the ID for comparison in the next step
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["deepgram_key.test"]
						firstID = rs.Primary.ID
						return nil
					},
				),
			},
			// Changing comment forces replace → new key with different ID
			{
				Config: testAccKeyConfig(projectID, "tf-acc-updated", []string{"usage:read"}, []string{}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("deepgram_key.test", "comment", "tf-acc-updated"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["deepgram_key.test"]
						if rs.Primary.ID == firstID {
							return fmt.Errorf("expected a new key ID after replace, got the same: %s", firstID)
						}
						return nil
					},
				),
			},
		},
	})
}

// testAccKeyImportID returns an ImportStateIdFunc that produces "project_id/key_id".
func testAccKeyImportID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource %s not found", resourceName)
		}
		projectID := rs.Primary.Attributes["project_id"]
		keyID := rs.Primary.ID
		return fmt.Sprintf("%s/%s", projectID, keyID), nil
	}
}
