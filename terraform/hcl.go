package terraform

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/getmodules"
	"github.com/hashicorp/terraform/registry"
	"github.com/hashicorp/terraform/registry/regsrc"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/cycloidio/terracost/query"
	"github.com/cycloidio/terracost/usage"
	"github.com/cycloidio/terracost/util"
)

const (
	// hclRefPrefix is used to prefix all the attributes of the resource that are
	// references to another resource, so then on the second iteration we can
	// replace it if it has the value
	hclRefPrefix = "_tc_ref"
)

// ExtractQueriesFromHCL returns the resources found in the module identified by the modPath.
func ExtractQueriesFromHCL(fs afero.Fs, providerInitializers []ProviderInitializer, modPath string, u usage.Usage, inputs map[string]interface{}) ([]query.Resource, string, error) {
	parser := configs.NewParser(fs)
	mod, diags := parser.LoadConfigDir(modPath)
	if diags.HasErrors() {
		return nil, "", fmt.Errorf(diags.Error())
	}

	evalCtx := getEvalCtx(mod, nil, inputs)

	modules := make([]string, 0, 0)
	for k := range mod.ModuleCalls {
		modules = append(modules, k)
	}
	sort.Strings(modules)

	modName := strings.Join(modules, ", ")

	providers, err := getHCLProviders(mod, evalCtx, providerInitializers)
	if err != nil {
		return nil, modName, err
	}

	queries, err := extractHCLModule(fs, providers, parser, modPath, "", mod, evalCtx, u)
	if err != nil {
		return nil, modName, err
	}

	err = validateProviders(queries, providers)
	if err != nil {
		return nil, modName, err
	}

	return queries, modName, nil
}

// extractHCLModule returns the resources found in the provided module.
func extractHCLModule(fs afero.Fs, providers map[string]Provider, parser *configs.Parser, modPath, modName string, mod *configs.Module, evalCtx *hcl.EvalContext, u usage.Usage) ([]query.Resource, error) {
	queries := make([]query.Resource, 0, len(mod.ManagedResources))

	rss := make(map[string]Resource)
	for rk, rv := range mod.ManagedResources {
		if modName != "" {
			rk = fmt.Sprintf("%s.%s", modName, rk)
		}

		providerKey := rv.Provider.Type
		if rv.ProviderConfigRef != nil {
			providerKey = rv.ProviderConfigRef.String()
		}
		provider := providers[providerKey]

		each := make(map[string]interface{})
		if rv.ForEach != nil {
			if fe, ok := rv.ForEach.(*hclsyntax.ForExpr); ok {
				fev, err := fe.Value(evalCtx)
				if err != nil {
					return nil, fmt.Errorf("could not get value from ForEach: %w", err)
				}
				v, ok := convertCtyValue("", nil, fev)
				if !ok {
					// TODO: Return an error?
				}
				each = v.(map[string]interface{})
			}
		}
		// When we only have to do it once
		if len(each) == 0 {
			each[""] = nil
		}
		// Parse the HCL body of the resource block and evaluate it. The JSON (in the form of map[string]interface{} type)
		// is then placed into the cfg.
		body, ok := rv.Config.(*hclsyntax.Body)
		if !ok {
			return nil, fmt.Errorf("invalid resource configuration body")
		}
		for k, v := range each {
			if v != nil {
				delete(evalCtx.Variables, "each")
				vals := make(map[string]cty.Value)
				for kv, vv := range v.(map[string]interface{}) {
					it, err := gocty.ImpliedType(vv)
					if err != nil {
						continue
					}
					ctyv, err := gocty.ToCtyValue(vv, it)
					if err != nil {
						continue
					}
					vals[kv] = ctyv
				}

				//each := map[string]interface{}{"value": vals}
				vt, err := gocty.ImpliedType(vals)
				if err != nil {
					continue
				}
				ctyv, err := gocty.ToCtyValue(vals, vt)
				if err != nil {
					continue
				}

				// TODO: potentially overwrite the rk with the [k]
				evalCtx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{"value": ctyv})
			}
			cfg := getBodyJSON(modName, body, evalCtx)
			// We delte the `for_each` key as we do not need it
			delete(cfg, "for_each")

			if k == "" {
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

					// Only retrieve resource if the provider is valid. If it's not, the comps will be nil, which signifies
					// that the resource was "skipped" from estimation.
					if provider != nil {
						rss[addr] = Resource{
							Address:      addr,
							Index:        i,
							Mode:         "managed",
							Type:         rv.Type,
							Name:         rv.Name,
							ProviderName: rv.Provider.Type,
							Values:       cfg,
						}
					}
				}
			} else {
				addr := rk
				addr = fmt.Sprintf("%s[%s]", rk, k)

				// Only retrieve resource if the provider is valid. If it's not, the comps will be nil, which signifies
				// that the resource was "skipped" from estimation.
				if provider != nil {
					rss[addr] = Resource{
						Address:      addr,
						Mode:         "managed",
						Type:         rv.Type,
						Name:         rv.Name,
						ProviderName: rv.Provider.Type,
						Values:       cfg,
					}
				}
			}
		}
	}

	for kr, r := range rss {
		for k, v := range r.Values {
			switch vv := v.(type) {
			case string:
				if strings.HasPrefix(vv, hclRefPrefix) {
					vv = strings.Replace(vv, hclRefPrefix, "", -1)
					vsp := strings.Split(vv, ".")
					attr := vsp[len(vsp)-1]
					refk := strings.Join(vsp[0:len(vsp)-1], ".")
					if refr, ok := rss[refk]; ok {
						if a, ok := refr.Values[attr]; ok {
							rss[kr].Values[k] = a
							continue
						}
					}
					rss[kr].Values[k] = refk
				}
			case []interface{}:
				vals := make([]interface{}, 0, 0)
				for _, v := range vv {
					switch vv := v.(type) {
					case string:
						if strings.HasPrefix(vv, hclRefPrefix) {
							vv = strings.Replace(vv, hclRefPrefix, "", -1)
							vsp := strings.Split(vv, ".")
							attr := vsp[len(vsp)-1]
							refk := strings.Join(vsp[0:len(vsp)-1], ".")
							if refr, ok := rss[refk]; ok {
								if a, ok := refr.Values[attr]; ok {
									vals = append(vals, a)
									continue
								}
							}
							vals = append(vals, refk)
						}
					default:
						vals = append(vals, v)
					}
				}
				rss[kr].Values[k] = vals
			}
		}
		r.Values[usage.Key] = u.GetUsage(r.Type)
		provider := providers[r.ProviderName]
		queries = append(queries, query.Resource{
			Address:    r.Address,
			Type:       r.Type,
			Provider:   r.ProviderName,
			Components: provider.ResourceComponents(rss, r),
		})
	}

	// Recursively extract resources from all child module calls.
	for mk, mv := range mod.ModuleCalls {
		p := joinPath(modPath, mv.SourceAddr.String())

		// EntersNewPackage checks if the module is a local
		// one or a Remote one.
		if mv.EntersNewPackage() {
			dir, err := installModule(fs, mv)
			if err != nil {
				return nil, fmt.Errorf("failed to install remote module: %w", err)
			}
			maddr, err := addrs.ParseModuleSource(dir)
			if err != nil {
				return nil, fmt.Errorf("failed parse module source %q: %w", dir, err)
			}
			// We ignore the SourceAddrRange for now
			// as it's the reference to where the Source is
			// located on the file which we do not need anymore
			mv.SourceAddr = maddr
			mv.SourceAddrRaw = dir
			p = dir
		}

		child, diags := parser.LoadConfigDir(p)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to load config dir: %w", diags)
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

		nextEvalCtx := getEvalCtx(child, vars, nil)

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

		qs, err := extractHCLModule(fs, childProvs, parser, p, nextModPath, child, nextEvalCtx, u)
		if err != nil {
			return nil, err
		}
		queries = append(queries, qs...)
	}

	return queries, nil
}

// installModule will walk the ModuleCall and download any remote module
func installModule(fs afero.Fs, mc *configs.ModuleCall) (string, error) {
	dir, err := os.MkdirTemp("", "terracost")
	if err != nil {
		return "", fmt.Errorf("failed to create a temp dir: %w", err)
	}
	// We remove it to create just the folder and then
	// we pass the 'dir' path to the Fetcher and we know
	// it's a valid unexistent dir. If we did not delete it
	// the 'FetchPackage' would try to 'git pull' instead
	// of 'git clone'
	os.RemoveAll(dir)
	fetcher := getmodules.NewPackageFetcher()
	src := mc.SourceAddr.String()
	ctx := context.Background()
	if strings.Contains(src, "registry") {
		rm, err := regsrc.ParseModuleSource(src)
		if err != nil {
			return "", fmt.Errorf("failed to parse module source %s: %w", src, err)
		}
		// we initialize the client with default values
		// so everything is nil
		rc := registry.NewClient(nil, nil)
		noVersion := ""
		ml, err := rc.ModuleLocation(ctx, rm, noVersion)
		if err != nil {
			return "", fmt.Errorf("failed to locate module %s: %w", src, err)
		}
		src = ml
	}
	err = fetcher.FetchPackage(ctx, dir, src)
	if err != nil {
		return "", fmt.Errorf("failed to download the module %q: %w", src, err)
	}

	// If the dir already exists we assume it's the right Fs
	// so no need to copy the content
	if ok, _ := afero.DirExists(fs, dir); ok {
		return dir, nil
	}

	// Once everything is copied we can remove it
	defer os.RemoveAll(dir)

	err = fs.MkdirAll(dir, 0700)
	if err != nil {
		return "", fmt.Errorf("could not create path %q: %w", dir, err)
	}
	err = util.FromOSToAfero(fs, dir, dir)
	if err != nil {
		return "", fmt.Errorf("failed to copy the module to %s: %w", dir, err)
	}
	return dir, nil
}

// getEvalCtx returns the evaluation context of the given module with variable values set.
func getEvalCtx(mod *configs.Module, vars map[string]cty.Value, inputs map[string]interface{}) *hcl.EvalContext {
	// Set default values for undefined variables.
	if vars == nil {
		vars = make(map[string]cty.Value)
	}
	for vk, vv := range mod.Variables {
		if _, ok := vars[vk]; !ok {
			// If it has no value we do not set it
			if !vv.Default.IsNull() {
				vars[vk] = vv.Default
			}
		}
		if iv, ok := inputs[vk]; ok {
			ctyv, err := gocty.ToCtyValue(iv, vv.Type)
			if err != nil {
				continue
			}
			vars[vk] = ctyv
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

// getBodyJSON gets all the variables in a JSON format of the actual representation and the references it may have
func getBodyJSON(modulePrefix string, b *hclsyntax.Body, evalCtx *hcl.EvalContext) map[string]interface{} {
	cfg := make(map[string]interface{})
	// Each attribute of the body is casted to the correct type and placed into the cfg map.
	for attrk, attrv := range b.Attributes {
		val, _ := attrv.Expr.Value(evalCtx)
		if !val.IsKnown() && len(attrv.Expr.Variables()) == 0 {
			continue
		}
		vv, ok := convertCtyValue(modulePrefix, attrv.Expr.Variables(), val)
		if !ok {
			continue
		}
		cfg[attrk] = vv
	}
	for _, block := range b.Blocks {
		ncfg := getBodyJSON(modulePrefix, block.Body, evalCtx)
		// We continue to not add empty information to the config
		// so it's clean and only has required information
		if len(ncfg) == 0 {
			continue
		}
		if _, ok := cfg[block.Type]; !ok {
			cfg[block.Type] = make([]interface{}, 0)
		}
		cfg[block.Type] = append(cfg[block.Type].([]interface{}), ncfg)
	}

	return cfg
}

// convertCtyValue converts the value v to a normal type, the second
// return indicates if there is something to convert or no
func convertCtyValue(modulePrefix string, attrvars []hcl.Traversal, val cty.Value) (interface{}, bool) {
	switch val.Type() {
	case cty.String:
		// If the attribute points to a variable without a default value, "cty.UnknownVal(cty.String)" is returned
		// Skip Unknown values to avoid panic.
		// Other types such bool/Number without default value ends up here too
		if !val.IsKnown() || val.IsNull() {
			return nil, false
		}
		return val.AsString(), true
	case cty.Number:
		if !val.IsKnown() || val.IsNull() {
			return nil, false
		}
		f, _ := val.AsBigFloat().Float64()
		return f, true
	case cty.Bool:
		if !val.IsKnown() || val.IsNull() {
			return nil, false
		}
		return val.True(), true
	default:
		if val.Type().IsTupleType() {
			values := make([]interface{}, 0, 0)
			iter := val.ElementIterator()
			for iter.Next() {
				_, nval := iter.Element()
				switch nval.Type() {
				case cty.String:
					// If the attribute points to a variable without a default value, "cty.UnknownVal(cty.String)" is returned
					// Skip Unknow values to avoid panic.
					// Other types such bool/Number without default value ends up here too
					if !nval.IsKnown() || nval.IsNull() {
						continue
					}
					values = append(values, nval.AsString())
				case cty.Number:
					if !nval.IsKnown() || nval.IsNull() {
						continue
					}
					f, _ := nval.AsBigFloat().Float64()
					values = append(values, f)
				case cty.Bool:
					if !nval.IsKnown() || nval.IsNull() {
						continue
					}
					values = append(values, nval.True())
				default:
					vars := make([]string, 0, 0)
					for _, vr := range attrvars {
						v := string(hclwrite.TokensForTraversal(vr).Bytes())
						sv := strings.Split(v, ".")
						// The variables are also in here, if a variable
						// has not been interpolated, which means it has no default,
						// it'll be set as plain text and we don't want it
						if sv[0] == "var" || sv[0] == "each" {
							//if sv[0] == "var" {
							continue
						}
						// With this we remove the last element of the reference, which is the
						// attribute it's linking to from the resource
						//v = strings.Join(sv[0:len(sv)-1], ".")

						// We prefix this attribute with the hclRefPrefix so then
						// we know it's a reference and we can use it
						v = strings.Join(sv, ".")
						vars = append(vars, fmt.Sprintf("%s%s.%s", hclRefPrefix, modulePrefix, v))
						// TODO: Here is where the references are
					}
					if len(vars) != 0 {
						values = append(values, vars[0])
					}
				}
			}
			return values, true
		} else if val.Type().IsObjectType() {
			cfg := make(map[string]interface{})
			for k, v := range val.AsValueMap() {
				nv, ok := convertCtyValue(modulePrefix, nil, v)
				if !ok {
					continue
				}
				cfg[k] = nv
			}
			return cfg, true
		} else {
			vars := make([]string, 0, 0)
			for _, vr := range attrvars {
				v := string(hclwrite.TokensForTraversal(vr).Bytes())
				sv := strings.Split(v, ".")
				// The variables are also in here, if a variable
				// has not been interpolated, which means it has no default,
				// it'll be set as plain text and we don't want it
				if sv[0] == "var" || sv[0] == "each" {
					continue
				}
				// With this we remove the last element of the reference, which is the
				// attribute it's linking to from the resource
				//v = strings.Join(sv[0:len(sv)-1], ".")

				// We prefix this attribute with the hclRefPrefix so then
				// we know it's a reference and we can use it
				v = strings.Join(sv, ".")
				vars = append(vars, fmt.Sprintf("%s%s.%s", hclRefPrefix, modulePrefix, v))
			}
			if len(vars) != 0 {
				return vars[0], true
			}
		}
	}
	return nil, false
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

		cfg := getBodyJSON("", body, evalCtx)
		values := make(map[string]interface{})
		for k, v := range cfg {
			values[k] = v
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
