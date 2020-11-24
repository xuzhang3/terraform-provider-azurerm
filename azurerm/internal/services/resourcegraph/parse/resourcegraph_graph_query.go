package parse

import (
	"fmt"
	"log"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

type ResourceGraphGraphQueryId struct {
	ResourceGroup string
	Name          string
}

func NewResourceGraphGraphQueryID(resourcegroup string, name string) ResourceGraphGraphQueryId {
	return ResourceGraphGraphQueryId{
		ResourceGroup: resourcegroup,
		Name:          name,
	}
}

func (id ResourceGraphGraphQueryId) ID(subscriptionId string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ResourceGraph/queries/%s", subscriptionId, id.ResourceGroup, id.Name)
}

func ResourceGraphGraphQueryID(input string) (*ResourceGraphGraphQueryId, error) {
	log.Printf("TESTETSETESTTEST Resource ID %+v ,", input)
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, fmt.Errorf("parsing resourceGraphGraphQuery ID %q: %+v", input, err)
	}

	resourceGraphGraphQuery := ResourceGraphGraphQueryId{
		ResourceGroup: id.ResourceGroup,
	}
	if resourceGraphGraphQuery.Name, err = id.PopSegment("queries"); err != nil {
		return nil, err
	}
	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceGraphGraphQuery, nil
}
