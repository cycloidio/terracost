# AzureRM

## Support new resources

For the AzureRM services pricing we have to be aware of this 2 APIs:
* [List of Services](https://azure.microsoft.com/en-us/services/): The left side are the Families and the content of the page are the services
* [List of Service SKUs](https://docs.microsoft.com/en-us/rest/api/cost-management/retail-prices/azure-retail-prices)

If you need to add a **new resource**, the first step is to determine **which service** it belongs to. You can retrieve a list of available services from the Azure Prices API using the following script:

```
#!/bin/bash

# Initial API URL
url="https://prices.azure.com/api/retail/prices"

:> /tmp/azuresvc
# Loop to paginate through all the results
while [[ "$url" != "null" ]]; do
    # Fetch the JSON response
    response=$(curl -s "$url")

    # Extract and print all the service names (unique)
    echo "$response" | jq -r '.Items[] | "\(.serviceName) -- \(.serviceFamily) -- \(.productName)"' >> /tmp/azuresvc

    if [ $? -eq 0 ]; then
      # Get the next page link
      url=$(echo "$response" | jq -r '.NextPageLink')
      echo $url
    else
      echo "retry $url"
    fi
done
echo "serviceName -- serviceFamily -- productName"
echo ""
cat /tmp/azuresvc | sort -u
```

**If we are not yet supporting the resource** in the codebase (`azurerm/service.go`), you'll need to [Add new service](#adding-new-service) to enable the importer to handle that resource. This will ensure the service is supported, and any associated resources can be properly managed.

### (Optional) Adding new service

1. Add it to [the list](https://github.com/cycloidio/terracost/blob/master/azurerm/service.go#L13-L23) and `services` map structure
2. Run `make generate` to regenerate the `service_string.go` list.


### Adding new resource

If we **already support the service**, the only remaining step is to add the new resource.

1. Add the new resource into the `terraform/` with a file name of the resource removing the provider prefix (ex: `azurerm_public_ip`->`public_ip.go`)
2. As a starting point, copy the content from `public_ip.go` into your new resource file
3. Replace function/variable names such as
```
sed 's/PublicIP/NewResource/g' -i new_resource.go
sed 's/publicIP/newResource/g' -i new_resource.go
```
4. Add the resource in the `azurerm/terraform/provider.go` on the `ResourceComponents`
5. Check the resource parameters in terraform and note the variables that impact the prices. Don't forget to divide them by optional and required ones. Compare those parameters with the ones from the API of azure. You can use curl like this to get more information :
```
curl -s "https://prices.azure.com/api/retail/prices?\$filter=productName eq '$API_PRODUCT_NAME'" | jq '.Items[] | {skuName, meterName, unitOfMeasure}' | sort -u
curl -s "https://prices.azure.com/api/retail/prices?\$filter=productName eq '$API_PRODUCT_NAME' and meterName eq '$METER_NAME'"| jq .

```
5. Go back to your resource file and update the `{newResource}` that contains the values that are required to calculate the price of an object and `{newResource}Values` that contains the values that are defined in terraform.
The `decode{RESOURCE}Values` function utilizes the `{newResource}Values` struct to map the values defined in the Terraform.
These mapped values are then used to calculate the resource's pricing based on the relevant attributes defined in Terraform, such as `size` mapped with `mapstructure:"size"`

6. An important tip here is in the case your resource needs either to access parameters from another resource to get some parameters required for the billing API queries. You can check the storage_share.go for an example, where the resource needs to access some elements defined in the storage_account.

7. If you need to defined pre-configured values you should do it at usage/usage.go. Check storage_share.go for an example.

8. Define your `query.Component` function with the right `AttributeFilters` of the object, usually the `skuName`, `meterName` or `productName` as they are defined in the billing API, these will allow to query it and get the specific price value for an object.

The `Components()` func that is used to calculate the price of an specific resource by all the components/attributes we support by creating a `query.Component` that would return the specific product+price for that attribute.

Ensure that you define the correct attributes and values in your resource to accurately reflect the Terraform configuration and match the corresponding Azure pricing.

Azure pricing attributes can be referenced from the [official documentation](https://learn.microsoft.com/en-us/rest/api/cost-management/retail-prices/azure-retail-prices#api-property-details)

Example to get the attributes available for your specific service:
```bash
REQ="https://prices.azure.com/api/retail/prices?\$filter=serviceName eq 'VPN Gateway'"
# REQ="https://prices.azure.com/api/retail/prices?\$filter=serviceName eq 'Virtual Machines' and (armRegionName eq 'northeurope' or armRegionName eq 'Zone 1')"
# repalce escape and quote
REQ=$(echo $REQ | sed "s/ /%20/g;s/'/%27/;s/(/%28/g;s/)/%29/g")

curl $REQ | jq .
```

9. You can test the implementation by following the tutorial at [example-azure](../examples/README.md)

10. Verify the cost implemented match [Azure Calculator](https://azure.microsoft.com/en-us/pricing/calculator/)

11. Don't forget to add the resource in the list of supported resources bellow!

## List of supported resources and attributes

<!--
for i in $(grep azurerm_ azurerm/terraform/provider.go | sed -E 's/[^"]+"azurerm_//;s/".*//' | sort);do
  echo '* [`azurerm_'$i'`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/'$i')';
done
-->
* [`azurerm_bastion_host`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/bastion_host)
* [`azurerm_dns_zone`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/dns_zone)
* [`azurerm_linux_virtual_machine`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/linux_virtual_machine)
* [`azurerm_managed_disk`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/managed_disk)
* [`azurerm_nat_gateway`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/nat_gateway)
* [`azurerm_private_dns_zone`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/private_dns_zone)
* [`azurerm_private_endpoint`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/private_endpoint)
* [`azurerm_public_ip`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/public_ip)
* [`azurerm_storage_account`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/storage_account)
* [`azurerm_storage_share`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/storage_share)
* [`azurerm_virtual_machine`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_machine)
* [`azurerm_virtual_network_gateway`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network_gateway)
* [`azurerm_virtual_network_gateway_connection`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/virtual_network_gateway_connection)
* [`azurerm_windows_virtual_machine`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/windows_virtual_machine)
* [`azurerm_postgresql_flexible_server`](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/postgresql_flexible_server)
