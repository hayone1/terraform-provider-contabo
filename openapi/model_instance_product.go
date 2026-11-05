/*
Contabo API

# Introduction  Contabo API allows you to manage your resources using HTTP requests. This documentation includes a set of HTTP endpoints that are designed to RESTful principles. Each endpoint includes descriptions, request syntax, and examples.  Contabo provides also a CLI tool which enables you to manage your resources easily from the command line. [CLI Download and  Installation instructions.](https://github.com/contabo/cntb)  ## Product documentation  If you are looking for description about the products themselves and their usage in general or for specific purposes, please check the [Contabo Product Documentation](https://docs.contabo.com/).  ## Getting Started  In order to use the Contabo API you will need the following credentials which are available from the [Customer Control Panel](https://my.contabo.com/api/details): 1. ClientId 2. ClientSecret 3. API User (your email address to login to the [Customer Control Panel](https://my.contabo.com/api/details)) 4. API Password (this is a new password which you'll set or change in the [Customer Control Panel](https://my.contabo.com/api/details))  You can either use the API directly or by using the `cntb` CLI (Command Line Interface) tool.  ### Using the API directly  #### Via `curl` for Linux/Unix like systems  This requires `curl` and `jq` in your shell (e.g. `bash`, `zsh`). Please replace the first four placeholders with actual values.  ```sh CLIENT_ID=<ClientId from Customer Control Panel> CLIENT_SECRET=<ClientSecret from Customer Control Panel> API_USER=<API User from Customer Control Panel> API_PASSWORD='<API Password from Customer Control Panel>' ACCESS_TOKEN=$(curl -d \"client_id=$CLIENT_ID\" -d \"client_secret=$CLIENT_SECRET\" --data-urlencode \"username=$API_USER\" --data-urlencode \"password=$API_PASSWORD\" -d 'grant_type=password' 'https://auth.contabo.com/auth/realms/contabo/protocol/openid-connect/token' | jq -r '.access_token') # get list of your instances curl -X GET -H \"Authorization: Bearer $ACCESS_TOKEN\" -H \"x-request-id: 51A87ECD-754E-4104-9C54-D01AD0F83406\" \"https://api.contabo.com/v1/compute/instances\" | jq ```  #### Via `PowerShell` for Windows  Please open `PowerShell` and execute the following code after replacing the first four placeholders with actual values.  ```powershell $client_id='<ClientId from Customer Control Panel>' $client_secret='<ClientSecret from Customer Control Panel>' $api_user='<API User from Customer Control Panel>' $api_password='<API Password from Customer Control Panel>' $body = @{grant_type='password' client_id=$client_id client_secret=$client_secret username=$api_user password=$api_password} $response = Invoke-WebRequest -Uri 'https://auth.contabo.com/auth/realms/contabo/protocol/openid-connect/token' -Method 'POST' -Body $body $access_token = (ConvertFrom-Json $([String]::new($response.Content))).access_token # get list of your instances $headers = @{} $headers.Add(\"Authorization\",\"Bearer $access_token\") $headers.Add(\"x-request-id\",\"51A87ECD-754E-4104-9C54-D01AD0F83406\") Invoke-WebRequest -Uri 'https://api.contabo.com/v1/compute/instances' -Method 'GET' -Headers $headers ```  ### Using the Contabo API via the `cntb` CLI tool  1. Download `cntb` for your operating system (MacOS, Windows and Linux supported) [here](https://github.com/contabo/cntb) 2. Unzip the downloaded file 3. You might move the executable to any location on your disk. You may update your `PATH` environment variable for easier invocation. 4. Configure it once to use your credentials                                                                ```sh    cntb config set-credentials --oauth2-clientid=<ClientId from Customer Control Panel> --oauth2-client-secret=<ClientSecret from Customer Control Panel> --oauth2-user=<API User from Customer Control Panel> --oauth2-password='<API Password from Customer Control Panel>'    ```  5. Use the CLI                                                                ```sh    # get list of your instances    cntb get instances    # help    cntb help    ```  ## API Overview  ### [Compute Management](#tag/Instances)  The Compute Management API allows you to manage compute resources (e.g. creation, deletion, starting, stopping) of VPS and VDS (please note that Storage VPS are not supported via API or CLI) as well as managing snapshots and custom images. It also offers you to take advantage of [cloud-init](https://cloud-init.io/) at least on our default / standard images (for custom images you'll need to provide cloud-init support packages). The API offers provisioning of cloud-init scripts via the `user_data` field.  Custom images must be provided in `.qcow2` or `.iso` format. This gives you even more flexibility for setting up your environment.  ### [Object Storage](#tag/Object-Storages)  The Object Storage API allows you to order, upgrade, cancel and control the auto-scaling feature for [S3](https://en.wikipedia.org/wiki/Amazon_S3) compatible object storage. You may also get some usage statistics. You can only buy one object storage per location. In case you need more storage space in a location you can purchase more space or enable the auto-scaling feature to purchase automatically more storage space up to the specified monthly limit.  Please note that this is not the S3 compatible API. It is not documented here. The S3 compatible API needs to be used with the corresponding credentials, namely an `access_key` and `secret_key`. Those can be retrieved by invoking the User Management API. All purchased object storages in different locations share the same credentials. You are free to use S3 compatible tools like [`aws`](https://aws.amazon.com/cli/) cli or similar.  ### [Private Networking](#tag/Private-Networks)  The Private Networking API allows you to manage private networks / Virtual Private Clouds (VPC) for your Cloud VPS and VDS (please note that Storage VPS are not supported via API or CLI). Having a private network allows the associated instances to have a private and direct network connection. The traffic won't leave the data center and cannot be accessed by any other instance.  With this feature you can create multi layer systems, e.g. having a database server being only accessible from your application servers in one private network and keep the database replication in a second, separate network. This increases the speed as the traffic is NOT routed to the internet and also security as the traffic is within it's own secured VLAN.  Adding a Cloud VPS or VDS to a private network requires a reinstallation to make sure that all relevant parts for private networking are in place. When adding the same instance to another private network it will require a restart in order to make additional virtual network interface cards (NICs) available.  Please note that for each instance being part of one or several private networks a payed add-on is required. You can automatically purchase it via the Compute Management API.  ### [Secrets Management](#tag/Secrets)  You can optionally save your passwords or public ssh keys using the Secrets Management API. You are not required to use it there will be no functional disadvantages.  By using that API you can easily reuse you public ssh keys when setting up different servers without the need to look them up every time. It can also be used to allow Contabo Supporters to access your machine without sending the passwords via potentially unsecure emails.  ### [User Management](#tag/Users)  If you need to allow other persons or automation scripts to access specific API endpoints resp. resources the User Management API comes into play. With that API you are able to manage users having possibly restricted access. You are free to define those restrictions to fit your needs. So beside an arbitrary number of users you basically define any number of so called `roles`. Roles allows access and must be one of the following types:  * `apiPermission`                                                               This allows you to specify a restriction to certain functions of an API by allowing control over POST (=Create), GET (=Read), PUT/PATCH (=Update) and DELETE (=Delete) methods for each API endpoint (URL) individually. * `resourcePermission`                                                               In order to restrict access to specific resources create a role with `resourcePermission` type by specifying any number of [tags](#tag-management). These tags need to be assigned to resources for them to take effect. E.g. a tag could be assigned to several compute resources. So that a user with that role (and of course access to the API endpoints via `apiPermission` role type) could only access those compute resources.  The `roles` are then assigned to a `user`. You can assign one or several roles regardless of the role's type. Of course you could also assign a user `admin` privileges without specifying any roles.  ### [Tag Management](#tag/Tags)  The Tag Management API allows you to manage your tags in order to organize your resources in a more convenient way. Simply assign a tag to resources like a compute resource to manage them.The assignments of tags to resources will also enable you to control access to these specific resources to users via the [User Management API](#user-management). For convenience reasons you might choose a color for tag. The Customer Control Panel will use that color to display the tags.  ## Requests  The Contabo API supports HTTP requests like mentioned below. Not every endpoint supports all methods. The allowed methods are listed within this documentation.  Method | Description ---    | --- GET    | To retrieve information about a resource, use the GET method.<br>The data is returned as a JSON object. GET methods are read-only and do not affect any resources. POST   | Issue a POST method to create a new object. Include all needed attributes in the request body encoded as JSON. PATCH  | Some resources support partial modification with PATCH,<br>which modifies specific attributes without updating the entire object representation. PUT    | Use the PUT method to update information about a resource.<br>PUT will set new values on the item without regard to their current values. DELETE | Use the DELETE method to destroy a resource in your account.<br>If it is not found, the operation will return a 4xx error and an appropriate message.  ## Responses  Usually the Contabo API should respond to your requests. The data returned is in [JSON](https://www.json.org/) format allowing easy processing in any programming language or tools.  As common for HTTP requests you will get back a so called HTTP status code. This gives you overall information about success or error. The following table lists common HTTP status codes.  Please note that the description of the endpoints and methods are not listing all possibly status codes in detail as they are generic. Only special return codes with their resp. response data are explicitly listed.  Response Code | Description --- | --- 200 | The response contains your requested information. 201 | Your request was accepted. The resource was created. 204 | Your request succeeded, there is no additional information returned. 400 | Your request was malformed. 401 | You did not supply valid authentication credentials. 402 | Request refused as it requires additional payed service. 403 | You are not allowed to perform the request. 404 | No results were found for your request or resource does not exist. 409 | Conflict with resources. For example violation of unique data constraints detected when trying to create or change resources. 429 | Rate-limit reached. Please wait for some time before doing more requests. 500 | We were unable to perform the request due to server-side problems. In such cases please retry or contact the support.  Not every endpoint returns data. For example DELETE requests usually don't return any data. All others do return data. For easy handling the return values consists of metadata denoted with and underscore (\"_\") like `_links` or `_pagination`. The actual data is returned in a field called `data`. For convenience reasons this `data` field is always returned as an array even if it consists of only one single element.  Some general details about Contabo API from [Contabo](https://contabo.com).

API version: 1.0.0
Contact: support@contabo.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
)

// InstanceProduct struct for InstanceProduct
type InstanceProduct struct {
	// Tenant id
	TenantId string `json:"tenantId"`
	// Customer ID
	CustomerId string `json:"customerId"`
	// Instance ID
	InstanceId int64 `json:"instanceId"`
	// Product ID
	ProductId string `json:"productId"`
	// Product name
	Name string `json:"name"`
	// Short product name intended for display in the frontend. Falls back to the full product name if no dedicated frontend name is defined.
	FrontendName *string `json:"frontendName,omitempty"`
	// Instance's category depending on Product Id
	Category string `json:"category"`
	// RAM size in GB
	RamSizeGb float32 `json:"ramSizeGb"`
	// Disk size in GB
	DiskSizeGb float32 `json:"diskSizeGb"`
	// CPU core count
	CpuCores int64 `json:"cpuCores"`
	// Network speed in Mbit/s
	NetSpeed int64 `json:"netSpeed"`
	// Number of snapshots included in the product
	Snapshots int64 `json:"snapshots"`
	// Base virtual server price
	VsPrice float32 `json:"vsPrice"`
	// Additional price for Windows licensing
	WindowsPrice float32 `json:"windowsPrice"`
	// Price for automated backup service
	BackupPrice float32 `json:"backupPrice"`
	// Price for storage extension add-on
	StorageExtensionPrice float32 `json:"storageExtensionPrice"`
	// Additional fee applied for specific datacenter locations
	LocationFeePrice float32 `json:"locationFeePrice"`
	// Aggregated price for all active add-ons
	AddonsPrice float32 `json:"addonsPrice"`
	// Indicates whether a storage extension is attached
	HasStorageExtension bool `json:"hasStorageExtension"`
	// Indicates whether the automated backup add-on is attached
	HasBackupAddon *bool `json:"hasBackupAddon,omitempty"`
	// Indicates whether the Windows add-on is attached
	HasWindows *bool `json:"hasWindows,omitempty"`
	// Indicates whether a location fee add-on is attached
	HasLocationFee *bool `json:"hasLocationFee,omitempty"`
	// True if this product entry reflects the currently active subscription for the instance
	Current bool `json:"current"`
	// Identifier of the upgrade offer. Provide it as `offerId` when submitting the upgrade. Not set on the currently assigned product.
	OfferId *int64 `json:"offerId,omitempty"`
	// The offer is not available at the current location
	LocationChangeRequired *bool `json:"locationChangeRequired,omitempty"`
	// The upgrade can be performed while keeping the existing data (`migration` provisioning type)
	LiveMigrationAvailable *bool `json:"liveMigrationAvailable,omitempty"`
	// Reason why the existing data cannot be kept for this offer. Not set when it can be kept.
	LiveMigrationDisabledReason NullableString `json:"liveMigrationDisabledReason,omitempty"`
	// The storage extension add-on can be selected for this offer
	StorageExtensionAvailable *bool `json:"storageExtensionAvailable,omitempty"`
	// The automated backup add-on can be selected for this offer
	BackupAvailable *bool `json:"backupAvailable,omitempty"`
	// Upgrade discount percentage applied to this offer
	UpgradeDiscount *int64 `json:"upgradeDiscount,omitempty"`
	// Original gross price without add-ons and before the upgrade discount
	VsOriginalPrice *float32 `json:"vsOriginalPrice,omitempty"`
	// Confirmations the customer has to accept for this offer
	RequiredConfirmations *[]string `json:"requiredConfirmations,omitempty"`
	// Price difference charged when the storage extension is selected
	PriceDifferenceWithStorageExtension *float32 `json:"priceDifferenceWithStorageExtension,omitempty"`
	// Price difference charged when the storage extension is not selected
	PriceDifferenceWithoutStorageExtension *float32 `json:"priceDifferenceWithoutStorageExtension,omitempty"`
}

// NewInstanceProduct instantiates a new InstanceProduct object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewInstanceProduct(tenantId string, customerId string, instanceId int64, productId string, name string, category string, ramSizeGb float32, diskSizeGb float32, cpuCores int64, netSpeed int64, snapshots int64, vsPrice float32, windowsPrice float32, backupPrice float32, storageExtensionPrice float32, locationFeePrice float32, addonsPrice float32, hasStorageExtension bool, current bool) *InstanceProduct {
	this := InstanceProduct{}
	this.TenantId = tenantId
	this.CustomerId = customerId
	this.InstanceId = instanceId
	this.ProductId = productId
	this.Name = name
	this.Category = category
	this.RamSizeGb = ramSizeGb
	this.DiskSizeGb = diskSizeGb
	this.CpuCores = cpuCores
	this.NetSpeed = netSpeed
	this.Snapshots = snapshots
	this.VsPrice = vsPrice
	this.WindowsPrice = windowsPrice
	this.BackupPrice = backupPrice
	this.StorageExtensionPrice = storageExtensionPrice
	this.LocationFeePrice = locationFeePrice
	this.AddonsPrice = addonsPrice
	this.HasStorageExtension = hasStorageExtension
	this.Current = current
	return &this
}

// NewInstanceProductWithDefaults instantiates a new InstanceProduct object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewInstanceProductWithDefaults() *InstanceProduct {
	this := InstanceProduct{}
	return &this
}

// GetTenantId returns the TenantId field value
func (o *InstanceProduct) GetTenantId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TenantId
}

// GetTenantIdOk returns a tuple with the TenantId field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetTenantIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TenantId, true
}

// SetTenantId sets field value
func (o *InstanceProduct) SetTenantId(v string) {
	o.TenantId = v
}

// GetCustomerId returns the CustomerId field value
func (o *InstanceProduct) GetCustomerId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.CustomerId
}

// GetCustomerIdOk returns a tuple with the CustomerId field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetCustomerIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CustomerId, true
}

// SetCustomerId sets field value
func (o *InstanceProduct) SetCustomerId(v string) {
	o.CustomerId = v
}

// GetInstanceId returns the InstanceId field value
func (o *InstanceProduct) GetInstanceId() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.InstanceId
}

// GetInstanceIdOk returns a tuple with the InstanceId field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetInstanceIdOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.InstanceId, true
}

// SetInstanceId sets field value
func (o *InstanceProduct) SetInstanceId(v int64) {
	o.InstanceId = v
}

// GetProductId returns the ProductId field value
func (o *InstanceProduct) GetProductId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ProductId
}

// GetProductIdOk returns a tuple with the ProductId field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetProductIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ProductId, true
}

// SetProductId sets field value
func (o *InstanceProduct) SetProductId(v string) {
	o.ProductId = v
}

// GetName returns the Name field value
func (o *InstanceProduct) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *InstanceProduct) SetName(v string) {
	o.Name = v
}

// GetFrontendName returns the FrontendName field value if set, zero value otherwise.
func (o *InstanceProduct) GetFrontendName() string {
	if o == nil || o.FrontendName == nil {
		var ret string
		return ret
	}
	return *o.FrontendName
}

// GetFrontendNameOk returns a tuple with the FrontendName field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetFrontendNameOk() (*string, bool) {
	if o == nil || o.FrontendName == nil {
		return nil, false
	}
	return o.FrontendName, true
}

// HasFrontendName returns a boolean if a field has been set.
func (o *InstanceProduct) HasFrontendName() bool {
	if o != nil && o.FrontendName != nil {
		return true
	}

	return false
}

// SetFrontendName gets a reference to the given string and assigns it to the FrontendName field.
func (o *InstanceProduct) SetFrontendName(v string) {
	o.FrontendName = &v
}

// GetCategory returns the Category field value
func (o *InstanceProduct) GetCategory() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Category
}

// GetCategoryOk returns a tuple with the Category field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetCategoryOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Category, true
}

// SetCategory sets field value
func (o *InstanceProduct) SetCategory(v string) {
	o.Category = v
}

// GetRamSizeGb returns the RamSizeGb field value
func (o *InstanceProduct) GetRamSizeGb() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.RamSizeGb
}

// GetRamSizeGbOk returns a tuple with the RamSizeGb field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetRamSizeGbOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.RamSizeGb, true
}

// SetRamSizeGb sets field value
func (o *InstanceProduct) SetRamSizeGb(v float32) {
	o.RamSizeGb = v
}

// GetDiskSizeGb returns the DiskSizeGb field value
func (o *InstanceProduct) GetDiskSizeGb() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.DiskSizeGb
}

// GetDiskSizeGbOk returns a tuple with the DiskSizeGb field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetDiskSizeGbOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.DiskSizeGb, true
}

// SetDiskSizeGb sets field value
func (o *InstanceProduct) SetDiskSizeGb(v float32) {
	o.DiskSizeGb = v
}

// GetCpuCores returns the CpuCores field value
func (o *InstanceProduct) GetCpuCores() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.CpuCores
}

// GetCpuCoresOk returns a tuple with the CpuCores field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetCpuCoresOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CpuCores, true
}

// SetCpuCores sets field value
func (o *InstanceProduct) SetCpuCores(v int64) {
	o.CpuCores = v
}

// GetNetSpeed returns the NetSpeed field value
func (o *InstanceProduct) GetNetSpeed() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.NetSpeed
}

// GetNetSpeedOk returns a tuple with the NetSpeed field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetNetSpeedOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.NetSpeed, true
}

// SetNetSpeed sets field value
func (o *InstanceProduct) SetNetSpeed(v int64) {
	o.NetSpeed = v
}

// GetSnapshots returns the Snapshots field value
func (o *InstanceProduct) GetSnapshots() int64 {
	if o == nil {
		var ret int64
		return ret
	}

	return o.Snapshots
}

// GetSnapshotsOk returns a tuple with the Snapshots field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetSnapshotsOk() (*int64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Snapshots, true
}

// SetSnapshots sets field value
func (o *InstanceProduct) SetSnapshots(v int64) {
	o.Snapshots = v
}

// GetVsPrice returns the VsPrice field value
func (o *InstanceProduct) GetVsPrice() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.VsPrice
}

// GetVsPriceOk returns a tuple with the VsPrice field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetVsPriceOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.VsPrice, true
}

// SetVsPrice sets field value
func (o *InstanceProduct) SetVsPrice(v float32) {
	o.VsPrice = v
}

// GetWindowsPrice returns the WindowsPrice field value
func (o *InstanceProduct) GetWindowsPrice() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.WindowsPrice
}

// GetWindowsPriceOk returns a tuple with the WindowsPrice field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetWindowsPriceOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.WindowsPrice, true
}

// SetWindowsPrice sets field value
func (o *InstanceProduct) SetWindowsPrice(v float32) {
	o.WindowsPrice = v
}

// GetBackupPrice returns the BackupPrice field value
func (o *InstanceProduct) GetBackupPrice() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.BackupPrice
}

// GetBackupPriceOk returns a tuple with the BackupPrice field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetBackupPriceOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.BackupPrice, true
}

// SetBackupPrice sets field value
func (o *InstanceProduct) SetBackupPrice(v float32) {
	o.BackupPrice = v
}

// GetStorageExtensionPrice returns the StorageExtensionPrice field value
func (o *InstanceProduct) GetStorageExtensionPrice() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.StorageExtensionPrice
}

// GetStorageExtensionPriceOk returns a tuple with the StorageExtensionPrice field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetStorageExtensionPriceOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.StorageExtensionPrice, true
}

// SetStorageExtensionPrice sets field value
func (o *InstanceProduct) SetStorageExtensionPrice(v float32) {
	o.StorageExtensionPrice = v
}

// GetLocationFeePrice returns the LocationFeePrice field value
func (o *InstanceProduct) GetLocationFeePrice() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.LocationFeePrice
}

// GetLocationFeePriceOk returns a tuple with the LocationFeePrice field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetLocationFeePriceOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LocationFeePrice, true
}

// SetLocationFeePrice sets field value
func (o *InstanceProduct) SetLocationFeePrice(v float32) {
	o.LocationFeePrice = v
}

// GetAddonsPrice returns the AddonsPrice field value
func (o *InstanceProduct) GetAddonsPrice() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.AddonsPrice
}

// GetAddonsPriceOk returns a tuple with the AddonsPrice field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetAddonsPriceOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.AddonsPrice, true
}

// SetAddonsPrice sets field value
func (o *InstanceProduct) SetAddonsPrice(v float32) {
	o.AddonsPrice = v
}

// GetHasStorageExtension returns the HasStorageExtension field value
func (o *InstanceProduct) GetHasStorageExtension() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.HasStorageExtension
}

// GetHasStorageExtensionOk returns a tuple with the HasStorageExtension field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetHasStorageExtensionOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.HasStorageExtension, true
}

// SetHasStorageExtension sets field value
func (o *InstanceProduct) SetHasStorageExtension(v bool) {
	o.HasStorageExtension = v
}

// GetHasBackupAddon returns the HasBackupAddon field value if set, zero value otherwise.
func (o *InstanceProduct) GetHasBackupAddon() bool {
	if o == nil || o.HasBackupAddon == nil {
		var ret bool
		return ret
	}
	return *o.HasBackupAddon
}

// GetHasBackupAddonOk returns a tuple with the HasBackupAddon field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetHasBackupAddonOk() (*bool, bool) {
	if o == nil || o.HasBackupAddon == nil {
		return nil, false
	}
	return o.HasBackupAddon, true
}

// HasHasBackupAddon returns a boolean if a field has been set.
func (o *InstanceProduct) HasHasBackupAddon() bool {
	if o != nil && o.HasBackupAddon != nil {
		return true
	}

	return false
}

// SetHasBackupAddon gets a reference to the given bool and assigns it to the HasBackupAddon field.
func (o *InstanceProduct) SetHasBackupAddon(v bool) {
	o.HasBackupAddon = &v
}

// GetHasWindows returns the HasWindows field value if set, zero value otherwise.
func (o *InstanceProduct) GetHasWindows() bool {
	if o == nil || o.HasWindows == nil {
		var ret bool
		return ret
	}
	return *o.HasWindows
}

// GetHasWindowsOk returns a tuple with the HasWindows field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetHasWindowsOk() (*bool, bool) {
	if o == nil || o.HasWindows == nil {
		return nil, false
	}
	return o.HasWindows, true
}

// HasHasWindows returns a boolean if a field has been set.
func (o *InstanceProduct) HasHasWindows() bool {
	if o != nil && o.HasWindows != nil {
		return true
	}

	return false
}

// SetHasWindows gets a reference to the given bool and assigns it to the HasWindows field.
func (o *InstanceProduct) SetHasWindows(v bool) {
	o.HasWindows = &v
}

// GetHasLocationFee returns the HasLocationFee field value if set, zero value otherwise.
func (o *InstanceProduct) GetHasLocationFee() bool {
	if o == nil || o.HasLocationFee == nil {
		var ret bool
		return ret
	}
	return *o.HasLocationFee
}

// GetHasLocationFeeOk returns a tuple with the HasLocationFee field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetHasLocationFeeOk() (*bool, bool) {
	if o == nil || o.HasLocationFee == nil {
		return nil, false
	}
	return o.HasLocationFee, true
}

// HasHasLocationFee returns a boolean if a field has been set.
func (o *InstanceProduct) HasHasLocationFee() bool {
	if o != nil && o.HasLocationFee != nil {
		return true
	}

	return false
}

// SetHasLocationFee gets a reference to the given bool and assigns it to the HasLocationFee field.
func (o *InstanceProduct) SetHasLocationFee(v bool) {
	o.HasLocationFee = &v
}

// GetCurrent returns the Current field value
func (o *InstanceProduct) GetCurrent() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.Current
}

// GetCurrentOk returns a tuple with the Current field value
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetCurrentOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Current, true
}

// SetCurrent sets field value
func (o *InstanceProduct) SetCurrent(v bool) {
	o.Current = v
}

// GetOfferId returns the OfferId field value if set, zero value otherwise.
func (o *InstanceProduct) GetOfferId() int64 {
	if o == nil || o.OfferId == nil {
		var ret int64
		return ret
	}
	return *o.OfferId
}

// GetOfferIdOk returns a tuple with the OfferId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetOfferIdOk() (*int64, bool) {
	if o == nil || o.OfferId == nil {
		return nil, false
	}
	return o.OfferId, true
}

// HasOfferId returns a boolean if a field has been set.
func (o *InstanceProduct) HasOfferId() bool {
	if o != nil && o.OfferId != nil {
		return true
	}

	return false
}

// SetOfferId gets a reference to the given int64 and assigns it to the OfferId field.
func (o *InstanceProduct) SetOfferId(v int64) {
	o.OfferId = &v
}

// GetLocationChangeRequired returns the LocationChangeRequired field value if set, zero value otherwise.
func (o *InstanceProduct) GetLocationChangeRequired() bool {
	if o == nil || o.LocationChangeRequired == nil {
		var ret bool
		return ret
	}
	return *o.LocationChangeRequired
}

// GetLocationChangeRequiredOk returns a tuple with the LocationChangeRequired field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetLocationChangeRequiredOk() (*bool, bool) {
	if o == nil || o.LocationChangeRequired == nil {
		return nil, false
	}
	return o.LocationChangeRequired, true
}

// HasLocationChangeRequired returns a boolean if a field has been set.
func (o *InstanceProduct) HasLocationChangeRequired() bool {
	if o != nil && o.LocationChangeRequired != nil {
		return true
	}

	return false
}

// SetLocationChangeRequired gets a reference to the given bool and assigns it to the LocationChangeRequired field.
func (o *InstanceProduct) SetLocationChangeRequired(v bool) {
	o.LocationChangeRequired = &v
}

// GetLiveMigrationAvailable returns the LiveMigrationAvailable field value if set, zero value otherwise.
func (o *InstanceProduct) GetLiveMigrationAvailable() bool {
	if o == nil || o.LiveMigrationAvailable == nil {
		var ret bool
		return ret
	}
	return *o.LiveMigrationAvailable
}

// GetLiveMigrationAvailableOk returns a tuple with the LiveMigrationAvailable field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetLiveMigrationAvailableOk() (*bool, bool) {
	if o == nil || o.LiveMigrationAvailable == nil {
		return nil, false
	}
	return o.LiveMigrationAvailable, true
}

// HasLiveMigrationAvailable returns a boolean if a field has been set.
func (o *InstanceProduct) HasLiveMigrationAvailable() bool {
	if o != nil && o.LiveMigrationAvailable != nil {
		return true
	}

	return false
}

// SetLiveMigrationAvailable gets a reference to the given bool and assigns it to the LiveMigrationAvailable field.
func (o *InstanceProduct) SetLiveMigrationAvailable(v bool) {
	o.LiveMigrationAvailable = &v
}

// GetLiveMigrationDisabledReason returns the LiveMigrationDisabledReason field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *InstanceProduct) GetLiveMigrationDisabledReason() string {
	if o == nil || o.LiveMigrationDisabledReason.Get() == nil {
		var ret string
		return ret
	}
	return *o.LiveMigrationDisabledReason.Get()
}

// GetLiveMigrationDisabledReasonOk returns a tuple with the LiveMigrationDisabledReason field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *InstanceProduct) GetLiveMigrationDisabledReasonOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return o.LiveMigrationDisabledReason.Get(), o.LiveMigrationDisabledReason.IsSet()
}

// HasLiveMigrationDisabledReason returns a boolean if a field has been set.
func (o *InstanceProduct) HasLiveMigrationDisabledReason() bool {
	if o != nil && o.LiveMigrationDisabledReason.IsSet() {
		return true
	}

	return false
}

// SetLiveMigrationDisabledReason gets a reference to the given NullableString and assigns it to the LiveMigrationDisabledReason field.
func (o *InstanceProduct) SetLiveMigrationDisabledReason(v string) {
	o.LiveMigrationDisabledReason.Set(&v)
}

// SetLiveMigrationDisabledReasonNil sets the value for LiveMigrationDisabledReason to be an explicit nil
func (o *InstanceProduct) SetLiveMigrationDisabledReasonNil() {
	o.LiveMigrationDisabledReason.Set(nil)
}

// UnsetLiveMigrationDisabledReason ensures that no value is present for LiveMigrationDisabledReason, not even an explicit nil
func (o *InstanceProduct) UnsetLiveMigrationDisabledReason() {
	o.LiveMigrationDisabledReason.Unset()
}

// GetStorageExtensionAvailable returns the StorageExtensionAvailable field value if set, zero value otherwise.
func (o *InstanceProduct) GetStorageExtensionAvailable() bool {
	if o == nil || o.StorageExtensionAvailable == nil {
		var ret bool
		return ret
	}
	return *o.StorageExtensionAvailable
}

// GetStorageExtensionAvailableOk returns a tuple with the StorageExtensionAvailable field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetStorageExtensionAvailableOk() (*bool, bool) {
	if o == nil || o.StorageExtensionAvailable == nil {
		return nil, false
	}
	return o.StorageExtensionAvailable, true
}

// HasStorageExtensionAvailable returns a boolean if a field has been set.
func (o *InstanceProduct) HasStorageExtensionAvailable() bool {
	if o != nil && o.StorageExtensionAvailable != nil {
		return true
	}

	return false
}

// SetStorageExtensionAvailable gets a reference to the given bool and assigns it to the StorageExtensionAvailable field.
func (o *InstanceProduct) SetStorageExtensionAvailable(v bool) {
	o.StorageExtensionAvailable = &v
}

// GetBackupAvailable returns the BackupAvailable field value if set, zero value otherwise.
func (o *InstanceProduct) GetBackupAvailable() bool {
	if o == nil || o.BackupAvailable == nil {
		var ret bool
		return ret
	}
	return *o.BackupAvailable
}

// GetBackupAvailableOk returns a tuple with the BackupAvailable field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetBackupAvailableOk() (*bool, bool) {
	if o == nil || o.BackupAvailable == nil {
		return nil, false
	}
	return o.BackupAvailable, true
}

// HasBackupAvailable returns a boolean if a field has been set.
func (o *InstanceProduct) HasBackupAvailable() bool {
	if o != nil && o.BackupAvailable != nil {
		return true
	}

	return false
}

// SetBackupAvailable gets a reference to the given bool and assigns it to the BackupAvailable field.
func (o *InstanceProduct) SetBackupAvailable(v bool) {
	o.BackupAvailable = &v
}

// GetUpgradeDiscount returns the UpgradeDiscount field value if set, zero value otherwise.
func (o *InstanceProduct) GetUpgradeDiscount() int64 {
	if o == nil || o.UpgradeDiscount == nil {
		var ret int64
		return ret
	}
	return *o.UpgradeDiscount
}

// GetUpgradeDiscountOk returns a tuple with the UpgradeDiscount field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetUpgradeDiscountOk() (*int64, bool) {
	if o == nil || o.UpgradeDiscount == nil {
		return nil, false
	}
	return o.UpgradeDiscount, true
}

// HasUpgradeDiscount returns a boolean if a field has been set.
func (o *InstanceProduct) HasUpgradeDiscount() bool {
	if o != nil && o.UpgradeDiscount != nil {
		return true
	}

	return false
}

// SetUpgradeDiscount gets a reference to the given int64 and assigns it to the UpgradeDiscount field.
func (o *InstanceProduct) SetUpgradeDiscount(v int64) {
	o.UpgradeDiscount = &v
}

// GetVsOriginalPrice returns the VsOriginalPrice field value if set, zero value otherwise.
func (o *InstanceProduct) GetVsOriginalPrice() float32 {
	if o == nil || o.VsOriginalPrice == nil {
		var ret float32
		return ret
	}
	return *o.VsOriginalPrice
}

// GetVsOriginalPriceOk returns a tuple with the VsOriginalPrice field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetVsOriginalPriceOk() (*float32, bool) {
	if o == nil || o.VsOriginalPrice == nil {
		return nil, false
	}
	return o.VsOriginalPrice, true
}

// HasVsOriginalPrice returns a boolean if a field has been set.
func (o *InstanceProduct) HasVsOriginalPrice() bool {
	if o != nil && o.VsOriginalPrice != nil {
		return true
	}

	return false
}

// SetVsOriginalPrice gets a reference to the given float32 and assigns it to the VsOriginalPrice field.
func (o *InstanceProduct) SetVsOriginalPrice(v float32) {
	o.VsOriginalPrice = &v
}

// GetRequiredConfirmations returns the RequiredConfirmations field value if set, zero value otherwise.
func (o *InstanceProduct) GetRequiredConfirmations() []string {
	if o == nil || o.RequiredConfirmations == nil {
		var ret []string
		return ret
	}
	return *o.RequiredConfirmations
}

// GetRequiredConfirmationsOk returns a tuple with the RequiredConfirmations field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetRequiredConfirmationsOk() (*[]string, bool) {
	if o == nil || o.RequiredConfirmations == nil {
		return nil, false
	}
	return o.RequiredConfirmations, true
}

// HasRequiredConfirmations returns a boolean if a field has been set.
func (o *InstanceProduct) HasRequiredConfirmations() bool {
	if o != nil && o.RequiredConfirmations != nil {
		return true
	}

	return false
}

// SetRequiredConfirmations gets a reference to the given []string and assigns it to the RequiredConfirmations field.
func (o *InstanceProduct) SetRequiredConfirmations(v []string) {
	o.RequiredConfirmations = &v
}

// GetPriceDifferenceWithStorageExtension returns the PriceDifferenceWithStorageExtension field value if set, zero value otherwise.
func (o *InstanceProduct) GetPriceDifferenceWithStorageExtension() float32 {
	if o == nil || o.PriceDifferenceWithStorageExtension == nil {
		var ret float32
		return ret
	}
	return *o.PriceDifferenceWithStorageExtension
}

// GetPriceDifferenceWithStorageExtensionOk returns a tuple with the PriceDifferenceWithStorageExtension field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetPriceDifferenceWithStorageExtensionOk() (*float32, bool) {
	if o == nil || o.PriceDifferenceWithStorageExtension == nil {
		return nil, false
	}
	return o.PriceDifferenceWithStorageExtension, true
}

// HasPriceDifferenceWithStorageExtension returns a boolean if a field has been set.
func (o *InstanceProduct) HasPriceDifferenceWithStorageExtension() bool {
	if o != nil && o.PriceDifferenceWithStorageExtension != nil {
		return true
	}

	return false
}

// SetPriceDifferenceWithStorageExtension gets a reference to the given float32 and assigns it to the PriceDifferenceWithStorageExtension field.
func (o *InstanceProduct) SetPriceDifferenceWithStorageExtension(v float32) {
	o.PriceDifferenceWithStorageExtension = &v
}

// GetPriceDifferenceWithoutStorageExtension returns the PriceDifferenceWithoutStorageExtension field value if set, zero value otherwise.
func (o *InstanceProduct) GetPriceDifferenceWithoutStorageExtension() float32 {
	if o == nil || o.PriceDifferenceWithoutStorageExtension == nil {
		var ret float32
		return ret
	}
	return *o.PriceDifferenceWithoutStorageExtension
}

// GetPriceDifferenceWithoutStorageExtensionOk returns a tuple with the PriceDifferenceWithoutStorageExtension field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *InstanceProduct) GetPriceDifferenceWithoutStorageExtensionOk() (*float32, bool) {
	if o == nil || o.PriceDifferenceWithoutStorageExtension == nil {
		return nil, false
	}
	return o.PriceDifferenceWithoutStorageExtension, true
}

// HasPriceDifferenceWithoutStorageExtension returns a boolean if a field has been set.
func (o *InstanceProduct) HasPriceDifferenceWithoutStorageExtension() bool {
	if o != nil && o.PriceDifferenceWithoutStorageExtension != nil {
		return true
	}

	return false
}

// SetPriceDifferenceWithoutStorageExtension gets a reference to the given float32 and assigns it to the PriceDifferenceWithoutStorageExtension field.
func (o *InstanceProduct) SetPriceDifferenceWithoutStorageExtension(v float32) {
	o.PriceDifferenceWithoutStorageExtension = &v
}

func (o InstanceProduct) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["tenantId"] = o.TenantId
	}
	if true {
		toSerialize["customerId"] = o.CustomerId
	}
	if true {
		toSerialize["instanceId"] = o.InstanceId
	}
	if true {
		toSerialize["productId"] = o.ProductId
	}
	if true {
		toSerialize["name"] = o.Name
	}
	if o.FrontendName != nil {
		toSerialize["frontendName"] = o.FrontendName
	}
	if true {
		toSerialize["category"] = o.Category
	}
	if true {
		toSerialize["ramSizeGb"] = o.RamSizeGb
	}
	if true {
		toSerialize["diskSizeGb"] = o.DiskSizeGb
	}
	if true {
		toSerialize["cpuCores"] = o.CpuCores
	}
	if true {
		toSerialize["netSpeed"] = o.NetSpeed
	}
	if true {
		toSerialize["snapshots"] = o.Snapshots
	}
	if true {
		toSerialize["vsPrice"] = o.VsPrice
	}
	if true {
		toSerialize["windowsPrice"] = o.WindowsPrice
	}
	if true {
		toSerialize["backupPrice"] = o.BackupPrice
	}
	if true {
		toSerialize["storageExtensionPrice"] = o.StorageExtensionPrice
	}
	if true {
		toSerialize["locationFeePrice"] = o.LocationFeePrice
	}
	if true {
		toSerialize["addonsPrice"] = o.AddonsPrice
	}
	if true {
		toSerialize["hasStorageExtension"] = o.HasStorageExtension
	}
	if o.HasBackupAddon != nil {
		toSerialize["hasBackupAddon"] = o.HasBackupAddon
	}
	if o.HasWindows != nil {
		toSerialize["hasWindows"] = o.HasWindows
	}
	if o.HasLocationFee != nil {
		toSerialize["hasLocationFee"] = o.HasLocationFee
	}
	if true {
		toSerialize["current"] = o.Current
	}
	if o.OfferId != nil {
		toSerialize["offerId"] = o.OfferId
	}
	if o.LocationChangeRequired != nil {
		toSerialize["locationChangeRequired"] = o.LocationChangeRequired
	}
	if o.LiveMigrationAvailable != nil {
		toSerialize["liveMigrationAvailable"] = o.LiveMigrationAvailable
	}
	if o.LiveMigrationDisabledReason.IsSet() {
		toSerialize["liveMigrationDisabledReason"] = o.LiveMigrationDisabledReason.Get()
	}
	if o.StorageExtensionAvailable != nil {
		toSerialize["storageExtensionAvailable"] = o.StorageExtensionAvailable
	}
	if o.BackupAvailable != nil {
		toSerialize["backupAvailable"] = o.BackupAvailable
	}
	if o.UpgradeDiscount != nil {
		toSerialize["upgradeDiscount"] = o.UpgradeDiscount
	}
	if o.VsOriginalPrice != nil {
		toSerialize["vsOriginalPrice"] = o.VsOriginalPrice
	}
	if o.RequiredConfirmations != nil {
		toSerialize["requiredConfirmations"] = o.RequiredConfirmations
	}
	if o.PriceDifferenceWithStorageExtension != nil {
		toSerialize["priceDifferenceWithStorageExtension"] = o.PriceDifferenceWithStorageExtension
	}
	if o.PriceDifferenceWithoutStorageExtension != nil {
		toSerialize["priceDifferenceWithoutStorageExtension"] = o.PriceDifferenceWithoutStorageExtension
	}
	return json.Marshal(toSerialize)
}

type NullableInstanceProduct struct {
	value *InstanceProduct
	isSet bool
}

func (v NullableInstanceProduct) Get() *InstanceProduct {
	return v.value
}

func (v *NullableInstanceProduct) Set(val *InstanceProduct) {
	v.value = val
	v.isSet = true
}

func (v NullableInstanceProduct) IsSet() bool {
	return v.isSet
}

func (v *NullableInstanceProduct) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableInstanceProduct(val *InstanceProduct) *NullableInstanceProduct {
	return &NullableInstanceProduct{value: val, isSet: true}
}

func (v NullableInstanceProduct) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableInstanceProduct) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
