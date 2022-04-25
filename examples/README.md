# TerraCost examples

Examples help you to understand how to test Terracost.

## Requirements

### Mysql

Cloud Provider pricing datas need to be ingested in a Mysql server. For testing purpose, local docker can be used

```
docker run  -p 3306:3306 -d --privileged  -e MYSQL_ROOT_PASSWORD=password  mysql:8.0.21 --default-authentication-plugin=mysql_native_password

# Once Mysql started, create the terracost database
mysql -h 127.0.0.1 -uroot -ppassword -e "CREATE DATABASE terracost"
```

## Examples

### Pricing ingestion

To start prices ingestion you need to first decide which cloud provider to ingest. In the example `AWS` is available with `-ingest-aws`.

 and services

 in `terracost-ingester.go`

```
go run terracost.go -ingest-aws
```

Here we ingest all supported services by Terracost, ingestion could takes severals minutes depending the amonts of datas.
If you don't need all services, the list could be defined to reduce time.

> Note
> Before to ingest datas, the database migrations are needed. In order to correctly run migrations, we need `?multiStatements=true` param on DB connector
> ```
> sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/terracost?multiStatements=true")
> ```

### Pricing Estimation (from Plan)

To estimate a terraform plan, you need to first generate it as json format

```
terraform init
terraform plan -out update.tfplan
terraform show -json update.tfplan > tfplan.json
```

Then run the estimation by specifying your json plan file path.

```
go run terracost.go -estimate-plan ./terraform-plan.json
```

### Pricing Estimation (from HCL)

To estimate terraform code, define the terraform provider to use and the path of your terraform hcl code.
```
go run terracost.go -estimate-hcl-provider aws -estimate-hcl ../testdata/aws/stack-aws
```
