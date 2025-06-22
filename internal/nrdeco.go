package internal

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
	"unicode"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

//go:embed nrdeco.tmpl
var nrdecoTemplate string

func Generate(_ context.Context, source, dest, version string) ([]byte, error) {
	tpl, err := parseTemplate()
	if err != nil {
		return nil, err
	}

	nodes, err := parser.ParseFile(token.NewFileSet(), source, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", source, err)
	}

	pkgs, err := packages.Load(&packages.Config{}, path.Dir(source))
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}
	pkgIdx := slices.IndexFunc(pkgs, func(pkg *packages.Package) bool {
		return !strings.HasSuffix(pkg.Name, "_test")
	})
	if pkgIdx == -1 {
		return nil, fmt.Errorf("non-test package not found in '%s'", path.Dir(source))
	}
	f := File{
		Version:     version,
		PackageName: filepath.Base(pkgs[pkgIdx].ID),
		Imports: map[string]Package{
			"os": {
				Path: "os",
			},
			"strings": {
				Path: "strings",
			},
			"newrelic": {
				Path: "github.com/newrelic/go-agent/v3/newrelic",
			},
		},
	}

	absSource, err := filepath.Abs(source)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path of source file %s: %w", source, err)
	}
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path of destination file %s: %w", dest, err)
	}
	if filepath.Dir(absSource) != filepath.Dir(absDest) {
		f.OriginalPackageName = f.PackageName
		destPkgs, _ := packages.Load(&packages.Config{}, path.Dir(source))
		switch {
		case len(destPkgs) > 0:
			f.PackageName = filepath.Base(destPkgs[0].ID)
		default:
			destDir, _ := filepath.Split(dest)
			f.PackageName = filepath.Base(destDir)
		}

		f.Imports[pkgs[pkgIdx].Name] = Package{
			Path: pkgs[pkgIdx].ID,
		}
		f.DifferInDest = true
	}
	visitor := newVisitor(&f, nodes.Imports)
	astutil.Apply(nodes, nil, visitor.Visit)
	if visitor.err != nil {
		return nil, fmt.Errorf("error while visiting AST: %w", visitor.err)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, &f)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Visitor visits each *astutil.Cursor to find interfaces and their methods
type Visitor struct {
	f               *File
	importSpecs     []*ast.ImportSpec
	importPathCache map[string]string
	err             error
}

func (v *Visitor) Visit(c *astutil.Cursor) bool {
	typeSpec, ok := c.Node().(*ast.TypeSpec)
	if !ok {
		return true
	}
	interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return true
	}
	t := Interface{
		Name:    typeSpec.Name.Name,
		Methods: make([]Method, 0, len(interfaceType.Methods.List)),
	}
	for _, field := range interfaceType.Methods.List {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		method := Method{
			Name:    field.Names[0].Name,
			Params:  make([]Value, 0, len(funcType.Params.List)),
			Returns: make([]Value, 0, len(funcType.Results.List)),
		}
		for _, param := range funcType.Params.List {
			val, err := v.valueFromExpr(param.Type)
			if err != nil {
				v.err = err
				return false
			}
			method.Params = append(method.Params, *val)
		}
		for _, result := range funcType.Results.List {
			val, err := v.valueFromExpr(result.Type)
			if err != nil {
				v.err = err
				return false
			}
			method.Returns = append(method.Returns, *val)
		}
		if !method.Params.BeGenerated() {
			continue
		}
		t.Methods = append(t.Methods, method)
	}
	if len(t.Methods) != 0 {
		v.f.Interfaces = append(v.f.Interfaces, t)
	}
	return true
}

func (v *Visitor) getImportPath(pkg string) string {
	if p, ok := v.f.Imports[pkg]; ok {
		return p.Path
	}
	importIdx := slices.IndexFunc(v.importSpecs, func(importSpec *ast.ImportSpec) bool {
		if importSpec.Name == nil {
			return false
		}
		return importSpec.Name.Name == pkg
	})

	if importIdx == -1 {
		importIdx = slices.IndexFunc(v.importSpecs, func(importSpec *ast.ImportSpec) bool {
			return strings.HasSuffix(importSpec.Path.Value, fmt.Sprintf(`%s"`, pkg))
		})
	}
	if importIdx == -1 {
		return ""
	}
	if v.importSpecs[importIdx].Path == nil {
		return ""
	}

	impPath := v.importSpecs[importIdx].Path.Value
	p := strings.Trim(impPath, `"`)
	v.f.Imports[pkg] = Package{
		Path: p,
	}
	return p
}

func (v *Visitor) valueFromExpr(t ast.Expr) (*Value, error) {
	switch t := t.(type) {
	case *ast.ArrayType:
		element, err := v.valueFromExpr(t.Elt)
		if err != nil {
			return nil, err
		}
		return &Value{
			Type:    typeSlice,
			Element: element,
		}, nil
	case *ast.MapType:
		k, err := v.valueFromExpr(t.Key)
		if err != nil {
			return nil, err
		}
		v, err := v.valueFromExpr(t.Value)
		if err != nil {
			return nil, err
		}
		return &Value{
			Type:    typeMap,
			Key:     k.Type,
			Element: v,
		}, nil
	case *ast.StarExpr:
		el, err := v.valueFromExpr(t.X)
		if err != nil {
			return nil, err
		}
		return &Value{
			Type:    typePointer,
			Element: el,
		}, nil
	case *ast.ChanType:
		el, err := v.valueFromExpr(t.Value)
		if err != nil {
			return nil, err
		}
		switch t.Dir {
		case ast.SEND:
			return &Value{
				Type:    typeChannelSend,
				Element: el,
			}, nil
		case ast.RECV:
			return &Value{
				Type:    typeChannelReceive,
				Element: el,
			}, nil
		}
		return &Value{
			Type:    typeChannel,
			Element: el,
		}, nil
	case *ast.FuncType:
		params := make(Params, 0, len(t.Params.List))
		for _, field := range t.Params.List {
			for _, p := range field.Names {
				val, err := v.valueFromExpr(p)
				if err != nil {
					return nil, err
				}
				params = append(params, *val)
			}
		}
		rets := make(Returns, 0, len(t.Results.List))
		for _, field := range t.Results.List {
			val, err := v.valueFromExpr(field.Type)
			if err != nil {
				return nil, err
			}
			rets = append(rets, *val)
		}
		return &Value{
			Type:    typeFunction,
			Params:  params,
			Returns: rets,
		}, nil
	case *ast.Ident:
		if v.f.DifferInDest && unicode.IsUpper(rune(t.Name[0])) {
			// If an identifier begins with an uppercase letter,
			// it is assumed to be of the type defined in the original package.
			importPath := v.getImportPath(v.f.OriginalPackageName)
			if importPath == "" {
				return nil, fmt.Errorf("import '%s' not found", t.Name)
			}
			return &Value{
				Type:    t.Name,
				Package: &Package{Path: importPath},
			}, nil
		}
		return &Value{
			Type: t.Name,
		}, nil
	case *ast.SelectorExpr:
		pkgName := t.X.(*ast.Ident).Name
		importPath := v.getImportPath(pkgName)
		if importPath == "" {
			return nil, fmt.Errorf("import '%s' not found", pkgName)
		}
		return &Value{
			Type: t.Sel.Name,
			Package: &Package{
				Path: importPath,
			},
		}, nil
	case *ast.Ellipsis:
		el, err := v.valueFromExpr(t.Elt)
		if err != nil {
			return nil, err
		}
		return &Value{
			Type:    typeVariadic,
			Element: el,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported type expression: %T", t)
	}
}

func newVisitor(f *File, importSpecs []*ast.ImportSpec) *Visitor {
	return &Visitor{
		f:               f,
		importSpecs:     importSpecs,
		importPathCache: make(map[string]string),
	}
}

func parseTemplate() (*template.Template, error) {
	tpl, err := template.New("nrdeco").Parse(nrdecoTemplate)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}
