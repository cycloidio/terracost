package terraform_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cycloidio/terracost/mock"
	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/terraform"
)

func TestExtractQueriesFromHCL(t *testing.T) {
	t.Run("AWS", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			fs := afero.NewOsFs()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			provider := mock.NewTerraformProvider(ctrl)
			providerInitializers := []terraform.ProviderInitializer{{
				MatchNames: []string{"aws", "aws-test"},
				Provider: func(_ map[string]string) (terraform.Provider, error) {
					return provider, nil
				},
			}}

			provider.EXPECT().ResourceComponents(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(rss map[string]terraform.Resource, res terraform.Resource) []query.Component {
				switch res.Address {
				case "aws_instance.example":
					assert.Equal(t, "aws_instance", res.Type)
					assert.Equal(t, "example", res.Name)
					assert.Equal(t, map[string]interface{}{
						"ami":           "some-ami",
						"instance_type": "t2.micro",
						"provider":      ".aws",
					}, res.Values)

				case "module.ec2.aws_instance.front":
					assert.Equal(t, "aws_instance", res.Type)
					assert.Equal(t, "front", res.Name)
					assert.Equal(t, map[string]interface{}{
						"ami":           "module.ec2.data.aws_ami.debian",
						"count":         float64(1),
						"instance_type": "t3.small",
						"root_block_device": []interface{}{
							map[string]interface{}{
								"delete_on_termination": true,
								"volume_size":           float64(123),
								"volume_type":           "gp2",
							},
						},
					}, res.Values)

				case "module.ec2.aws_elb.front":
					assert.Equal(t, "aws_elb", res.Type)
					assert.Equal(t, "front", res.Name)
					assert.Equal(t, map[string]interface{}{
						"instances": []string{"module.ec2.aws_instance.front[0]"},
						"listener": []interface{}{
							map[string]interface{}{
								"instance_port":     float64(80),
								"instance_protocol": "tcp",
								"lb_port":           float64(80),
								"lb_protocol":       "tcp",
							},
						},
					}, res.Values)

				case "module.ec2.module.ebs.aws_ebs_volume.volume":
					assert.Equal(t, "aws_ebs_volume", res.Type)
					assert.Equal(t, "volume", res.Name)
					assert.Equal(t, map[string]interface{}{
						"size": float64(20),
						"type": "gp2",
					}, res.Values)

				case "module.rds.aws_db_instance.db":
					assert.Equal(t, "aws_db_instance", res.Type)
					assert.Equal(t, "db", res.Name)
					assert.Equal(t, map[string]interface{}{
						"allocated_storage": float64(10),
						"engine":            "mysql",
						"instance_class":    "db.t3.small",
						"multi_az":          true,
						"storage_type":      "gp2",
					}, res.Values)

				default:
					t.Errorf("unexpected resource: %s", res.Address)
				}

				assert.Equal(t, "managed", res.Mode)
				assert.Equal(t, "aws", res.ProviderName)
				return nil
			})

			queries, err := terraform.ExtractQueriesFromHCL(fs, providerInitializers, "../testdata/aws/stack-aws")
			require.NoError(t, err)
			require.Len(t, queries, 5)
			for _, q := range queries {
				require.Equal(t, "aws", q.Provider)
				require.NotEmpty(t, q.Type)
			}
		})

		t.Run("BadProvider", func(t *testing.T) {
			fs := afero.NewOsFs()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			providerInitializers := []terraform.ProviderInitializer{{
				MatchNames: []string{"aws", "aws-test"},
				Provider: func(_ map[string]string) (terraform.Provider, error) {
					return nil, errors.New("bad provider")
				},
			}}

			queries, err := terraform.ExtractQueriesFromHCL(fs, providerInitializers, "../testdata/aws/stack-aws")
			require.Error(t, err)
			require.Len(t, queries, 0)
		})

		t.Run("FailedResourceComponents", func(t *testing.T) {
			fs := afero.NewOsFs()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			provider := mock.NewTerraformProvider(ctrl)
			providerInitializers := []terraform.ProviderInitializer{{
				MatchNames: []string{"aws", "aws-test"},
				Provider: func(_ map[string]string) (terraform.Provider, error) {
					return provider, nil
				},
			}}

			provider.EXPECT().ResourceComponents(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(rss map[string]terraform.Resource, res terraform.Resource) []query.Component {
				return nil
			})

			queries, err := terraform.ExtractQueriesFromHCL(fs, providerInitializers, "../testdata/aws/stack-aws")
			require.NoError(t, err)
			require.Len(t, queries, 5)

			for _, res := range queries {
				assert.Len(t, res.Components, 0)
			}
		})
	})
}
