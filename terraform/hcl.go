package terraform

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/getmodules"
	"github.com/hashicorp/terraform/lang"
	"github.com/hashicorp/terraform/registry"
	"github.com/hashicorp/terraform/registry/regsrc"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/cycloidio/terracost/log"
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
	log.Logger.Debug("hcl: Loading module", "path", modPath)
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

	pns := make([]string, 0, 0)
	for pn := range providers {
		pns = append(pns, pn)
	}
	log.Logger.Debug("hcl: Providers found", "providers", pns)

	queries, err := extractHCLModule(fs, providers, parser, modPath, "", mod, 1, evalCtx, u)
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
func extractHCLModule(fs afero.Fs, providers map[string]Provider, parser *configs.Parser, modPath, modName string, mod *configs.Module, mcount int, evalCtx *hcl.EvalContext, u usage.Usage) ([]query.Resource, error) {
	queries := make([]query.Resource, 0, len(mod.ManagedResources))

	rss := make(map[string]Resource)
	for rk, rv := range mod.ManagedResources {

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
					log.Logger.Error("hcl: could not get value from ForEach", err, err)
				} else {
					v, ok := convertCtyValue("", nil, fev)
					if !ok {
						// TODO: Return an error?
					}
					each = v.(map[string]interface{})
				}
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
				types := make(map[string]cty.Type)
				switch vt := v.(type) {
				case map[string]interface{}:
					for kv, vv := range vt {
						it, err := goTypeToCty(vv)
						if err != nil {
							continue
						}
						ctyv, err := gocty.ToCtyValue(vv, it)
						if err != nil {
							continue
						}
						vals[kv] = ctyv
						types[kv] = it
					}
				default:
					it, err := goTypeToCty(vt)
					if err != nil {
						continue
					}
					ctyv, err := gocty.ToCtyValue(vt, it)
					if err != nil {
						continue
					}
					vals[k] = ctyv
					types[k] = it
				}

				ctyv, err := gocty.ToCtyValue(vals, cty.Object(types))
				if err != nil {
					continue
				}

				// TODO: potentially overwrite the rk with the [k]
				evalCtx.Variables["each"] = cty.ObjectVal(map[string]cty.Value{"value": ctyv})
			}
			cfg := getBodyJSON(modName, body, evalCtx)
			// We delete the `for_each` key as we do not need it
			delete(cfg, "for_each")

			for mc := 0; mc < mcount; mc++ {
				nrk := rk
				if modName != "" {
					if mcount > 1 {
						nrk = fmt.Sprintf("%s[%d].%s", modName, mc, rk)
					} else {
						nrk = fmt.Sprintf("%s.%s", modName, rk)
					}
				}
				// k is empty when there is not 'for_each'
				if k == "" {
					// Assume this is a single instance of this resource unless the cfg contains the "count" parameter.
					count := 1
					if c, ok := cfg["count"]; ok {
						if cf, ok := c.(float64); ok {
							count = int(cf)
						}
					}

					if count != 1 {
						log.Logger.Debug("hcl: Found count on resource", "count", count, "resource", nrk)
					}
					for i := 0; i < count; i++ {
						addr := nrk
						if count > 1 {
							addr = fmt.Sprintf("%s[%d]", nrk, i)
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
							log.Logger.Debug("hcl: Found resource", "resource", rss[addr])
						}
					}
				} else {
					addr := nrk
					addr = fmt.Sprintf("%s[%s]", nrk, k)

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
						log.Logger.Debug("hcl: Found resource", "resource", rss[addr])
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

		log.Logger.Debug("hcl: Found child module", "path", p)
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
			log.Logger.Debug("hcl: Was a remote module, pulled to new path", "path", p, "source_addr", maddr, "source_addr_raw", dir)
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
			// TODO: Check if the attribute has variables
			// and if so read those first, add the values to the CTX
			// and then evaluate the current one with that new context
			// with the defined value

			if _, ok := vars[attr.Name]; ok {
				// This means it has been pulled before as a dependency
				continue
			}
			for _, vr := range attr.Expr.Variables() {
				v := string(hclwrite.TokensForTraversal(vr).Bytes())
				sv := strings.Split(v, ".")
				if val, ok := vars[sv[1]]; ok {
					appendToCtx(evalCtx, sv[0], sv[1], val)
				} else if sv[0] == "var" || sv[0] == "local" {
					depAttr, ok := body.Attributes[sv[1]]
					if ok {
						val, diags := depAttr.Expr.Value(evalCtx)
						if diags != nil && diags.HasErrors() {
							log.Logger.Error("hcl: Error on abstracting value for 'vars'", "name", depAttr.Name, "reason", diags.Error())
							continue
						}
						appendToCtx(evalCtx, sv[0], depAttr.Name, val)
						vars[depAttr.Name] = val
					}
				}
			}
			val, diags := attr.Expr.Value(evalCtx)
			if diags != nil && diags.HasErrors() {
				log.Logger.Error("hcl: Error on abstracting value for 'vars'", "name", attr.Name, "reason", diags.Error())
				continue
			}
			vars[attr.Name] = val
		}

		nextEvalCtx := getEvalCtx(child, vars, nil)

		// TODO: Check if this should use nextEvalCtx
		log.Logger.Debug("hcl: Fetching module count")
		mcfg := getBodyJSON(modName, body, nextEvalCtx)
		nmcount := 1
		if c, ok := mcfg["count"]; ok {
			if cf, ok := c.(float64); ok {
				nmcount = int(cf)
			}
		}
		log.Logger.Debug("hcl: End fetching module count")

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

		qs, err := extractHCLModule(fs, childProvs, parser, p, nextModPath, child, nmcount, nextEvalCtx, u)
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
	lvars := make(map[string]interface{})
	llocal := make(map[string]interface{})
	// Set default values for undefined variables.
	if vars == nil {
		vars = make(map[string]cty.Value)
	}
	// TODO: Set variables that are on the Module
	for vk, vv := range mod.Variables {
		if _, ok := vars[vk]; !ok {
			vars[vk] = vv.Default
			lv, ok := convertCtyValue("", nil, vv.Default)
			if ok {
				lvars[vk] = lv
			}
		}
		if iv, ok := inputs[vk]; ok {
			iv, it := convertGoTypesToExpectedCtyType(iv, vv.Type)
			ctyv, err := gocty.ToCtyValue(iv, it)
			if err != nil {
				log.Logger.Error("hcl: Error on abstracting value for 'input'", "key", vk, "reason", err.Error())
				// NOTE: There are some types that we don't how to
				// parse yet but we want to continue so we ignore
				// the error
				continue
			}
			vars[vk] = ctyv

			lv, ok := convertCtyValue("", nil, ctyv)
			if ok {
				lvars[vk] = lv
			}
			continue
		}
	}

	scope := lang.Scope{}
	// Initialize the evaluation context that will be used by the Terraform parser to fill in values
	// of variables. E.g. the `var` element contains variable values accessible from HCL using `var.*`.
	evalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var": cty.ObjectVal(vars),
		},
		Functions: scope.Functions(),
	}

	// Set values of locals.
	lm := make(map[string]cty.Value)
	for lk, lv := range mod.Locals {
		val, diags := lv.Expr.Value(evalCtx)
		if diags != nil && diags.HasErrors() {
			log.Logger.Error("hcl: Error on abstracting value for 'local'", "key", lk, "reason", diags.Error())
			continue
		}
		lm[lk] = val
		lv, ok := convertCtyValue("", nil, val)
		if ok {
			llocal[lk] = lv
		}
	}
	evalCtx.Variables["local"] = cty.ObjectVal(lm)

	log.Logger.Debug("hcl: New variables/locals found", "var", lvars, "local", llocal)

	return evalCtx
}

func appendToCtx(ctx *hcl.EvalContext, t, name string, v cty.Value) {
	vars := ctx.Variables[t]

	mvars := make(map[string]cty.Value)
	// TODO: Do a foreach or a func (val Value) ElementIterator() ElementIterator {
	iter := vars.ElementIterator()
	for iter.Next() {
		k, v := iter.Element()
		var key string
		if err := gocty.FromCtyValue(k, &key); err != nil {
			log.Logger.Error("hcl: Failed to get KEY from Context to append", "key", k, "reason", err.Error())
		}
		mvars[key] = v
	}
	mvars[name] = v

	ctx.Variables[t] = cty.ObjectVal(mvars)
}

// convertGoTypesToExpectedCtyType will take a GO value and a cty.Type and convert the GO value into the cty.Type as much
// as possible by trying to weak-type assertions
func convertGoTypesToExpectedCtyType(v interface{}, t cty.Type) (interface{}, cty.Type) {
	var (
		nv interface{}
		nt cty.Type = t
	)
	// We check if the expected type on the module
	// matches the type we have on the inputs, if
	// not we have to convert the type on the inputs
	// to the one expected on the module definition.
	switch t {
	case cty.String:
		var ok bool
		nv, ok = v.(string)
		if !ok {
			nv = fmt.Sprint(v)
		}
	case cty.Number:
		switch t := v.(type) {
		case float64:
			// This is the right type numbers are parsed
			nv = v
		case string:
			niv, err := strconv.Atoi(t)
			if err != nil {
				// NOTE: This is so it does not
				// break when parsing
				nv = 0
				break
			}
			nv = niv
		case bool:
			nv = 1
			if !t {
				nv = 0
			}
		default:
			// NOTE: This is so it does not
			// break when parsing
			nv = false
		}
	case cty.Bool:
		switch t := v.(type) {
		case bool:
			// This is the right type
			nv = v
		case string:
			nv = false
			if t == "true" {
				nv = true
			}
		case float64:
			nv = false
			if t > 0.0 {
				nv = true
			}
		default:
			// NOTE: This is so it does not
			// break when parsing
			nv = false
		}
	case cty.DynamicPseudoType:
		// NOTE: This is when we do not actually know which is
		// the type of the value
		ct, err := goTypeToCty(v)
		if err != nil {
			log.Logger.Error("hcl: Error on abstracting DynamicPseudoType", "error", err.Error())
			return nil, nt
		}
		nv, _ = convertGoTypesToExpectedCtyType(v, ct)
		nt = ct
	default:
		// Here we check for complex types
		// TODO: Some of this types I don't have examples
		// of or may not even be possible for them to happen.
		// Gonna leave the IF statements so we know them
		if v == nil {
			return nil, nt
		}
		if t.IsTupleType() {
		} else if t.IsObjectType() {
			cfg := make(map[string]interface{})
			vm, ok := v.(map[string]interface{})
			if !ok {
				return nil, nt
			}
			for vk, vv := range vm {
				if t.HasAttribute(vk) {
					at := t.AttributeType(vk)
					cfg[vk], _ = convertGoTypesToExpectedCtyType(vv, at)
				} else {
					// If the recieving object does not have the expected
					cfg[vk], _ = convertGoTypesToExpectedCtyType(vv, cty.String)
				}
			}
			nv = cfg
		} else if t.IsMapType() {
			// A map is a list of the same type
			mv := make(map[string]interface{})
			et := t.MapElementType()
			vm, ok := v.(map[string]interface{})
			if !ok {
				return nil, nt
			}
			for vk, vv := range vm {
				mv[vk], _ = convertGoTypesToExpectedCtyType(vv, *et)
			}
			nv = mv
			break
		} else if t.IsListType() {
			lv := make([]interface{}, 0, 0)
			et := t.ListElementType()
			va, ok := v.([]interface{})
			if !ok {
				return nil, nt
			}
			for _, vv := range va {
				nnv, _ := convertGoTypesToExpectedCtyType(vv, *et)
				lv = append(lv, nnv)
			}
			nv = lv
			break
		} else {
		}
	}
	return nv, nt
}

// getBodyJSON gets all the variables in a JSON format of the actual representation and the references it may have
func getBodyJSON(modulePrefix string, b *hclsyntax.Body, evalCtx *hcl.EvalContext) map[string]interface{} {
	cfg := make(map[string]interface{})
	// Each attribute of the body is casted to the correct type and placed into the cfg map.
	for attrk, attrv := range b.Attributes {
		val, diags := attrv.Expr.Value(evalCtx)
		if diags != nil && diags.HasErrors() && !val.IsKnown() && len(attrv.Expr.Variables()) == 0 {
			log.Logger.Error("hcl: Error on abstracting value for 'attribute'", "name", attrk, "reason", diags.Error())
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
	if val.IsNull() {
		return nil, true
	}
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

func goTypeToCty(t interface{}) (cty.Type, error) {
	switch tv := t.(type) {
	case int, int32, int64, float64, float32:
		return cty.Number, nil
	case string:
		return cty.String, nil
	case bool:
		return cty.Bool, nil
	case map[string]interface{}:
		tm := make(map[string]cty.Type)
		for k, v := range tv {
			ct, err := goTypeToCty(v)
			if err != nil {
				return cty.Type{}, err
			}
			tm[k] = ct
		}
		return cty.Object(tm), nil
	case []interface{}:
		if len(tv) == 0 {
			return cty.Type{}, fmt.Errorf("empty array found cannot deduce internal type")
		}
		first := tv[0]
		ft, err := goTypeToCty(first)
		if err != nil {
			return cty.Type{}, err
		}
		return cty.List(ft), nil
	default:
		return cty.Type{}, fmt.Errorf("failed to convert type %T to CTY type", t)
	}
}
