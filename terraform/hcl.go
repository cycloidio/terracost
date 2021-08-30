package terraform

import (
	"fmt"
	"path"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/terraform/configs"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"

	"github.com/cycloidio/terracost/query"
)

// ExtractQueriesFromHCL returns the resources found in the module identified by the modPath.
func ExtractQueriesFromHCL(fs afero.Fs, providerInitializers []ProviderInitializer, modPath string) ([]query.Resource, error) {
	parser := configs.NewParser(fs)
	mod, diags := parser.LoadConfigDir(modPath)
	if diags.HasErrors() {
		return nil, fmt.Errorf(diags.Error())
	}

	evalCtx := getEvalCtx(mod, nil)

	providers, err := getHCLProviders(mod, evalCtx, providerInitializers)
	if err != nil {
		return nil, err
	}

	queries, err := extractHCLModule(providers, parser, modPath, "", mod, evalCtx)
	if err != nil {
		return nil, err
	}
	return queries, nil
}

// extractHCLModule returns the resources found in the provided module.
func extractHCLModule(providers map[string]Provider, parser *configs.Parser, modPath string, modName string, mod *configs.Module, evalCtx *hcl.EvalContext) ([]query.Resource, error) {
	queries := make([]query.Resource, 0, len(mod.ManagedResources))

	// knownProvider will remain false, until one resource from a
	// known provider is encountered. This allows to know if through
	// all the given resources some could be estimated or not
	knownProvider := false

	for rk, rv := range mod.ManagedResources {
		if modName != "" {
			rk = fmt.Sprintf("%s.%s", modName, rk)
		}

		providerKey := rv.Provider.Type
		if rv.ProviderConfigRef != nil {
			providerKey = rv.ProviderConfigRef.String()
		}
		provider, ok := providers[providerKey]
		if ok {
			knownProvider = true
		}

		// Parse the HCL body of the resource block and evaluate it. The JSON (in the form of map[string]interface{} type)
		// is then placed into the cfg.
		body, ok := rv.Config.(*hclsyntax.Body)
		if !ok {
			return nil, fmt.Errorf("invalid resource configuration body")
		}
		cfg := getBodyJSON(body, evalCtx)

		// Assume this is a single instance of this resource unless the cfg contains the "count" parameter.
		count := 1
		if c, ok := cfg["count"]; ok {
			if cf, ok := c.(float64); ok {
				count = int(cf)
			}
		}

		for i := 0; i < count; i++ {
			addr := rk
			if count > 1 {
				addr = fmt.Sprintf("%s[%d]", rk, i)
			}

			// Only retrieve components if the provider is valid. If it's not, the comps will be nil, which signifies
			// that the resource was "skipped" from estimation.
			var comps []query.Component
			if provider != nil {
				comps = provider.ResourceComponents(Resource{
					Address:      addr,
					Index:        i,
					Mode:         "managed",
					Type:         rv.Type,
					Name:         rv.Name,
					ProviderName: rv.Provider.Type,
					Values:       cfg,
				})
			}
			queries = append(queries, query.Resource{
				Address:    addr,
				Type:       rv.Type,
				Provider:   rv.Provider.Type,
				Components: comps,
			})
		}
	}

	// Recursively extract resources from all child module calls.
	for mk, mv := range mod.ModuleCalls {
		p := joinPath(modPath, mv.SourceAddr)

		// Try to load a module from a config directory. Only local modules are supported, other types of modules
		// will be skipped.
		child, diags := parser.LoadConfigDir(p)
		if diags.HasErrors() {
			// Skip unsupported modules
			continue
		}

		body, ok := mv.Config.(*hclsyntax.Body)
		if !ok {
			return nil, fmt.Errorf("invalid module call body")
		}

		// Extract variables from the module call block to pass down to the module. It's a map of
		// variable names to their evaluated values.
		vars := make(map[string]cty.Value)
		for _, attr := range body.Attributes {
			val, diags := attr.Expr.Value(evalCtx)
			if diags != nil && diags.HasErrors() {
				continue
			}
			vars[attr.Name] = val
		}

		nextEvalCtx := getEvalCtx(child, vars)

		// If the module call contains a `providers` block, it should replace the implicit provider
		// inheritance. Instead, a new map of parent to child providers is created.
		// https://www.terraform.io/docs/language/modules/develop/providers.html#passing-providers-explicitly
		var childProvs map[string]Provider
		if len(mv.Providers) == 0 {
			childProvs = providers
		} else {
			childProvs = make(map[string]Provider)
			for _, p := range mv.Providers {
				prov, ok := providers[p.InParent.String()]
				if !ok {
					continue
				}
				childProvs[p.InChild.String()] = prov
			}
		}

		// Set the full path of the child module. A child module must start with the path to the parent module,
		// then the word "module", then the module name, all separated by dots.
		var nextModPath string
		if modName != "" {
			nextModPath = fmt.Sprintf("%s.module.%s", modName, mk)
		} else {
			nextModPath = fmt.Sprintf("module.%s", mk)
		}

		qs, err := extractHCLModule(childProvs, parser, p, nextModPath, child, nextEvalCtx)
		if err != nil {
			return nil, err
		}
		queries = append(queries, qs...)
	}
	if knownProvider == false {
		return nil, ErrNoKnownProvider
	}

	return queries, nil
}

// getEvalCtx returns the evaluation context of the given module with variable values set.
func getEvalCtx(mod *configs.Module, vars map[string]cty.Value) *hcl.EvalContext {
	// Set default values for undefined variables.
	if vars == nil {
		vars = make(map[string]cty.Value)
	}
	for vk, vv := range mod.Variables {
		if _, ok := vars[vk]; !ok {
			vars[vk] = vv.Default
		}
	}

	// Initialize the evaluation context that will be used by the Terraform parser to fill in values
	// of variables. E.g. the `var` element contains variable values accessible from HCL using `var.*`.
	evalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var": cty.ObjectVal(vars),
		},
	}

	// Set values of locals.
	lm := make(map[string]cty.Value)
	for lk, lv := range mod.Locals {
		val, diags := lv.Expr.Value(evalCtx)
		if diags != nil && diags.HasErrors() {
			continue
		}
		lm[lk] = val
	}
	evalCtx.Variables["local"] = cty.ObjectVal(lm)

	return evalCtx
}

// getBodyJSON gets all the variables in a JSON format of the actual representation
func getBodyJSON(b *hclsyntax.Body, evalCtx *hcl.EvalContext) map[string]interface{} {
	links := make(map[string]interface{})
	// Each attribute of the body is casted to the correct type and placed into the links map.
	for attrk, attrv := range b.Attributes {
		val, _ := attrv.Expr.Value(evalCtx)
		if !val.IsKnown() {
			continue
		}
		switch val.Type() {
		case cty.String:
			links[attrk] = val.AsString()
		case cty.Number:
			f, _ := val.AsBigFloat().Float64()
			links[attrk] = f
		case cty.Bool:
			links[attrk] = val.True()
		}
	}
	for _, block := range b.Blocks {
		cfg := getBodyJSON(block.Body, evalCtx)
		// We continue to not add empty information to the config
		// so it's clean and only has required information
		if len(cfg) == 0 {
			continue
		}
		if _, ok := links[block.Type]; !ok {
			links[block.Type] = make([]interface{}, 0)
		}
		links[block.Type] = append(links[block.Type].([]interface{}), cfg)
	}

	return links
}

// getHCLProviders extracts provider configurations from the module and initializes the providers using the
// providerInitializers slice. The resulting map of aliases to instantiated providers is then returned.
func getHCLProviders(mod *configs.Module, evalCtx *hcl.EvalContext, providerInitializers []ProviderInitializer) (map[string]Provider, error) {
	piMap := make(map[string]ProviderInitializer)
	for _, pi := range providerInitializers {
		for _, name := range pi.MatchNames {
			piMap[name] = pi
		}
	}

	providers := make(map[string]Provider)
	for pk, pv := range mod.ProviderConfigs {
		pi, ok := piMap[pv.Name]
		if !ok {
			continue
		}

		body, ok := pv.Config.(*hclsyntax.Body)
		if !ok {
			return nil, fmt.Errorf("bad body")
		}

		cfg := getBodyJSON(body, evalCtx)
		values := make(map[string]string)
		for k, v := range cfg {
			if s, ok := v.(string); ok {
				values[k] = s
			}
		}

		prov, err := pi.Provider(values)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize provider: %w", err)
		}
		providers[pk] = prov
	}

	return providers, nil
}

// joinPath joins two directory paths together, unless target is absolute, in which case it is returned instead.
func joinPath(source, target string) string {
	if path.IsAbs(target) {
		return target
	}
	return path.Join(source, target)
}
