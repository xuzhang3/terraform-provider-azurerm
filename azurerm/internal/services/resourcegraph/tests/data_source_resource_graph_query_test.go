package tests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
)

func TestAccDataSourceAzureRMResourceGraphQuery_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "data.azurerm_resource_graph_query", "test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    acceptance.SupportedProviders,
		CheckDestroy: testCheckAzureRMResourceGraphQueryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceResourceGraphQuery_basic(data),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMResourceGraphQueryExists(data.ResourceName),
					resource.TestCheckResourceAttrSet(data.ResourceName, "resource_group_name"),
					resource.TestCheckResourceAttrSet(data.ResourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceResourceGraphQuery_basic(data acceptance.TestData) string {
	config := testAccAzureRMResourceGraphQuery_basic(data)
	return fmt.Sprintf(`
%s

data "azurerm_resource_graph_query" "test" {
  resource_group_name = azurerm_resource_graph_query.test.resource_group_name
  name                = azurerm_resource_graph_query.test.name
}
`, config)
}
