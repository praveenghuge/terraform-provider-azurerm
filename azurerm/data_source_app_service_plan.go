package azurerm

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceAppServicePlan() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAppServicePlanRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"location": azure.SchemaLocationForDataSource(),

			"kind": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"sku": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tier": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"capacity": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"properties": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app_service_environment_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"reserved": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"per_site_scaling": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"maximum_number_of_workers": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"maximum_elastic_worker_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"is_xenon": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"tags": tagsForDataSourceSchema(),
		},
	}
}

func dataSourceAppServicePlanRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*ArmClient).appServicePlansClient

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	ctx := meta.(*ArmClient).StopContext
	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return fmt.Errorf("Error making Read request on App Service Plan %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if utils.ResponseWasNotFound(resp.Response) {
		return fmt.Errorf("Error: App Service Plan %q (Resource Group %q) was not found", name, resourceGroup)
	}

	d.SetId(*resp.ID)

	d.Set("name", name)
	d.Set("resource_group_name", resourceGroup)
	d.Set("kind", resp.Kind)
	d.Set("is_xenon", resp.IsXenon)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.AppServicePlanProperties; props != nil {
		if err := d.Set("properties", flattenAppServiceProperties(props)); err != nil {
			return fmt.Errorf("Error setting `properties`: %+v", err)
		}

		if props.MaximumNumberOfWorkers != nil {
			d.Set("maximum_number_of_workers", int(*props.MaximumNumberOfWorkers))
		}

		if props.MaximumElasticWorkerCount != nil {
			d.Set("maximum_elastic_worker_count", int(*props.MaximumElasticWorkerCount))
		}
	}

	if err := d.Set("sku", flattenAppServicePlanSku(resp.Sku)); err != nil {
		return fmt.Errorf("Error setting `sku`: %+v", err)
	}

	flattenAndSetTags(d, resp.Tags)

	return nil
}
