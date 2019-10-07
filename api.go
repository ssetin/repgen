package main

import (
	"encoding/json"
	"go/ast"
	"log"
	"strings"
	"unicode"
)

// Describes all domain entities
type entityApi struct {
	Package string
	Structs map[string]*entityStruct
}

// Describes condition - Is entity active
type activeMethod struct {
	SQLCondition string `json:"active"`
	MethodName   string
}

// Describes domain entity
type entityStruct struct {
	RepositoryName string
	StructName     string
	VarName        string

	Table         string `json:"table"`
	SchemaVersion string `json:"schemaVersion"`
	SchemaName    string `json:"schemaName"`
	PrimaryKey    structField
	Fields        []structField
	SearchGroups  map[string][]*structField

	Active activeMethod
}

type structField struct {
	Name       string
	VarName    string
	Type       string
	IsPointer  bool
	IsExported bool
	Tag        tagInfo
}

type tagInfo struct {
	DbFieldName  string
	IsPrimary    bool
	IsSearchable bool
}

func newApi() *entityApi {
	api := &entityApi{}
	api.Structs = make(map[string]*entityStruct)
	return api
}

func getFieldType(field *ast.Field) (string, bool) {
	var (
		isPointer bool
		fieldType string
	)

	if ident, ok := field.Type.(*ast.Ident); ok {
		fieldType = ident.Name
	} else if star, ok := field.Type.(*ast.StarExpr); ok {
		isPointer = true

		if ident, ok := star.X.(*ast.Ident); ok {
			fieldType = ident.Name
		}

		if selectorExpr, ok := star.X.(*ast.SelectorExpr); ok {
			if ident, ok := selectorExpr.X.(*ast.Ident); ok {
				fieldType = ident.Name + "." + selectorExpr.Sel.Name
			}
		}
	}

	return fieldType, isPointer
}

func unTitle(s string) string {
	var newRune rune
	runesStr := []rune(s)

	result := strings.Builder{}
	for idx := 0; idx < len(runesStr); idx++ {
		if idx > 0 && idx+1 < len(s) && unicode.IsLower(runesStr[idx+1]) {
			newRune = runesStr[idx]
		} else {
			newRune = unicode.ToLower(runesStr[idx])
		}
		result.WriteRune(newRune)
	}

	return result.String()
}

func (e *entityStruct) addMethod(methodName, jsonStr string) {
	method := activeMethod{
		MethodName: methodName,
	}
	err := json.Unmarshal([]byte(jsonStr[len(generateTag):]), &method)
	if err != nil {
		log.Fatalf("Error parsing comment %s. %s", jsonStr[len(generateTag):], err.Error())
	}
	e.Active = method
}

// `db:"field=phone,searchable"`
func (e *entityStruct) parseTag(f *structField, tag string) tagInfo {
	var (
		result tagInfo
		val    string
	)
	inner := tag[5 : 5+strings.Index(tag[5:], "\"")]
	split := strings.Split(inner, ",")

	for _, pair := range split {
		paramValue := strings.Split(pair, "=")
		if len(paramValue) == 2 {
			val = paramValue[1]
		} else {
			val = ""
		}

		switch paramValue[0] {
		case "primary":
			result.IsPrimary = true
		case "searchable":
			result.IsSearchable = true
		case "field":
			result.DbFieldName = "\"" + val + "\""
		case "searchGroup":
			e.SearchGroups[val] = append(e.SearchGroups[val], f)
		}
	}

	if result.IsPrimary && !f.IsPointer {
		log.Fatalf("Primary key %s should be pointer type", f.Name)
	}

	return result
}
