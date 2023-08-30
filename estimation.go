package terracost

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/afero"

	"github.com/cycloidio/terracost/backend"
	"github.com/cycloidio/terracost/cost"
	"github.com/cycloidio/terracost/terraform"
	"github.com/cycloidio/terracost/usage"
	"github.com/cycloidio/terracost/util"
	"github.com/gruntwork-io/terragrunt/cli"
	"github.com/gruntwork-io/terragrunt/cli/tfsource"
	"github.com/gruntwork-io/terragrunt/config"
	"github.com/gruntwork-io/terragrunt/configstack"
	"github.com/gruntwork-io/terragrunt/options"
)

// EstimateTerraformPlan is a helper function that reads a Terraform plan using the provided io.Reader,
// generates the prior and planned cost.State, and then creates a cost.Plan from them that is returned.
// It uses the Backend to retrieve the pricing data.
func EstimateTerraformPlan(ctx context.Context, be backend.Backend, plan io.Reader, u usage.Usage, providerInitializers ...terraform.ProviderInitializer) (*cost.Plan, error) {
	if len(providerInitializers) == 0 {
		providerInitializers = getDefaultProviders()
	}

	tfplan := terraform.NewPlan(providerInitializers...)
	if err := tfplan.Read(plan); err != nil {
		return nil, err
	}
	tfplan.SetUsage(u)

	priorQueries, err := tfplan.ExtractPriorQueries()
	if err != nil {
		return nil, err
	}

	// If it's the first time we run the plan, then we might not have
	// prior queries so we ignore it and move forward
	prior, err := cost.NewState(ctx, be, priorQueries)
	if err != nil && err != terraform.ErrNoQueries {
		return nil, err
	}

	plannedQueries, err := tfplan.ExtractPlannedQueries()
	if err != nil {
		return nil, err
	}
	planned, err := cost.NewState(ctx, be, plannedQueries)
	if err != nil {
		return nil, err
	}

	modules := make([]string, 0, 0)
	for k := range tfplan.Configuration.RootModule.ModuleCalls {
		modules = append(modules, k)
	}
	sort.Strings(modules)

	return cost.NewPlan(strings.Join(modules, ", "), prior, planned), nil
}

// EstimateHCL is a helper function that recursively reads Terraform modules from a directory at the
// given stackPath and generates a planned cost.State that is returned wrapped in a cost.Plan.
// It uses the Backend to retrieve the pricing data. The modulePath is used to know if the module
// is not defined on the root of the stack,
// If Force Terragrunt(ftg) is set then we'll just run Terragrunt
func EstimateHCL(ctx context.Context, be backend.Backend, afs afero.Fs, stackPath, modulePath string, ftg bool, u usage.Usage, providerInitializers ...terraform.ProviderInitializer) ([]*cost.Plan, error) {
	if len(providerInitializers) == 0 {
		providerInitializers = getDefaultProviders()
	}
	var (
		relModulePath string
		err           error
	)
	if modulePath != "" {
		relModulePath, err = filepath.Rel(stackPath, modulePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get relative path of %q from %q: %w", modulePath, stackPath, err)
		}
	} else {
		modulePath = stackPath
	}

	if afs == nil {
		afs = afero.NewOsFs()
	}

	if !ftg {
		var hasTG bool
		// We first check if the main main modulePath has a Terragrunt file to know what we have to run
		err = afero.Walk(afs, modulePath, func(p string, info fs.FileInfo, err error) error {
			if info.IsDir() || hasTG {
				return nil
			}
			relpath, _ := filepath.Rel(modulePath, p)
			// As we only want to check on the root directory anything with / can be skipped
			if strings.Contains(relpath, string(os.PathSeparator)) {
				return nil
			}
			if relpath == config.DefaultTerragruntConfigPath || relpath == config.DefaultTerragruntJsonConfigPath {
				hasTG = true
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk path %q: %w", modulePath, err)
		}

		// If no Terragrunt file is found then we execute the normal code
		if !hasTG {
			plannedQueries, modAddr, err := terraform.ExtractQueriesFromHCL(afs, providerInitializers, modulePath, u, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to extract queries from hcl: %w", err)
			}
			planned, err := cost.NewState(ctx, be, plannedQueries)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize a state: %w", err)
			}

			return []*cost.Plan{cost.NewPlan(modAddr, nil, planned)}, nil
		}
	}

	// We create a tmp dir to move the files from fs to it so we can
	// run Terragrunt on it. Terragrunt only runs on OS
	tmpdir, err := os.MkdirTemp("", "terracost-terragrunt")
	if err != nil {
		return nil, fmt.Errorf("failed to create a temp dir: %w", err)
	}
	defer os.RemoveAll(tmpdir)

	// We move the files from afs stackPath to the just created tmpdir
	err = util.FromAferoToOS(afs, stackPath, tmpdir)
	if err != nil {
		return nil, fmt.Errorf("failed to move content from Afero(%q) to OS: %w", stackPath, err)
	}

	tgo, err := options.NewTerragruntOptions(filepath.Join(tmpdir, relModulePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create terragrunt options for %s: %w", tmpdir, err)
	}
	tgo.RunTerragrunt = cli.RunTerragrunt
	// DryRun is an specific option we added to the fork of Terragrunt we have.
	// This fork allows us to run everything except Terraform, so we have all
	// the Terragrunt code run that generates the modules and code so then we
	// can read that generated code and run TerraCost
	tgo.DryRun = true
	// We set Writer and ErrWriter to io.Discard so we do not get
	// any logs on the screen when running test of the tool itself
	tgo.Writer = io.Discard
	tgo.ErrWriter = io.Discard

	// We need to initialize the tmpdir as a git repository because if the Terragrunt
	// config has any of the functions like 'get_repo_root' it would fail if it's not
	// a git repository
	_, err = git.PlainInit(tmpdir, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git repo %q: %w", tmpdir, err)
	}

	// We initialize all the stacks from the modulePath URL
	stack, err := configstack.FindStackInSubfolders(tgo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to FindStackInSubfolders: %w", err)
	}

	// Runs Terragrunt which basically generates some submodules
	err = stack.Run(tgo)
	if err != nil {
		return nil, fmt.Errorf("failed to run stack %q: %w", stack.Path, err)
	}

	costs := make([]*cost.Plan, 0)
	for _, m := range stack.Modules {
		sourceUrl, err := config.GetTerraformSourceUrl(m.TerragruntOptions, &m.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to get terraform source url: %w", err)
		}

		// We need to get the terraformSource as it has the '.WorkingDir' which has the right path of the module just downloaded on the 'stack.Run'
		// this path is not predictable so we need to get it from this 'terraformSource'
		terraformSource, err := tfsource.NewTerraformSource(sourceUrl, m.TerragruntOptions.DownloadDir, m.TerragruntOptions.WorkingDir, m.TerragruntOptions.Logger)
		if err != nil {
			return nil, fmt.Errorf("failed to get terraform source: %w", err)
		}
		nfs := afero.NewMemMapFs()

		// We move the downloaded and generated code+module from the 'terraformSource.WorkingDir' (which is on the OS) to the 'nfs' which
		// is a Memory implementation
		err = util.FromOSToAfero(nfs, terraformSource.WorkingDir, "")
		if err != nil {
			return nil, fmt.Errorf("failed to move content from OS(%q) to Afero: %w", terraformSource.WorkingDir, err)
		}

		// We ReadTerragruntConfig so we can have the 'tgc.Inputs' which has the values+variables
		// that we need to set to the module. Normally those inputs are passed via ENV variables
		// when Terragrunt is running
		tgc, _ := config.ReadTerragruntConfig(m.TerragruntOptions)
		plannedQueries, modAddr, err := terraform.ExtractQueriesFromHCL(nfs, providerInitializers, "", u, tgc.Inputs)
		if err != nil {
			return nil, err
		}
		planned, err := cost.NewState(ctx, be, plannedQueries)
		if err != nil {
			return nil, err
		}

		// If no module is defined we can always use the name of the WorkingDir in which
		// TG found the modules
		if modAddr == "" {
			modAddr = filepath.Base(m.TerragruntOptions.WorkingDir)
		}

		costs = append(costs, cost.NewPlan(modAddr, nil, planned))
	}
	return costs, nil
}
