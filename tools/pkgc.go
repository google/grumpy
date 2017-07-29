// pkgc is a tool for generating wrappers for Go packages imported by Grumpy
// programs.
//
// usage: pkgc PACKAGE
//
// Where PACKAGE is the full Go package name. Generated code is dumped to
// stdout. Packages generated in this way can be imported by Grumpy programs
// using string literal import syntax, e.g.:
//
// import "__go__/encoding/json"
//
// Or:
//
// from "__go__/time" import Duration

package main

import (
	"bytes"
	"fmt"
	"go/constant"
	"go/importer"
	"go/types"
	"math"
	"os"
	"path"
)

const packageTemplate = `package %[1]s
import (
	"grumpy"
	"reflect"
	mod %[2]q
)
func fun(f *grumpy.Frame, _ []*grumpy.Object) (*grumpy.Object, *grumpy.BaseException) {
%[3]s
	return nil, nil
}
var Code = grumpy.NewCode("<module>", %[2]q, nil, 0, fun)
func init() {
	grumpy.RegisterModule("__go__/%[2]s", Code)
}
`

const typeTemplate = `	if true {
		var x mod.%[1]s
		if o, raised := grumpy.WrapNative(f, reflect.ValueOf(x)); raised != nil {
			return nil, raised
		} else if raised = f.Globals().SetItemString(f, %[1]q, o.Type().ToObject()); raised != nil {
			return nil, raised
		}
	}
`

const varTemplate = `	if o, raised := grumpy.WrapNative(f, reflect.ValueOf(%[1]s)); raised != nil {
		return nil, raised
	} else if raised = f.Globals().SetItemString(f, %[2]q, o); raised != nil {
		return nil, raised
	}
`

func getConst(name string, v constant.Value) string {
	format := "%s"
	switch v.Kind() {
	case constant.Int:
		if constant.Sign(v) >= 0 {
			if i, exact := constant.Uint64Val(v); exact {
				if i > math.MaxInt64 {
					format = "uint64(%s)"
				}
			} else {
				format = "float64(%s)"
			}
		}
	case constant.Float:
		format = "float64(%s)"
	}
	return fmt.Sprintf(format, name)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprint(os.Stderr, "usage: pkgc PACKAGE")
		os.Exit(1)
	}
	pkgPath := os.Args[1]
	pkg, err := importer.Default().Import(pkgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to import: %q: %v\n", pkgPath, err)
		os.Exit(2)
	}
	var buf bytes.Buffer
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		o := scope.Lookup(name)
		if !o.Exported() {
			continue
		}
		switch x := o.(type) {
		case *types.TypeName:
			if types.IsInterface(x.Type()) {
				continue
			}
			buf.WriteString(fmt.Sprintf(typeTemplate, name))
		case *types.Const:
			expr := getConst("mod." + name, x.Val())
			buf.WriteString(fmt.Sprintf(varTemplate, expr, name))
		default:
			expr := "mod." + name
			buf.WriteString(fmt.Sprintf(varTemplate, expr, name))
		}
	}
	fmt.Printf(packageTemplate, path.Base(pkgPath), pkgPath, buf.Bytes())
}
