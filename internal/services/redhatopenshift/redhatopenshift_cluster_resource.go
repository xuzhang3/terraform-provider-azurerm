package redhatopenshift

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/redhatopenshift/mgmt/2022-04-01/redhatopenshift"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/redhatopenshift/parse"
	openShiftValidate "github.com/hashicorp/terraform-provider-azurerm/internal/services/redhatopenshift/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tags"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

var randomDomainName = GenerateRandomDomainName()

func resourceOpenShiftCluster() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceOpenShiftClusterCreate,
		Read:   resourceOpenShiftClusterRead,
		Update: resourceOpenShiftClusterUpdate,
		Delete: resourceOpenShiftClusterDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.RedhatOpenShiftClusterID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(90 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(90 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"location": commonschema.Location(),

			"resource_group_name": commonschema.ResourceGroupName(),

			"service_principal": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"client_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: openShiftValidate.ClientID,
						},
						"client_secret": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"cluster_profile": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"pull_secret": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"domain": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"fips_enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"network_profile": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"pod_cidr": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validate.CIDR,
						},
						"service_cidr": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validate.CIDR,
						},
					},
				},
			},

			"main_profile": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"subnet_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
						"vm_size": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validation.StringIsNotEmpty,
						},
						"encryption_at_host_enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							Default:  false,
						},
						"disk_encryption_set_id": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
					},
				},
			},

			"worker_profile": {
				Type:     pluginsdk.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"vm_size": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validation.StringIsNotEmpty,
						},
						"disk_size_gb": {
							Type:         pluginsdk.TypeInt,
							Required:     true,
							ValidateFunc: openShiftValidate.DiskSizeGB,
						},
						"node_count": {
							Type:     pluginsdk.TypeInt,
							Required: true,
						},
						"subnet_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},

						"encryption_at_host_enabled": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							Default:  false,
						},
						"disk_encryption_set_id": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
					},
				},
			},

			"api_server_profile": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"visibility": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(redhatopenshift.VisibilityPublic),
								string(redhatopenshift.VisibilityPrivate),
							}, false),
						},
					},
				},
			},

			"ingress_profile": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				MaxItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"visibility": {
							Type:     pluginsdk.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								string(redhatopenshift.VisibilityPublic),
								string(redhatopenshift.VisibilityPrivate),
							}, false),
						},
					},
				},
			},

			"tags": tags.Schema(),

			"version": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},

			"console_url": {
				Type:     pluginsdk.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceOpenShiftClusterCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).RedHatOpenshift.OpenShiftClustersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroupName := d.Get("resource_group_name").(string)
	name := d.Get("name").(string)

	id := parse.NewRedhatOpenShiftClusterID(subscriptionId, resourceGroupName, name)

	existing, err := client.Get(ctx, resourceGroupName, name)
	if err != nil && !utils.ResponseWasNotFound(existing.Response) {
		return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
	}

	if existing.ID != nil && *existing.ID != "" {
		return tf.ImportAsExistsError("azurerm_redhatopenshift_cluster", *existing.ID)
	}

	location := azure.NormalizeLocation(d.Get("location").(string))

	parameters := redhatopenshift.OpenShiftCluster{
		Name:     &name,
		Location: &location,
		OpenShiftClusterProperties: &redhatopenshift.OpenShiftClusterProperties{
			ClusterProfile:          expandOpenshiftClusterProfile(d.Get("cluster_profile").([]interface{}), subscriptionId),
			ConsoleProfile:          &redhatopenshift.ConsoleProfile{},
			ServicePrincipalProfile: expandOpenshiftServicePrincipalProfile(d.Get("service_principal").([]interface{})),
			NetworkProfile:          expandOpenshiftNetworkProfile(d.Get("network_profile").([]interface{})),
			MasterProfile:           expandOpenshiftMasterProfile(d.Get("main_profile").([]interface{})),
			WorkerProfiles:          expandOpenshiftWorkerProfiles(d.Get("worker_profile").([]interface{})),
			ApiserverProfile:        expandOpenshiftApiServerProfile(d.Get("api_server_profile").([]interface{})),
			IngressProfiles:         expandOpenshiftIngressProfiles(d.Get("ingress_profile").([]interface{})),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	js, _ := json.Marshal(parameters) //TODO
	log.Printf(" create DDDDDD %s: ", js)
	future, err := client.CreateOrUpdate(ctx, resourceGroupName, name, parameters)
	if err != nil {
		return fmt.Errorf("creating %s: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of %s: %+v", id, err)
	}

	d.SetId(id.ID())
	return resourceOpenShiftClusterRead(d, meta)
}

func resourceOpenShiftClusterUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).RedHatOpenshift.OpenShiftClustersClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.RedhatOpenShiftClusterID(d.Id())
	if err != nil {
		return err
	}

	parameter := redhatopenshift.OpenShiftClusterUpdate{
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if d.HasChange("cluster_profile") {
		clusterProfileRaw := d.Get("cluster_profile").([]interface{})
		clusterProfile := expandOpenshiftClusterProfile(clusterProfileRaw, subscriptionId)
		parameter.OpenShiftClusterProperties.ClusterProfile = clusterProfile
	}

	if d.HasChange("main_profile") {
		mainProfileRaw := d.Get("main_profile").([]interface{})
		mainProfile := expandOpenshiftMasterProfile(mainProfileRaw)
		parameter.OpenShiftClusterProperties.MasterProfile = mainProfile
	}

	if d.HasChange("worker_profile") {
		workerProfilesRaw := d.Get("worker_profile").([]interface{})
		workerProfiles := expandOpenshiftWorkerProfiles(workerProfilesRaw)
		parameter.OpenShiftClusterProperties.WorkerProfiles = workerProfiles
	}

	future, err := client.Update(ctx, id.ResourceGroup, id.OpenShiftClusterName, parameter)
	if err != nil {
		return fmt.Errorf("updating %s: %+v", id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for updationg of %s: %+v", id, err)
	}

	return resourceOpenShiftClusterRead(d, meta)
}

func resourceOpenShiftClusterRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).RedHatOpenshift.OpenShiftClustersClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.RedhatOpenShiftClusterID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.OpenShiftClusterName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving %s: %+v", *id, err)
	}

	js, _ := json.Marshal(resp)
	log.Printf(" read DDDDDD %s: ", js)

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.OpenShiftClusterProperties; props != nil {
		clusterProfile := flattenOpenShiftClusterProfile(props.ClusterProfile)
		if err := d.Set("cluster_profile", clusterProfile); err != nil {
			return fmt.Errorf("setting `cluster_profile`: %+v", err)
		}

		servicePrincipalProfile := flattenOpenShiftServicePrincipalProfile(props.ServicePrincipalProfile, d)
		if err := d.Set("service_principal", servicePrincipalProfile); err != nil {
			return fmt.Errorf("setting `service_principal`: %+v", err)
		}

		networkProfile := flattenOpenShiftNetworkProfile(props.NetworkProfile)
		if err := d.Set("network_profile", networkProfile); err != nil {
			return fmt.Errorf("setting `network_profile`: %+v", err)
		}

		mainProfile := flattenOpenShiftMasterProfile(props.MasterProfile)
		if err := d.Set("main_profile", mainProfile); err != nil {
			return fmt.Errorf("setting `main_profile`: %+v", err)
		}

		workerProfiles := flattenOpenShiftWorkerProfiles(props.WorkerProfiles)
		if err := d.Set("worker_profile", workerProfiles); err != nil {
			return fmt.Errorf("setting `worker_profile`: %+v", err)
		}

		apiServerProfile := flattenOpenShiftAPIServerProfile(props.ApiserverProfile)
		if err := d.Set("api_server_profile", apiServerProfile); err != nil {
			return fmt.Errorf("setting `api_server_profile`: %+v", err)
		}

		ingressProfiles := flattenOpenShiftIngressProfiles(props.IngressProfiles)
		if err := d.Set("ingress_profile", ingressProfiles); err != nil {
			return fmt.Errorf("setting `ingress_profile`: %+v", err)
		}

		if props.ClusterProfile != nil && props.ClusterProfile.Version != nil {
			d.Set("version", props.ClusterProfile.Version)
		}

		if props.ConsoleProfile != nil && props.ConsoleProfile.URL != nil {
			d.Set("console_url", props.ConsoleProfile.URL)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceOpenShiftClusterDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).RedHatOpenshift.OpenShiftClustersClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.RedhatOpenShiftClusterID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.OpenShiftClusterName)
	if err != nil {
		return fmt.Errorf("deleting %s: %+v", *id, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for deletion of %s: %+v", *id, err)
	}

	return nil
}

func flattenOpenShiftClusterProfile(profile *redhatopenshift.ClusterProfile) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	pullSecret := ""
	if profile.PullSecret != nil {
		pullSecret = *profile.PullSecret
	}

	clusterDomain := ""
	if profile.Domain != nil {
		clusterDomain = *profile.Domain
	}

	fipsEnabled := profile.FipsValidatedModules == redhatopenshift.FipsValidatedModulesEnabled

	return []interface{}{
		map[string]interface{}{
			"pull_secret":  pullSecret,
			"domain":       clusterDomain,
			"fips_enabled": fipsEnabled,
		},
	}
}

func flattenOpenShiftServicePrincipalProfile(profile *redhatopenshift.ServicePrincipalProfile, d *pluginsdk.ResourceData) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	clientID := ""
	if profile.ClientID != nil {
		clientID = *profile.ClientID
	}

	clientSecret := ""
	if sp, ok := d.GetOk("service_principal"); ok {
		val := sp.([]interface{})

		if len(val) > 0 && val[0] != nil {
			raw := val[0].(map[string]interface{})
			clientSecret = raw["client_secret"].(string)
		}
	}

	return []interface{}{
		map[string]interface{}{
			"client_id":     clientID,
			"client_secret": clientSecret,
		},
	}
}

func flattenOpenShiftNetworkProfile(profile *redhatopenshift.NetworkProfile) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	podCidr := ""
	if profile.PodCidr != nil {
		podCidr = *profile.PodCidr
	}

	serviceCidr := ""
	if profile.ServiceCidr != nil {
		serviceCidr = *profile.ServiceCidr
	}

	return []interface{}{
		map[string]interface{}{
			"pod_cidr":     podCidr,
			"service_cidr": serviceCidr,
		},
	}
}

func flattenOpenShiftMasterProfile(profile *redhatopenshift.MasterProfile) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	subnetId := ""
	if profile.SubnetID != nil {
		subnetId = *profile.SubnetID
	}

	encryptionAtHostEnabled := profile.EncryptionAtHost == redhatopenshift.EncryptionAtHostEnabled

	diskEncryptionSetId := ""
	if profile.DiskEncryptionSetID != nil {
		diskEncryptionSetId = *profile.DiskEncryptionSetID
	}

	return []interface{}{
		map[string]interface{}{
			"vm_size":                    profile.VMSize,
			"subnet_id":                  subnetId,
			"encryption_at_host_enabled": encryptionAtHostEnabled,
			"disk_encryption_set_id":     diskEncryptionSetId,
		},
	}
}

func flattenOpenShiftWorkerProfiles(profiles *[]redhatopenshift.WorkerProfile) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)

	for _, profile := range *profiles {
		results = append(results, map[string]interface{}{
			"disk_size_gb":               profile.DiskSizeGB,
			"node_count":                 profile.Count,
			"vm_size":                    profile.VMSize,
			"subnet_id":                  profile.SubnetID,
			"encryption_at_host_enabled": profile.EncryptionAtHost == redhatopenshift.EncryptionAtHostEnabled,
			"disk_encryption_set_id":     profile.DiskEncryptionSetID,
		})
	}
	return results
}

func flattenOpenShiftAPIServerProfile(profile *redhatopenshift.APIServerProfile) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"visibility": string(profile.Visibility),
		},
	}
}

func flattenOpenShiftIngressProfiles(profiles *[]redhatopenshift.IngressProfile) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)

	for _, profile := range *profiles {
		result := make(map[string]interface{})
		result["visibility"] = string(profile.Visibility)

		results = append(results, result)
	}

	return results
}

func expandOpenshiftClusterProfile(input []interface{}, subscriptionId string) *redhatopenshift.ClusterProfile {
	resourceGroupName := fmt.Sprintf("aro-%s", randomDomainName)
	resourceGroupId := ResourceGroupID(subscriptionId, resourceGroupName)

	if len(input) == 0 {
		return &redhatopenshift.ClusterProfile{
			ResourceGroupID:      utils.String(resourceGroupId),
			Domain:               utils.String(randomDomainName),
			FipsValidatedModules: redhatopenshift.FipsValidatedModulesDisabled,
		}
	}

	config := input[0].(map[string]interface{})

	pullSecret := config["pull_secret"].(string)

	domain := config["domain"].(string)
	if domain == "" {
		domain = randomDomainName
	}

	fipsValidatedModules := redhatopenshift.FipsValidatedModulesDisabled
	fipsEnabled := config["fips_enabled"].(bool)
	if fipsEnabled {
		fipsValidatedModules = redhatopenshift.FipsValidatedModulesEnabled
	}

	return &redhatopenshift.ClusterProfile{
		ResourceGroupID:      utils.String(resourceGroupId),
		Domain:               utils.String(domain),
		PullSecret:           utils.String(pullSecret),
		FipsValidatedModules: fipsValidatedModules,
	}
}

func expandOpenshiftServicePrincipalProfile(
	input []interface{},
) *redhatopenshift.ServicePrincipalProfile {
	if len(input) == 0 {
		return nil
	}

	config := input[0].(map[string]interface{})

	clientId := config["client_id"].(string)
	clientSecret := config["client_secret"].(string)

	return &redhatopenshift.ServicePrincipalProfile{
		ClientID:     utils.String(clientId),
		ClientSecret: utils.String(clientSecret),
	}
}

func expandOpenshiftNetworkProfile(input []interface{}) *redhatopenshift.NetworkProfile {
	if len(input) == 0 {
		return &redhatopenshift.NetworkProfile{
			PodCidr:     utils.String("10.128.0.0/14"),
			ServiceCidr: utils.String("172.30.0.0/16"),
		}
	}

	config := input[0].(map[string]interface{})

	podCidr := config["pod_cidr"].(string)
	serviceCidr := config["service_cidr"].(string)

	return &redhatopenshift.NetworkProfile{
		PodCidr:     utils.String(podCidr),
		ServiceCidr: utils.String(serviceCidr),
	}
}

func expandOpenshiftMasterProfile(input []interface{}) *redhatopenshift.MasterProfile {
	if len(input) == 0 {
		return nil
	}

	config := input[0].(map[string]interface{})

	vmSize := config["vm_size"].(string)
	subnetId := config["subnet_id"].(string)

	encryptionAtHost := redhatopenshift.EncryptionAtHostDisabled
	enableEncryptionAtHost := config["encryption_at_host_enabled"].(bool)
	if enableEncryptionAtHost {
		encryptionAtHost = redhatopenshift.EncryptionAtHostEnabled
	}

	diskEncryptionSetId := config["disk_encryption_set_id"].(string)

	return &redhatopenshift.MasterProfile{
		VMSize:              utils.String(vmSize),
		SubnetID:            utils.String(subnetId),
		EncryptionAtHost:    encryptionAtHost,
		DiskEncryptionSetID: utils.String(diskEncryptionSetId),
	}
}

func expandOpenshiftWorkerProfiles(inputs []interface{}) *[]redhatopenshift.WorkerProfile {
	if len(inputs) == 0 {
		return nil
	}

	profiles := make([]redhatopenshift.WorkerProfile, 0)
	config := inputs[0].(map[string]interface{})

	name := "worker"
	vmSize := config["vm_size"].(string)
	diskSizeGb := int32(config["disk_size_gb"].(int))
	nodeCount := int32(config["node_count"].(int))
	subnetId := config["subnet_id"].(string)

	encryptionAtHost := redhatopenshift.EncryptionAtHostDisabled
	enableEncryptionAtHost := config["encryption_at_host_enabled"].(bool)
	if enableEncryptionAtHost {
		encryptionAtHost = redhatopenshift.EncryptionAtHostEnabled
	}

	diskEncryptionSetId := config["disk_encryption_set_id"].(string)

	profile := redhatopenshift.WorkerProfile{
		Name:                utils.String(name),
		VMSize:              utils.String(vmSize),
		DiskSizeGB:          utils.Int32(diskSizeGb),
		SubnetID:            utils.String(subnetId),
		Count:               utils.Int32(nodeCount),
		EncryptionAtHost:    encryptionAtHost,
		DiskEncryptionSetID: utils.String(diskEncryptionSetId),
	}

	profiles = append(profiles, profile)

	return &profiles
}

func expandOpenshiftApiServerProfile(input []interface{}) *redhatopenshift.APIServerProfile {
	if len(input) == 0 {
		return &redhatopenshift.APIServerProfile{
			Visibility: redhatopenshift.VisibilityPublic,
		}
	}

	config := input[0].(map[string]interface{})

	visibility := config["visibility"].(string)

	return &redhatopenshift.APIServerProfile{
		Visibility: redhatopenshift.Visibility(visibility),
	}
}

func expandOpenshiftIngressProfiles(inputs []interface{}) *[]redhatopenshift.IngressProfile {
	profiles := make([]redhatopenshift.IngressProfile, 0)

	name := "default"
	visibility := string(redhatopenshift.VisibilityPublic)

	if len(inputs) > 0 {
		input := inputs[0].(map[string]interface{})
		visibility = input["visibility"].(string)
	}

	profile := redhatopenshift.IngressProfile{
		Name:       utils.String(name),
		Visibility: redhatopenshift.Visibility(visibility),
	}

	profiles = append(profiles, profile)

	return &profiles
}