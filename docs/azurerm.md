# AzureRM

## Support or new resources

For the AzureRM services pricing we have to be aware of this 2 APIs:
* [List of Services](https://azure.microsoft.com/en-us/services/): The left side are the Families and the content of the page are the services
* [List of Service SKUs](https://docs.microsoft.com/en-us/rest/api/cost-management/retail-prices/azure-retail-prices)

If you need to add a **new resource**, the first step is to determine **which service** it belongs to. You can retrieve a list of available services from the Azure Prices API using the following command:

`curl https://prices.azure.com/api/retail/prices | jq -r .Items[].serviceName | sort -u`

**If we are not yet supporting the resource** in the codebase (`azurerm/service.go`), you'll need to [Add new service](#adding-new-service) to enable the importer to handle that resource. This will ensure the service is supported, and any associated resources can be properly managed.

### Adding new service

1. Add it to [the list](https://github.com/cycloidio/terracost/blob/master/azurerm/service.go#L13-L23) and `services` map structure
2. Run `make generate` to regenerate the `service_string.go` list.


### Adding new resource

If we **already support the service**, the only remaining step is to add the new resource.

1. Add the new resource into the `terraform/` with a file name of the resource removing the provider prefix (ex: `azurerm_linux_virtual_machine`->`linux_virtual_machine.go`)
2. As a starting point, copy the content from `linux_virtual_machine.go` into your new resource file
3. Replace function/variable names such as
```
sed 's/LinuxVirtualMachine/NewResource/g'
sed 's/linuxVirtualMachine/newResource/g'
```
4. Add the resource in the `azurerm/terraform/provider.go` on the `ResourceComponents`
5. Go back to your resource file and update the `{newResource}Values` struct
The `decode{RESOURCE}Values` function utilizes the `{newResource}Values` struct to map the values defined in the Terraform.
These mapped values are then used to calculate the resource's pricing based on the relevant attributes defined in Terraform, such as `size` mapped with `mapstructure:"size"`

6. Define your `query.Component` function with the right `AttributeFilters`

The `Components()` func that is used to calculate the price of an specific resource by all the components/attributes we support by creating a `query.Component` that would return the specific product+price for that attribute.

Ensure that you define the correct attributes and values in your resource to accurately reflect the Terraform configuration and match the corresponding Azure pricing.

Azure pricing attributes can be referenced from the [official documentation](https://learn.microsoft.com/en-us/rest/api/cost-management/retail-prices/azure-retail-prices#api-property-details)

Example to get the attributes available for your specific service:
```bash
REQ="https://prices.azure.com/api/retail/prices?\$filter=serviceName eq 'Virtual Machines'"
# repalce escape and quote
REQ=$(echo $REQ | sed "s/ /%20/g;s/'/%27/")

curl $REQ | jq .
```

## List of supported resources and attributes

* [`azurerm_virtual_machine`](#azurerm_virtual_machine)
* [`azurerm_linux_virtual_machine`](#azurerm_linux_virtual_machine)

### `azurerm_virtual_machine`

#### Cost factors

* Location
* VMSize

#### Additional notes

* We assume all the types are linux

### `azurerm_linux_virtual_machine`

#### Cost factors

* Location
* Size

#### Additional notes
