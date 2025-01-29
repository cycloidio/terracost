# TerraCost examples

Examples help you to understand how to test TerraCost.

## Requirements

### Mysql

Cloud Provider pricing data need to be ingested in a Mysql server. For testing purpose, local docker can be used

```
docker run  -p 3306:3306 -d --privileged  -e MYSQL_ROOT_PASSWORD=terracost  mysql:8.4.3
```

Once Mysql started, create the terracost database
```
mysql -h 127.0.0.1 -uroot -pterracost -e "CREATE DATABASE terracost_test"
```

## Examples

### Pricing ingestion

To start prices ingestion you need to first decide which cloud provider to ingest. In the example `aws` or `azurerm` are availables with `-ingest -provider`.

```
go run terracost.go -ingest -provider aws -ingest-region eu-west-1
go run terracost.go -ingest -provider azurerm -ingest-region francecentral
```

> Note
> Before to ingest datas, the database migrations are needed. In order to correctly run migrations, we need `?multiStatements=true` param on DB connector
> ```
> sql.Open("mysql", "root:terracost@tcp(127.0.0.1:3306)/terracost_test?multiStatements=true")
> ```

### Pricing Estimation (from Plan)

To estimate a terraform plan, you need to first generate it as json format

```
terraform init
terraform plan -out update.tfplan
terraform show -json update.tfplan > terraform-plan.json
```

Then run the estimation by specifying your json plan file path.

```
go run terracost.go -estimate-plan ./terraform-plan.json
```

### Pricing Estimation (from HCL)

To estimate terraform code, define the terraform provider to use and the path of your terraform hcl code.
```
go run terracost.go -provider aws -estimate-hcl ../testdata/aws/stack-aws
```

### Tips to check the billing queries

```
mysql -h 127.0.0.1 -uroot -pterracost terracost_test

mysql>select * from pricing_products where service="$SERVICE" and family=""$FAMILY"" and JSON_EXTRACT(attributes, "$.skuName") = "$SKUNAME" and JSON_EXTRACT(attributes, "$.meterName") = "$METERNAME";
+--+----------+-----+----------+---------+---------+-----------+
id | provider | sku | location | service | family  | attributes
+--+----------+-----+----------+---------+---------+-----------+
01 | azurem   | sku | location | service | family  | {...}
+--+----------+-----+----------+---------+---------+-----------+
```

Remember that the query above correspond to the query to the billing API like this
```
curl -s "https://prices.azure.com/api/retail/prices?\$filter=serviceName eq '$SERVICE'" and  serviceFamily eq '$FAMILY' and skuName eq 'SKUNAME' and meterName eq 'METERNAME'| jq .
```
