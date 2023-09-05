module github.com/cycloidio/terracost

go 1.16

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/cycloidio/sqlr v1.0.0
	github.com/dmarkham/enumer v1.5.3
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang/mock v1.6.0
	github.com/gruntwork-io/terragrunt v0.0.0-00010101000000-000000000000
	github.com/hashicorp/hcl/v2 v2.15.0
	github.com/hashicorp/terraform v1.0.11
	github.com/lopezator/migrator v0.3.0
	github.com/machinebox/progress v0.2.0
	github.com/matryer/is v1.4.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0
	github.com/shopspring/decimal v1.3.1
	github.com/spf13/afero v1.9.3
	github.com/stretchr/testify v1.7.2
	github.com/zclconf/go-cty v1.12.1
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/text v0.7.0
	golang.org/x/tools v0.4.0
	google.golang.org/api v0.100.0
)

replace github.com/hashicorp/terraform => github.com/cycloidio/terraform v1.1.9-cy

replace github.com/gruntwork-io/terragrunt => github.com/cycloidio/terragrunt v0.0.0-20230905115542-1fe1ff682fd9
