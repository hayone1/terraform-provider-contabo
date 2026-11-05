package contabo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"contabo.com/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// vipResponse mirrors a virtual IP (additional or floating IP) of the API.
type vipResponse struct {
	VipId               string `json:"vipId"`
	DataCenter          string `json:"dataCenter"`
	Region              string `json:"region"`
	ResourceId          string `json:"resourceId"`
	ResourceType        string `json:"resourceType"`
	ResourceName        string `json:"resourceName"`
	ResourceDisplayName string `json:"resourceDisplayName"`
	IpVersion           string `json:"ipVersion"`
	Type                string `json:"type"`
	V4                  struct {
		Ip          string  `json:"ip"`
		NetmaskCidr float64 `json:"netmaskCidr"`
		Gateway     string  `json:"gateway"`
		Network     string  `json:"network"`
		Broadcast   string  `json:"broadcast"`
	} `json:"v4"`
}

// assigned reports whether the VIP is assigned to a resource. The API uses
// an empty string or "null" for unassigned VIPs.
func (v vipResponse) assigned() bool {
	return v.ResourceId != "" && v.ResourceId != "null" && v.ResourceType != "" && v.ResourceType != "null"
}

func retrieveVip(ctx context.Context, client *openapi.APIClient, ip string) (*vipResponse, *http.Response, error) {
	var res struct {
		Data []vipResponse `json:"data"`
	}
	httpResp, err := apiRequest(ctx, client, http.MethodGet, "/v1/vips/"+url.PathEscape(ip), nil, nil, &res)
	if err != nil {
		return nil, httpResp, err
	}
	if len(res.Data) != 1 {
		return nil, httpResp, fmt.Errorf("expected one VIP with ip %s, got %d", ip, len(res.Data))
	}
	return &res.Data[0], httpResp, nil
}

func listVips(ctx context.Context, client *openapi.APIClient, query url.Values) ([]vipResponse, *http.Response, error) {
	var vips []vipResponse
	httpResp, err := listAll(ctx, client, "/v1/vips", query, func(body []byte) (int64, error) {
		var page struct {
			Data       []vipResponse `json:"data"`
			Pagination pagination    `json:"_pagination"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return 0, err
		}
		vips = append(vips, page.Data...)
		return page.Pagination.TotalPages, nil
	})
	return vips, httpResp, err
}

func vipAssignmentPath(ip, resourceType, resourceId string) string {
	return fmt.Sprintf("/v1/vips/%s/%s/%s", url.PathEscape(ip), url.PathEscape(resourceType), url.PathEscape(resourceId))
}

func resourceVipAssignment() *schema.Resource {
	return &schema.Resource{
		Description:   "Assigns a virtual IP (an additional or floating IPv4 address bought as an add-on) to an instance or a bare metal server. VIPs are ordered in the customer control panel or with the `add_ons` of an instance; this resource only manages the assignment. Destroying it unassigns the IP.",
		CreateContext: resourceVipAssignmentCreate,
		ReadContext:   resourceVipAssignmentRead,
		DeleteContext: resourceVipAssignmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"ip": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
				Description:  "The virtual IPv4 address. It is also the id of the resource.",
			},
			"resource_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "instances",
				ValidateFunc: validation.StringInSlice([]string{"instances", "bare-metal"}, false),
				Description:  "Type of the resource to assign the IP to: `instances` or `bare-metal`. Defaults to `instances`.",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the instance or bare metal server.",
			},
			"vip_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Id of the VIP.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the VIP: `additional` or `floating`.",
			},
			"region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Region of the VIP.",
			},
			"data_center": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data center of the VIP.",
			},
			"netmask_cidr": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Netmask of the VIP in CIDR notation.",
			},
			"gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway of the VIP.",
			},
		},
	}
}

func resourceVipAssignmentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	ip := d.Get("ip").(string)
	resourceType := d.Get("resource_type").(string)
	resourceId := d.Get("resource_id").(string)

	vip, httpResp, err := retrieveVip(ctx, client, ip)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	if vip.assigned() && !(vip.ResourceId == resourceId && vip.ResourceType == resourceType) {
		return diag.Errorf("VIP %s is assigned to %s %s; unassign it first", ip, vip.ResourceType, vip.ResourceId)
	}
	if !vip.assigned() {
		httpResp, err = apiRequest(ctx, client, http.MethodPost, vipAssignmentPath(ip, resourceType, resourceId), nil, nil, nil)
		if err != nil {
			return apiErrorDiags(httpResp, err)
		}
	}
	d.SetId(ip)
	return resourceVipAssignmentRead(ctx, d, m)
}

func resourceVipAssignmentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	vip, httpResp, err := retrieveVip(ctx, client, d.Id())
	if err != nil {
		if removeFromStateIfNotFound(d, httpResp) {
			return nil
		}
		return apiErrorDiags(httpResp, err)
	}
	if !vip.assigned() {
		d.SetId("")
		return nil
	}
	return setSchemaValues(d, resourceVipAssignment().Schema, map[string]interface{}{
		"ip":            vip.V4.Ip,
		"resource_type": vip.ResourceType,
		"resource_id":   vip.ResourceId,
		"vip_id":        vip.VipId,
		"type":          vip.Type,
		"region":        vip.Region,
		"data_center":   vip.DataCenter,
		"netmask_cidr":  int(vip.V4.NetmaskCidr),
		"gateway":       vip.V4.Gateway,
	})
}

func resourceVipAssignmentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	path := vipAssignmentPath(d.Id(), d.Get("resource_type").(string), d.Get("resource_id").(string))
	httpResp, err := apiRequest(ctx, client, http.MethodDelete, path, nil, nil, nil)
	if err != nil && !isNotFound(httpResp) {
		return apiErrorDiags(httpResp, err)
	}
	return nil
}

func vipAttributes() map[string]*schema.Schema {
	computed := func(t schema.ValueType, description string) *schema.Schema {
		return &schema.Schema{Type: t, Computed: true, Description: description}
	}
	return map[string]*schema.Schema{
		"ip":                    computed(schema.TypeString, "The IPv4 address."),
		"vip_id":                computed(schema.TypeString, "Id of the VIP."),
		"type":                  computed(schema.TypeString, "Type of the VIP: `additional` or `floating`."),
		"ip_version":            computed(schema.TypeString, "IP version, `v4`."),
		"region":                computed(schema.TypeString, "Region of the VIP."),
		"data_center":           computed(schema.TypeString, "Data center of the VIP."),
		"resource_id":           computed(schema.TypeString, "Id of the resource the VIP is assigned to, empty when unassigned."),
		"resource_type":         computed(schema.TypeString, "Type of the resource the VIP is assigned to, empty when unassigned."),
		"resource_name":         computed(schema.TypeString, "Name of the resource the VIP is assigned to."),
		"resource_display_name": computed(schema.TypeString, "Display name of the resource the VIP is assigned to."),
		"netmask_cidr":          computed(schema.TypeInt, "Netmask in CIDR notation."),
		"gateway":               computed(schema.TypeString, "Gateway."),
		"network":               computed(schema.TypeString, "Network address."),
		"broadcast":             computed(schema.TypeString, "Broadcast address."),
	}
}

func flattenVip(vip vipResponse) map[string]interface{} {
	resourceId, resourceType := vip.ResourceId, vip.ResourceType
	if !vip.assigned() {
		resourceId, resourceType = "", ""
	}
	return map[string]interface{}{
		"ip":                    vip.V4.Ip,
		"vip_id":                vip.VipId,
		"type":                  vip.Type,
		"ip_version":            vip.IpVersion,
		"region":                vip.Region,
		"data_center":           vip.DataCenter,
		"resource_id":           resourceId,
		"resource_type":         resourceType,
		"resource_name":         vip.ResourceName,
		"resource_display_name": vip.ResourceDisplayName,
		"netmask_cidr":          int(vip.V4.NetmaskCidr),
		"gateway":               vip.V4.Gateway,
		"network":               vip.V4.Network,
		"broadcast":             vip.V4.Broadcast,
	}
}

func dataSourceVips() *schema.Resource {
	return &schema.Resource{
		Description: "Lists the virtual IPs (additional and floating IPv4 addresses) of the account, optionally filtered.",
		ReadContext: dataSourceVipsRead,
		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"additional", "floating"}, false),
				Description:  "Only VIPs of this type: `additional` or `floating`.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Only VIPs in this region.",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Only VIPs assigned to this resource.",
			},
			"unassigned": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Only VIPs that are not assigned to a resource.",
			},
			"ips": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The matching IP addresses.",
			},
			"vips": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The matching VIPs.",
				Elem:        &schema.Resource{Schema: vipAttributes()},
			},
		},
	}
}

func dataSourceVipsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	query := url.Values{}
	if v := d.Get("type").(string); v != "" {
		query.Set("type", v)
	}
	if v := d.Get("resource_id").(string); v != "" {
		query.Set("resourceId", v)
	}
	vips, httpResp, err := listVips(ctx, client, query)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	region := d.Get("region").(string)
	resourceId := d.Get("resource_id").(string)
	vipType := d.Get("type").(string)
	unassigned := d.Get("unassigned").(bool)

	ips := []string{}
	out := []interface{}{}
	for _, vip := range vips {
		if region != "" && !strings.EqualFold(vip.Region, region) {
			continue
		}
		if resourceId != "" && vip.ResourceId != resourceId {
			continue
		}
		if vipType != "" && vip.Type != vipType {
			continue
		}
		if unassigned && vip.assigned() {
			continue
		}
		ips = append(ips, vip.V4.Ip)
		out = append(out, flattenVip(vip))
	}
	d.SetId(fmt.Sprintf("vips-%d", hashStrings(append([]string{vipType, region, resourceId, fmt.Sprint(unassigned)}, ips...))))
	if err := d.Set("ips", ips); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vips", out); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func dataSourceVip() *schema.Resource {
	attributes := vipAttributes()
	attributes["ip"] = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.IsIPv4Address,
		Description:  "The IPv4 address of the VIP.",
	}
	return &schema.Resource{
		Description: "Reads a virtual IP (additional or floating IPv4 address) by its address.",
		ReadContext: dataSourceVipRead,
		Schema:      attributes,
	}
}

func dataSourceVipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	vip, httpResp, err := retrieveVip(ctx, client, d.Get("ip").(string))
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	d.SetId(vip.V4.Ip)
	return setSchemaValues(d, dataSourceVip().Schema, flattenVip(*vip))
}
