package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccSnapshotResourceConfig(name string, size int) string {
	return fmt.Sprintf(`
resource "sagadata_snapshot" "test" {
  name = %[1]q
  size = %[2]q
}
`, name, size)
}

func TestAccSnapshotResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccSnapshotResourceConfig("one", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestCheckResourceAttr("sagadata_snapshot.test", "id", "ssh-key-id"),
					resource.TestCheckResourceAttr("sagadata_snapshot.test", "name", "one"),
					resource.TestCheckResourceAttr("sagadata_snapshot.test", "size", "1"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "sagadata_snapshot.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + testAccSnapshotResourceConfig("two", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("sagadata_snapshot.test", "name", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
