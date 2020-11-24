---
subcategory: "Resource Graph"
layout: "azurerm"
page_title: "Azure Resource Manager: Data Source: azurerm_resourcegraph_graph_query"
description: |-
  Gets information about an existing Resource Graph.
---

# Data Source: azurerm_resourcegraph_graph_query

Use this data source to access information about an existing Resource Graph.

## Example Usage

```hcl
data "azurerm_resourcegraph_graph_query" "example" {
  name                = "existing"
  resource_group_name = "existing"
}

output "id" {
  value = data.azurerm_resourcegraph_graph_query.example.id
}
```

## Arguments Reference

The following arguments are supported:

* `resource_group_name` - (Required) The name of the Resource Group where the Resource Graph exists. Changing this forces a new Resource Graph to be created.

* `resource_name` - (Required) The name of the Graph Query resource. Changing this forces a new Resource Graph to be created.

---

* `description` - (Optional) The description of a graph query.

* `tags` - (Optional) A mapping of tags which should be assigned to the Resource Graph.

## Attributes Reference

In addition to the Arguments listed above - the following Attributes are exported: 

* `id` - The ID of the Resource Graph.

* `name` - The Name of this Resource Graph.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `read` - (Defaults to 5 minutes) Used when retrieving the Resource Graph.
