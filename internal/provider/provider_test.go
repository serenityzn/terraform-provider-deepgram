package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/serenityzn/terraform-provider-deepgram/internal/provider"
)

// testAccProtoV6ProviderFactories is used by acceptance tests to instantiate
// the provider under test.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"deepgram": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck validates that the required environment variables are set
// before running acceptance tests.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("DEEPGRAM_API_KEY"); v == "" {
		t.Skip("DEEPGRAM_API_KEY must be set to run acceptance tests")
	}
	if v := os.Getenv("DEEPGRAM_PROJECT_ID"); v == "" {
		t.Skip("DEEPGRAM_PROJECT_ID must be set to run acceptance tests")
	}
}

// testAccProviderConfig returns a provider block using env vars.
const testAccProviderConfig = `
provider "deepgram" {}
`

// TestProvider_basic verifies the provider can be configured without error.
func TestProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig,
			},
		},
	})
}
