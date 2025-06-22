package internal

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// File represents a Go source file
type File struct {
	// Version represents the version of nrdeco used to generate this file.
	Version     string
	PackageName string
	Imports     map[string]Package
	Interfaces  []Interface
}

// StringOfImports returns a string representation of the imports in the file, sorted by package path
func (f *File) StringOfImports() string {
	v := make([]string, 0, len(f.Imports))
	for _, k := range slices.Sorted(maps.Keys(f.Imports)) {
		v = append(v, fmt.Sprintf("\t\"%s\"", f.Imports[k].Path))
	}
	return strings.Join(v, "\n")
}

// Interface represents a type
type Interface struct {
	Name    string
	Methods []Method
}

// Method represents a method
type Method struct {
	Name    string
	Params  Params
	Returns Returns
}

type Imports []Package

// Signature returns the method signature in the format "MethodName(ctx context.Context, arg1 Arg1Type, arg2 Arg2Type) (_ Return0Type, _ Return1Type)".
func (m *Method) Signature() string {
	if len(m.Returns) == 0 {
		return fmt.Sprintf("%s(%s)", m.Name, m.Params.Signature())
	}
	if len(m.Returns) == 1 {
		return fmt.Sprintf("%s(%s) %s", m.Name, m.Params.Signature(), m.Returns[0].StringOfType())
	}
	rets := make([]string, 0, len(m.Returns))
	for _, ret := range m.Returns {
		rets = append(rets, ret.StringOfType())
	}
	return fmt.Sprintf("%s(%s) (%s)", m.Name, m.Params.Signature(), strings.Join(rets, ", "))
}

// Params represents the method parameters
type Params []Value

// Signature returns the method parameters in the format "ctx context.Context, arg1 Arg1Type, arg2 Arg2Type".
func (p *Params) Signature() string {
	var v []string
	for i, param := range *p {
		if param.IsContext() {
			v = append(v, fmt.Sprintf("ctx %s", param.StringOfType()))
			continue
		}
		v = append(v, fmt.Sprintf("arg%d %s", i, param.StringOfType()))
	}
	return strings.Join(v, ", ")
}

// Call returns the method parameters in the format "ctx, arg1, arg2".
func (p *Params) Call() string {
	var v []string
	for i, param := range *p {
		if param.IsContext() {
			v = append(v, "ctx")
			continue
		}
		if param.Type == typeVariadic {
			v = append(v, fmt.Sprintf("arg%d...", i))
			continue
		}
		v = append(v, fmt.Sprintf("arg%d", i))
	}
	return strings.Join(v, ", ")
}

// BeGenerated checks if the Params slice contains follow parameters that should be generated.
//
// - context.Context
func (p *Params) BeGenerated() bool {
	if len(*p) == 0 {
		return false
	}
	return slices.IndexFunc(*p, func(param Value) bool {
		return param.IsContext()
	}) >= 0
}

// Returns represents the method return values
type Returns []Value

// Signature returns the method return values in the format "_ Return0Type, _ Return1Type".
func (r *Returns) Signature() string {
	var v []string
	for _, ret := range *r {
		v = append(v, fmt.Sprintf("_ %s", ret.StringOfType()))
	}
	return strings.Join(v, ", ")
}

const (
	typePointer        = "*"
	typeSlice          = "[]"
	typeMap            = "map"
	typeChannel        = "chan"
	typeChannelReceive = "<-chan"
	typeChannelSend    = "chan<-"
	typeFunction       = "func"
	typeVariadic       = "..."
	typeContext        = "Context"
)

// Value represents a parameter or return value.
type Value struct {
	Package *Package
	Type    string
	Key     string
	Element *Value
	Params  Params
	Returns Returns
}

// StringOfType returns the string representation of the value's type
func (v *Value) StringOfType() string {
	switch v.Type {
	case typeSlice, typePointer, typeVariadic, typeChannel:
		return fmt.Sprintf("%s%s", v.Type, v.Element.StringOfType())
	case typeChannelReceive, typeChannelSend:
		return fmt.Sprintf("%s %s", v.Type, v.Element.StringOfType())
	case typeMap:
		return fmt.Sprintf("map[%s]%s", v.Key, v.Element.StringOfType())
	case typeFunction:
		if len(v.Returns) == 0 {
			return fmt.Sprintf("func(%s)", v.Params.Signature())
		}
		if len(v.Returns) == 1 {
			return fmt.Sprintf("func(%s) %s", v.Params.Signature(), v.Returns[0].StringOfType())
		}
		rets := make([]string, 0, len(v.Returns))
		for _, ret := range v.Returns {
			rets = append(rets, ret.StringOfType())
		}
		return fmt.Sprintf("func(%s) (%s)", v.Params.Signature(), strings.Join(rets, ", "))
	}
	if v.Package != nil {
		if parts := strings.Split(v.Package.Path, "/"); len(parts) > 1 {
			return fmt.Sprintf("%s.%s", parts[len(parts)-1], v.Type)
		}
		return fmt.Sprintf("%s.%s", v.Package.Path, v.Type)
	}
	return v.Type
}

// IsContext return true if the value is a context.Context type, otherwise false.
func (v *Value) IsContext() bool {
	return v.Package != nil && v.Package.Path == "context" && v.Type == typeContext
}

// Package represents a Go package.
type Package struct {
	Path string
}
