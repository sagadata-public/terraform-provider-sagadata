package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccSecurityGroupResourceConfig(name, todo string) string {
	return fmt.Sprintf(`
resource "sagadata_security_group" "test" {
  name = %[1]q
  todo = %[2]q
}
`, name, todo)
}

func TestAccSecurityGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccSecurityGroupResourceConfig("one", samplePublicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("sagadata_security_group.test", "id", "ssh-key-id"),
					resource.TestCheckResourceAttr("sagadata_security_group.test", "name", "one"),
					resource.TestCheckResourceAttr("sagadata_security_group.test", "todo", "todo"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sagadata_security_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccSecurityGroupResourceConfig("two", samplePublicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sagadata_security_group.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
