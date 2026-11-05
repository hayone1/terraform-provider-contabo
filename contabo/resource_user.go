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

var userLocales = []string{"de-DE", "de", "en-US", "en", "es", "es-ES", "pt-BR", "pt"}

type userResponse struct {
	UserId           string `json:"userId"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Email            string `json:"email"`
	EmailVerified    bool   `json:"emailVerified"`
	Enabled          bool   `json:"enabled"`
	Totp             bool   `json:"totp"`
	SecondaryEmail   bool   `json:"secondaryEmail"`
	Locale           string `json:"locale"`
	Owner            bool   `json:"owner"`
	SendInvoiceEmail bool   `json:"sendInvoiceEmail"`
	Roles            []struct {
		RoleId int64  `json:"roleId"`
		Name   string `json:"name"`
	} `json:"roles"`
}

func userPath(userId string) string {
	return "/v1/users/" + url.PathEscape(userId)
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a user of the customer account. The API sends the user an email to set a password; with it, the user gets their own API credentials limited to the permissions of their roles. Destroying the resource deletes the user. The account owner cannot be managed here.",
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Email address of the user. Activation and password reset links are sent to it. Changing it creates a new user.",
			},
			"first_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "First name of the user.",
			},
			"last_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Last name of the user.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the user can log in. Defaults to `true`.",
			},
			"totp": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether two-factor authentication with a time-based one-time password is enabled. Defaults to `false`.",
			},
			"secondary_email": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether two-factor authentication with a code sent to a secondary email address is enabled. Defaults to `false`.",
			},
			"locale": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "en",
				ValidateFunc: validation.StringInSlice(userLocales, false),
				Description:  "Locale of the user: `de-DE`, `de`, `en-US`, `en`, `es-ES`, `es`, `pt-BR` or `pt`. Defaults to `en`.",
			},
			"roles": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Ids of the roles of the user, for example from `contabo_role`.",
			},
			"send_invoice_email": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the user receives invoice emails. Defaults to `false`.",
			},
			"email_verified": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user has verified the email address.",
			},
			"owner": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the user is the account owner.",
			},
		},
	}
}

func expandUser(d *schema.ResourceData, create bool) map[string]interface{} {
	request := map[string]interface{}{
		"firstName":        d.Get("first_name").(string),
		"lastName":         d.Get("last_name").(string),
		"enabled":          d.Get("enabled").(bool),
		"totp":             d.Get("totp").(bool),
		"secondaryEmail":   d.Get("secondary_email").(bool),
		"locale":           d.Get("locale").(string),
		"roles":            setToInt64s(d.Get("roles").(*schema.Set)),
		"sendInvoiceEmail": d.Get("send_invoice_email").(bool),
	}
	if create {
		request["email"] = d.Get("email").(string)
	}
	return request
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	var res struct {
		Data []struct {
			UserId string `json:"userId"`
		} `json:"data"`
	}
	httpResp, err := apiRequest(ctx, client, http.MethodPost, "/v1/users", nil, expandUser(d, true), &res)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	if len(res.Data) != 1 {
		return MultipleDataObjectsError(nil)
	}
	d.SetId(res.Data[0].UserId)
	return resourceUserRead(ctx, d, m)
}

func retrieveUser(ctx context.Context, client *openapi.APIClient, userId string) (*userResponse, *http.Response, error) {
	var res struct {
		Data []userResponse `json:"data"`
	}
	httpResp, err := apiRequest(ctx, client, http.MethodGet, userPath(userId), nil, nil, &res)
	if err != nil {
		return nil, httpResp, err
	}
	if len(res.Data) != 1 {
		return nil, httpResp, fmt.Errorf("expected one user with id %s", userId)
	}
	return &res.Data[0], httpResp, nil
}

func flattenUser(user userResponse) map[string]interface{} {
	roles := make([]interface{}, 0, len(user.Roles))
	for _, role := range user.Roles {
		roles = append(roles, int(role.RoleId))
	}
	return map[string]interface{}{
		"email":              user.Email,
		"first_name":         user.FirstName,
		"last_name":          user.LastName,
		"enabled":            user.Enabled,
		"totp":               user.Totp,
		"secondary_email":    user.SecondaryEmail,
		"locale":             user.Locale,
		"roles":              roles,
		"send_invoice_email": user.SendInvoiceEmail,
		"email_verified":     user.EmailVerified,
		"owner":              user.Owner,
	}
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	user, httpResp, err := retrieveUser(ctx, client, d.Id())
	if err != nil {
		if removeFromStateIfNotFound(d, httpResp) {
			return nil
		}
		return apiErrorDiags(httpResp, err)
	}
	if user.Owner {
		return diag.Errorf("user %s is the account owner, which contabo_user does not manage", d.Id())
	}
	return setSchemaValues(d, resourceUser().Schema, flattenUser(*user))
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	httpResp, err := apiRequest(ctx, client, http.MethodPatch, userPath(d.Id()), nil, expandUser(d, false), nil)
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	user, httpResp, err := retrieveUser(ctx, client, d.Id())
	if err != nil {
		if isNotFound(httpResp) {
			return nil
		}
		return apiErrorDiags(httpResp, err)
	}
	if user.Owner {
		return diag.Errorf("refusing to delete user %s: it is the account owner", d.Id())
	}
	httpResp, err = apiRequest(ctx, client, http.MethodDelete, userPath(d.Id()), nil, nil, nil)
	if err != nil && !isNotFound(httpResp) {
		return apiErrorDiags(httpResp, err)
	}
	return nil
}

// contabo_users data source

func dataSourceUsers() *schema.Resource {
	userAttributes := map[string]*schema.Schema{
		"id": {Type: schema.TypeString, Computed: true, Description: "Id of the user."},
	}
	for name, attribute := range resourceUser().Schema {
		computed := &schema.Schema{Type: attribute.Type, Elem: attribute.Elem, Computed: true, Description: attribute.Description}
		userAttributes[name] = computed
	}
	return &schema.Resource{
		Description: "Lists the users of the customer account, including the owner.",
		ReadContext: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Only the user with this email address.",
			},
			"users": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The matching users.",
				Elem:        &schema.Resource{Schema: userAttributes},
			},
		},
	}
}

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*openapi.APIClient)
	email := d.Get("email").(string)
	query := url.Values{}
	if email != "" {
		query.Set("email", email)
	}
	var users []userResponse
	httpResp, err := listAll(ctx, client, "/v1/users", query, func(body []byte) (int64, error) {
		var page struct {
			Data       []userResponse `json:"data"`
			Pagination pagination     `json:"_pagination"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return 0, err
		}
		users = append(users, page.Data...)
		return page.Pagination.TotalPages, nil
	})
	if err != nil {
		return apiErrorDiags(httpResp, err)
	}
	out := []interface{}{}
	ids := []string{}
	for _, user := range users {
		if email != "" && !strings.EqualFold(user.Email, email) {
			continue
		}
		values := flattenUser(user)
		values["id"] = user.UserId
		out = append(out, values)
		ids = append(ids, user.UserId)
	}
	d.SetId(fmt.Sprintf("users-%d", hashStrings(append(ids, email))))
	if err := d.Set("users", out); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
