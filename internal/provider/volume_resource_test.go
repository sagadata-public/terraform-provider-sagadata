package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccVolumeResourceConfig(name string, size int) string {
	return fmt.Sprintf(`
resource "sagadata_volume" "test" {
  name = %[1]q
  size = %[2]q
}
`, name, size)
}

func TestAccVolumeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccVolumeResourceConfig("one", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("sagadata_volume.test", "id", "ssh-key-id"),
					resource.TestCheckResourceAttr("sagadata_volume.test", "name", "one"),
					resource.TestCheckResourceAttr("sagadata_volume.test", "size", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sagadata_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccVolumeResourceConfig("two", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sagadata_volume.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
