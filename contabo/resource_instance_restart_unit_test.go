package contabo

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestUnitInstanceRestartResource(t *testing.T) {
	api := newFakeAPI(t)
	instanceId := api.addInstance(map[string]interface{}{"status": "stopped"})
	restartPath := fmt.Sprintf("POST /v1/compute/instances/%d/actions/restart", instanceId)

	config := func(trigger string) string {
		return api.providerConfig() + fmt.Sprintf(`
resource "contabo_instance_restart" "boot" {
  instance_id = "%d"
  triggers = {
    install = %q
  }
}
`, instanceId, trigger)
	}

	restarts := func(want int) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			if got := api.callCount(restartPath); got != want {
				return fmt.Errorf("expected %d restarts, got %d", want, got)
			}
			return nil
		}
	}

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config("first"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("contabo_instance_restart.boot", "status", "running"),
					resource.TestCheckResourceAttrSet("contabo_instance_restart.boot", "restarted_at"),
					restarts(1),
				),
			},
			{
				// The same triggers must not restart the instance again.
				Config: config("first"),
				Check:  restarts(1),
			},
			{
				Config: config("second"),
				Check:  restarts(2),
			},
		},
	})
}
