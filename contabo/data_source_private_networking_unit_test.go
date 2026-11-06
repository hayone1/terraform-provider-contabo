package contabo

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// The data source failed with "instances.0.instance_id: ” expected type
// 'string', got unconvertible type 'int64'" for every network with members.
func TestUnitPrivateNetworkDataSourceWithInstances(t *testing.T) {
	api := newFakeAPI(t)
	api.mux.HandleFunc("GET /v1/private-networks/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "12345" {
			writeAPIError(w, http.StatusNotFound, "private network not found")
			return
		}
		writeData(w, http.StatusOK, map[string]interface{}{
			"tenantId":         "INT",
			"customerId":       "12345",
			"privateNetworkId": 12345,
			"dataCenter":       "European Union 2",
			"region":           "EU",
			"regionName":       "European Union",
			"name":             "shared",
			"description":      "",
			"cidr":             "10.0.0.0/22",
			"availableIps":     1018,
			"createdDate":      "2026-01-05T10:00:00Z",
			"updatedDate":      "2026-01-05T10:00:00Z",
			"instances": []map[string]interface{}{
				{
					"instanceId":   200000001,
					"displayName":  "first",
					"name":         "vmi200000001",
					"productId":    "V91",
					"status":       "ok",
					"errorMessage": "",
					"privateIpConfig": map[string]interface{}{
						"v4": []map[string]interface{}{{"ip": "10.0.0.1", "netmaskCidr": 22, "gateway": "10.0.0.0"}},
					},
				},
				{
					"instanceId":   200000003,
					"displayName":  "second",
					"name":         "vmi200000003",
					"productId":    "V166",
					"status":       "ok",
					"errorMessage": "",
					"privateIpConfig": map[string]interface{}{
						"v4": []map[string]interface{}{{"ip": "10.0.0.3", "netmaskCidr": 22, "gateway": "10.0.0.0"}},
					},
				},
			},
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{{
			Config: api.providerConfig() + `
data "contabo_private_network" "shared" {
  id = "12345"
}
`,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.contabo_private_network.shared", "cidr", "10.0.0.0/22"),
				resource.TestCheckResourceAttr("data.contabo_private_network.shared", "instances.#", "2"),
				resource.TestCheckResourceAttr("data.contabo_private_network.shared", "instances.1.instance_id", "200000003"),
				resource.TestCheckResourceAttr("data.contabo_private_network.shared", "instances.1.private_ip_config.0.v4.0.ip", "10.0.0.3"),
				resource.TestCheckResourceAttr("data.contabo_private_network.shared", "instance_ids.#", "2"),
			),
		}},
	})
}
