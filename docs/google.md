# Google

## Adding new resources

For the Google services pricing we have to be aware of this 2 APIs:
* [List of Services](https://cloud.google.com/billing/v1/how-tos/catalog-api#listing_public_services_from_the_catalog)
* [List of Service SKUs](https://cloud.google.com/billing/v1/how-tos/catalog-api#getting_the_list_of_skus_for_a_service)

If it's from a **new service** (not listed on the `/google/service.go` you first have to add it to the list and find the right mapping to then add to the 
 services map so it can be validated, to get the actual service ID you have to use this [API](https://cloud.google.com/billing/v1/how-tos/catalog-api#getting_the_list_of_skus_for_a_service).

If you have to add a **new resource** then you have to first find out to **which service** does it belong, **if we are not supporting it** yet (`google/service.go`) then you'll have to add it to the list and find the right ID (using the list services API) to put then on the `services` variable on the same file, this will allow the importer to be able support that service.

If if **we support the service** then the only missing thing is to add the new resource, for that I would follow the pattern we already have already for any other resource in terms of code:
* Add the new resource into the `terraform/` with a file name of the resource removing the provider prefix (ex: `google_compute_instance`->`compute_instance.go`)
* Add it to the `google/terraform/provider.go`  on the `ResourceComponents`
* Create the `decode{RESOURCE}Values` which reads from the raw values from Terraform to get the needed information to calculate the price (ex: `machinge_type`) by having a `mapstructure` directly mapping it
* Create the `Components()` func that is used to calculate the price of an specific resource by all the components/attributes we support by creating a `query.Component` that would return the specific product+price for that attribute.

The last step, the query, is the most important as it may require changes to the `google/ingester.go` to add specific attributes to the price so the query can be precise enough to return what is needed and not anything else. To know that the only was it to play with the SKUs API for the specific Service and check the results and find a way to map it to what you want by normalizing some elements by adding them to the attributes.

When all this is done you may want to check the `google/filter.go#MinimalFilter` to add anything else you want to filter so we do not ingest all the SKUs but only the ones we need.

## List of supported resources and attributes

* [`google_compute_instance`](#google_compute_instance)

### `google_compute_instance`

#### Cost factors

* Location
* Machine type

#### Additional notes

* For the machine type we use the https://cloud.google.com/compute/docs/machine-types API and add those as generated SKU and calculate the price by mapping it with the SKUs of CPU & RAM
