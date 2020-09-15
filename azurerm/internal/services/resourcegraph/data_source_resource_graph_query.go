package resourcegraph

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceResourceGraphGraphQuery() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArmResourceGraphGraphQueryRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),
			"query": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tags.SchemaDataSource(),
		},
	}
}

func dataSourceArmResourceGraphGraphQueryRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).ResourceGraph.GraphQueryClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroup := d.Get("resource_group_name").(string)
	resourceName := d.Get("name").(string)

	resp, err := client.Get(ctx, resourceGroup, resourceName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] ResourceGraph %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("retrieving Resource Graph Query %q (Resource Group %q): %+v", resourceName, resourceGroup, err)
	}
	if id := resp.ID; id != nil {
		d.SetId(*resp.ID)
	}
	d.Set("resource_group_name", resourceGroup)
	d.Set("name", resp.Name)

	if props := resp.GraphQueryProperties; props != nil {
		d.Set("query", props.Query)
		d.Set("description", props.Description)
	}
	return nil
}
