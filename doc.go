// Package terracost provides functionality to estimate the costs of infrastructure based on Terrafom
// plan files.
//
// This package depends on the pricing data located in a MySQL database to work correctly. The following
// snippet will run all required database migrations and ingest pricing data from AmazonEC2 in eu-west-3 region:
//
//      db, err := sql.Open("mysql", "...")
//      backend := mysql.NewBackend(db)
//
//      // Run all database migrations
//      err = mysql.Migrate(ctx, db, "pricing_migrations")
//
//      // Ingest pricing data into the database
//      ingester := aws.NewIngester("AmazonEC2", "eu-west-3")
//      err = terracost.IngestPricing(ctx, backend, ingester)
//
// With pricing data in the database, a Terraform plan can be read and estimated:
//
//      file, err := os.Open("path/to/tfplan.json")
//      plan, err := terracost.EstimateTerraformPlan(ctx, backend, file)
//
//      for _, res := range plan.ResourceDifferences() {
//          fmt.Printf("%s: %s -> %s\n", res.Address, res.PriorCost().String(), res.PlannedCost().String())
//      }
package terracost
