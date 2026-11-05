package contabo

import (
	"context"
	"net/url"

	"contabo.com/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceFirewall() *schema.Resource {
	return &schema.Resource{
		Description: "Reads a firewall by id or by name.",
		ReadContext: dataSourceFirewallRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "The identifier of the Firewall. Set either `id` or `name`.",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
				Description:  "The name of the Firewall. Set either `id` or `name`; the name must match exactly one firewall.",
			},
			"created_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation date of the Firewall.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the Firewall.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the Firewall: `active` or `inactive`.",
			},
			"instance_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Computed:    true,
				Description: "Ids of the instances the firewall applies to.",
			},
			"instances_status": firewallInstancesStatusSchema(),
			"rules": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Rules of the firewall.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inbound": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Inbound rules.",
							Elem:        firewallRuleSchema(true),
						},
					},
				},
			},
		},
	}
}

func dataSourceFirewallRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)

	if firewallId := d.Get("id").(string); firewallId != "" {
		firewall, httpResp, err := retrieveFirewall(ctx, client, firewallId)
		if err != nil {
			return apiErrorDiags(httpResp, err)
		}
		return setFirewallData(*firewall, d)
	}

	name := d.Get("name").(string)
	firewalls, httpResp, err := listFirewalls(ctx, client, url.Values{"name": []string{name}})
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	var matches []firewallResponse
	for _, firewall := range firewalls {
		if firewall.Name == name {
			matches = append(matches, firewall)
		}
	}
	switch len(matches) {
	case 0:
		return diag.Errorf("no firewall named %q", name)
	case 1:
		return setFirewallData(matches[0], d)
	default:
		return diag.Errorf("%d firewalls are named %q; look the firewall up by id", len(matches), name)
	}
}
