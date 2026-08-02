package generator

import (
	"bytes"
	"fmt"
	"github.com/EmptyZeroRain/json2struct/option"
	"github.com/EmptyZeroRain/json2struct/schema"
	"go/ast"
	"go/format"
	"go/token"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Generate(s *schema.Field, opts option.Options) ([]byte, error) {
	if s == nil {
		return nil, fmt.Errorf("schema is nil")
	}
	opts.Defaults()
	if !isIdentifier(opts.Package) || !isIdentifier(opts.Name) {
		return nil, fmt.Errorf("invalid package or struct name")
	}
	file := &ast.File{Name: ast.NewIdent(opts.Package)}
	file.Decls = []ast.Decl{&ast.GenDecl{Tok: token.TYPE, Specs: []ast.Spec{structSpec(opts.Name, s, opts)}}}
	var out bytes.Buffer
	if err := format.Node(&out, token.NewFileSet(), file); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func structSpec(name string, f *schema.Field, o option.Options) ast.Spec {
	fields := []*ast.Field{}
	used := map[string]int{}
	for _, k := range sortedKeys(f.Children) {
		child := f.Children[k]
		fieldName := uniqueName(exportedName(k), used)
		typ := goType(child, fieldName, o)
		tag := ""
		if o.JsonTag {
			tag = `json:"` + k
			if o.Omitempty && !child.Required {
				tag += `,omitempty`
			}
			tag += `"`
		}
		if child.Tags != nil {
			if custom, ok := child.Tags["json"]; ok {
				tag = custom
			}
		}
		fields = append(fields, &ast.Field{Names: []*ast.Ident{ast.NewIdent(fieldName)}, Type: typ, Tag: tagLit(tag)})
	}
	return &ast.TypeSpec{Name: ast.NewIdent(name), Type: &ast.StructType{Fields: &ast.FieldList{List: fields}}}
}
func isIdentifier(s string) bool {
	if s == "" || !((s[0] >= 'A' && s[0] <= 'Z') || (s[0] >= 'a' && s[0] <= 'z') || s[0] == '_') {
		return false
	}
	for _, r := range s[1:] {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}
func goType(f *schema.Field, nested string, o option.Options) ast.Expr {
	var e ast.Expr
	switch f.Type {
	case schema.TypeString:
		e = ast.NewIdent("string")
	case schema.TypeInteger:
		if o.NumberType == "float" {
			e = ast.NewIdent("float64")
		} else {
			e = ast.NewIdent("int")
		}
	case schema.TypeNumber:
		e = ast.NewIdent("float64")
	case schema.TypeBoolean:
		e = ast.NewIdent("bool")
	case schema.TypeObject:
		e = &ast.StructType{Fields: &ast.FieldList{List: objectFields(f, o)}}
	case schema.TypeArray:
		if f.Element == nil {
			e = ast.NewIdent("interface{}")
		} else {
			e = goType(f.Element, nested, o)
		}
	case schema.TypeNull:
		// A JSON null has no concrete Go type. String is the safe default and
		// Nullable makes the generated field *string.
		e = ast.NewIdent("string")
	default:
		e = ast.NewIdent("string")
	}
	if f.Array || f.Type == schema.TypeArray {
		return &ast.ArrayType{Elt: e}
	}
	if f.Nullable && f.Type != schema.TypeAny {
		return &ast.StarExpr{X: e}
	}
	return e
}
func objectFields(f *schema.Field, o option.Options) []*ast.Field {
	r := []*ast.Field{}
	used := map[string]int{}
	for _, k := range sortedKeys(f.Children) {
		c := f.Children[k]
		fieldName := uniqueName(exportedName(k), used)
		tag := ""
		if o.JsonTag {
			tag = `json:"` + k
			if o.Omitempty && !c.Required {
				tag += `,omitempty`
			}
			tag += `"`
		}
		if c.Tags != nil {
			if custom, ok := c.Tags["json"]; ok {
				tag = custom
			}
		}
		r = append(r, &ast.Field{Names: []*ast.Ident{ast.NewIdent(fieldName)}, Type: goType(c, fieldName, o), Tag: tagLit(tag)})
	}
	return r
}
func tagLit(s string) *ast.BasicLit {
	if s == "" {
		return nil
	}
	if strings.Contains(s, "`") {
		return &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(s)}
	}
	return &ast.BasicLit{Kind: token.STRING, Value: "`" + s + "`"}
}
func sortedKeys(m map[string]*schema.Field) []string {
	r := make([]string, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	sortStrings(r)
	return r
}
func sortStrings(a []string) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}

func exportedName(s string) string {
	// JSON keys are unrestricted strings. Keep letters and digits, use
	// separators as word boundaries, and always produce a legal exported Go
	// identifier. The JSON tag remains the original key.
	var b strings.Builder
	upperNext := true
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			upperNext = true
			continue
		}
		if upperNext {
			r = unicode.ToUpper(r)
			upperNext = false
		}
		b.WriteRune(r)
	}
	name := b.String()
	if name == "" {
		name = "Field"
	}
	first, _ := utf8.DecodeRuneInString(name)
	if unicode.IsDigit(first) {
		name = "Field" + name
	}
	if isGoKeyword(name) {
		name += "Field"
	}
	// Preserve common initialisms even when the key is written in lowercase.
	for short, acronym := range map[string]string{"Id": "ID", "Ip": "IP", "Url": "URL", "Http": "HTTP", "Https": "HTTPS", "Api": "API", "Json": "JSON", "Uuid": "UUID"} {
		if strings.HasPrefix(name, short) {
			name = acronym + name[len(short):]
			break
		}
	}
	return name
}

func uniqueName(base string, used map[string]int) string {
	if used[base] == 0 {
		used[base] = 1
		return base
	}
	used[base]++
	return base + strconv.Itoa(used[base])
}

func isGoKeyword(s string) bool {
	switch s {
	case "break", "default", "func", "interface", "select", "case", "defer", "go", "map", "struct", "chan", "else", "goto", "package", "switch", "const", "fallthrough", "if", "range", "type", "continue", "for", "import", "return", "var":
		return true
	}
	return false
}
