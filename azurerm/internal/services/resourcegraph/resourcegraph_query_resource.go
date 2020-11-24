package resourcegraph

import (
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/resourcegraph/mgmt/2018-09-01/resourcegraph"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/resourcegraph/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmResourceGraphGraphQuery() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmResourceGraphGraphQueryCreate,
		Read:   resourceArmResourceGraphGraphQueryRead,
		Update: resourceArmResourceGraphGraphQueryUpdate,
		Delete: resourceArmResourceGraphGraphQueryDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			_, err := parse.ResourceGraphGraphQueryID(id)
			return err
		}),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"query": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tags": tags.Schema(),
		},
	}
}
func resourceArmResourceGraphGraphQueryCreate(d *schema.ResourceData, meta interface{}) error {
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	client := meta.(*clients.Client).ResourceGraph.GraphQueryClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	existing, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for present of existing Resourcegraph GraphQuery %q (Resource Group %q): %+v", name, resourceGroup, err)
		}
	}
	if existing.ID != nil && *existing.ID != "" {
		return tf.ImportAsExistsError("azurerm_resourcegraph_graph_query", *existing.ID)
	}

	props := resourcegraph.GraphQueryResource{
		Location: utils.String("global"),
		GraphQueryProperties: &resourcegraph.GraphQueryProperties{
			Description: utils.String(d.Get("description").(string)),
			Query:       utils.String(d.Get("query").(string)),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}
	if _, err := client.CreateOrUpdate(ctx, resourceGroup, name, props); err != nil {
		return fmt.Errorf("creating Resourcegraph GraphQuery %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return fmt.Errorf("retrieving Resourcegraph GraphQuery %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if resp.ID == nil || *resp.ID == "" {
		return fmt.Errorf("empty or nil ID returned for Resourcegraph GraphQuery %q (Resource Group %q) ID", name, resourceGroup)
	}

	id, err := parse.ResourceGraphGraphQueryID(*resp.ID)
	if err != nil {
		return err
	}
	d.SetId(id.ID(subscriptionId))

	return resourceArmResourceGraphGraphQueryRead(d, meta)
}

func resourceArmResourceGraphGraphQueryRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ResourceGraph.GraphQueryClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceGraphGraphQueryID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] resourcegraph %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Resourcegraph GraphQuery %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	if props := resp.GraphQueryProperties; props != nil {
		d.Set("description", props.Description)
		d.Set("query", props.Query)
	}
	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmResourceGraphGraphQueryUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ResourceGraph.GraphQueryClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceGraphGraphQueryID(d.Id())
	if err != nil {
		return err
	}

	body := resourcegraph.GraphQueryUpdateParameters{
		GraphQueryPropertiesUpdateParameters: &resourcegraph.GraphQueryPropertiesUpdateParameters{
			Query: utils.String(d.Get("query").(string)),
		},
	}

	if d.HasChange("description") {
		body.GraphQueryPropertiesUpdateParameters.Description = utils.String(d.Get("description").(string))
	}

	if d.HasChange("tags") {
		body.Tags = tags.Expand(d.Get("tags").(map[string]interface{}))
	}

	if _, err := client.Update(ctx, id.ResourceGroup, id.Name, body); err != nil {
		return fmt.Errorf("updating Resourcegraph GraphQuery %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	return resourceArmResourceGraphGraphQueryRead(d, meta)
}

func resourceArmResourceGraphGraphQueryDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ResourceGraph.GraphQueryClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ResourceGraphGraphQueryID(d.Id())
	if err != nil {
		return err
	}

	if _, err := client.Delete(ctx, id.ResourceGroup, id.Name); err != nil {
		return fmt.Errorf("deleting Resourcegraph GraphQuery %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	return nil
}
