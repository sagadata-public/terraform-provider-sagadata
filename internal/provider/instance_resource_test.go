package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccInstanceResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "sagadata_instance" "test" {
  name = %[1]q
}
`, name)
}

func TestAccInstanceResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccInstanceResourceConfig("one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sagadata_instance.test", "id", "instance-id"),
					resource.TestCheckResourceAttr("sagadata_instance.test", "name", "one"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sagadata_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccInstanceResourceConfig("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sagadata_instance.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
