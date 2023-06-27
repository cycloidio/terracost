module github.com/cycloidio/terracost

go 1.16

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/cycloidio/sqlr v1.0.0
	github.com/dmarkham/enumer v1.5.3
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/golang/mock v1.6.0
	github.com/hashicorp/go-hclog v1.0.0 // indirect
	github.com/hashicorp/hcl/v2 v2.11.1
	github.com/hashicorp/terraform v1.0.11
	github.com/lopezator/migrator v0.3.0
	github.com/machinebox/progress v0.2.0
	github.com/matryer/is v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mitchellh/mapstructure v1.5.0
	github.com/shopspring/decimal v1.3.1
	github.com/spf13/afero v1.6.0
	github.com/stretchr/testify v1.7.0
	github.com/zclconf/go-cty v1.10.0
	golang.org/x/crypto v0.0.0-20211117183948-ae814b36b871 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.0.0-20211124211545-fe61309f8881 // indirect
	golang.org/x/text v0.3.7
	golang.org/x/tools v0.1.7
	google.golang.org/api v0.60.0
	google.golang.org/genproto v0.0.0-20211118181313-81c1377c94b1 // indirect
	google.golang.org/grpc v1.42.0 // indirect
)

replace github.com/hashicorp/terraform => github.com/cycloidio/terraform v1.1.9-cy
