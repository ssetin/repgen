package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"log"
	"strings"
	"unicode"
)

// packageFile: syntax tree + file name
type packageFile struct {
	*ast.File
	fileName string
}

type structsMap map[string]*entityStruct

// Describes entities of the whole domain package
type entities struct {
	homeDir, sourcePath, implementPath, mockPath string

	Package string
	// aggregated entities
	Structs map[string]structsMap
	// nested entities
	NestedStructs map[string]*entityStruct

	// all package types
	Types map[string]struct{}
}

// Describes entities associated with one file
type fileEntities struct {
	Package       string
	Structs       map[string]*entityStruct
	NestedStructs map[string]*entityStruct
	Types         map[string]struct{}
}

// Describes condition - Is entity active
type activeMethod struct {
	SQLCondition string `json:"active"`
	MethodName   string
}

type nestedObject struct {
	ForeignKey structField
	Object     *entityStruct
}

// Describes domain entity
type entityStruct struct {
	RepositoryName string
	StructName     string
	// name prefixed with package
	FullStructName string
	VarName        string
	Alias          string

	Table  string `json:"table"`
	Nested bool   `json:"nested"`

	PrimaryKey structField
	Fields     []structField
	// nested structs, which foreign keys points to
	NestedObjects []*nestedObject
	// groups of fields for generating find methods
	SearchGroups map[string][]*structField

	Active activeMethod
}

type structField struct {
	Name        string
	VarName     string
	ForeignName string

	// type identity
	Type string
	// type identity prefixed with package
	FullType   string
	IsPointer  bool
	IsExported bool
	IsArray    bool
	Tag        tagInfo
}

type tagInfo struct {
	DbFieldName  string
	IsPrimary    bool
	IsSearchable bool
	IsForeign    bool
}

func newEntitiesCollection(homeDir, sourcePath, implementPath, mockPath string) *entities {
	api := &entities{
		homeDir:       homeDir,
		sourcePath:    sourcePath,
		implementPath: implementPath,
		mockPath:      mockPath,
	}
	api.Structs = make(map[string]structsMap)
	api.NestedStructs = make(map[string]*entityStruct)
	api.Types = make(map[string]struct{})
	return api
}

func getFieldType(field *ast.Field) (string, bool, bool) {
	var (
		isPointer bool
		isArray   bool
		fieldType string
	)

	switch val := field.Type.(type) {
	case *ast.ArrayType:
		isArray = true
		switch ident := val.Elt.(type) {
		case *ast.Ident:
			fieldType = ident.Obj.Name
		case *ast.StarExpr:
			isPointer, fieldType = getStartType(ident)
		}
	case *ast.Ident:
		fieldType = val.Name
	case *ast.SelectorExpr:
		isPointer, fieldType = getSelectorType(val)
	case *ast.StarExpr:
		isPointer, fieldType = getStartType(val)
	}

	return fieldType, isPointer, isArray
}

func getSelectorType(val *ast.SelectorExpr) (bool, string) {
	var (
		packageName string
		fieldType   string
	)
	switch ident := val.X.(type) {
	case *ast.Ident:
		packageName = ident.Name
	}
	fieldType = fmt.Sprintf("%s.%s", packageName, val.Sel.Name)
	return false, fieldType
}

func getStartType(val *ast.StarExpr) (bool, string) {
	var fieldType string
	if ident, ok := val.X.(*ast.Ident); ok {
		fieldType = ident.Name
	}
	if selectorExpr, ok := val.X.(*ast.SelectorExpr); ok {
		if ident, ok := selectorExpr.X.(*ast.Ident); ok {
			fieldType = ident.Name + "." + selectorExpr.Sel.Name
		}
	}
	return true, fieldType
}

// ===================================== entities ==============================================

func (api *entities) addStruct(commentStr, fileName string, currType *ast.TypeSpec, currStruct *ast.StructType) {
	entityStruct := &entityStruct{
		StructName:     currType.Name.Name,
		FullStructName: api.getFullTypeName(currType.Name.Name),
		RepositoryName: currType.Name.Name + "Repository",
		VarName:        api.notExportable(currType.Name.Name),
		Fields:         make([]structField, 0, 3),
		SearchGroups:   make(map[string][]*structField),
		NestedObjects:  make([]*nestedObject, 0),
	}
	err := json.Unmarshal([]byte(commentStr[len(generateTag):]), &entityStruct)
	if err != nil {
		log.Fatalf("Error parsing comment %s. %s", commentStr[len(generateTag):], err.Error())
	}

	entityStruct.Table = "\"" + entityStruct.Table + "\""

	for _, field := range currStruct.Fields.List {
		if field.Tag == nil {
			continue
		}

		fieldName := field.Names[0].Name
		fieldType, isPointer, isArray := getFieldType(field)

		structField := structField{
			Name:       fieldName,
			VarName:    api.notExportable(fieldName),
			FullType:   api.getFullTypeName(fieldType),
			Type:       fieldType,
			IsPointer:  isPointer,
			IsArray:    isArray,
			IsExported: field.Names[0].IsExported(),
		}
		structField.Tag = entityStruct.parseTag(&structField, field.Tag.Value)

		if structField.Tag.IsPrimary {
			entityStruct.PrimaryKey = structField
		}

		if structField.Tag.IsForeign {
			nested := &nestedObject{
				ForeignKey: structField,
			}
			entityStruct.NestedObjects = append(entityStruct.NestedObjects, nested)
		} else {
			entityStruct.Fields = append(entityStruct.Fields, structField)
		}
	}

	if entityStruct.Nested {
		api.NestedStructs[entityStruct.StructName] = entityStruct
	} else {
		api.Structs[fileName][entityStruct.StructName] = entityStruct
	}
}

// ShapeNestedObjects injects full information about nested objects according to foreign keys
func (api *entities) ShapeNestedObjects() error {
	for _, file := range api.Structs {
		for _, item := range file {
			for _, nested := range item.NestedObjects {
				if s, ok := api.NestedStructs[nested.ForeignKey.Type]; ok {
					nested.Object = s
					nested.ForeignKey.ForeignName = s.getForeignKeyName(nested.ForeignKey.Tag.DbFieldName)
				} else {
					return fmt.Errorf("nested struct [%s] not found", nested.ForeignKey.Type)
				}
			}
		}
	}
	return nil
}

func (api entities) getFullTypeName(typeName string) string {
	if _, ok := api.Types[typeName]; ok {
		return api.Package + "." + typeName
	}
	return typeName
}

func (api entities) notExportable(s string) string {
	var newRune rune
	runesStr := []rune(s)

	if strings.ToUpper(s) == strings.ToUpper(api.Package) {
		return "my" + strings.Title(api.Package)
	}

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

// Copy returns copy of entities associated with specified file
func (api *entities) Copy(fileName string) *fileEntities {
	if _, ok := api.Structs[fileName]; !ok {
		return nil
	}
	if len(api.Structs[fileName]) == 0 {
		return nil
	}

	return &fileEntities{
		Package:       api.Package,
		Structs:       api.Structs[fileName],
		NestedStructs: api.NestedStructs,
		Types:         api.Types,
	}
}

// ===================================== entityStruct ==============================================

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

func (e *entityStruct) getForeignKeyName(dbForeignFieldName string) string {
	for _, item := range e.Fields {
		if item.Tag.DbFieldName == dbForeignFieldName {
			return item.Name
		}
	}
	return ""
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
		case "foreign":
			result.IsForeign = true
			result.DbFieldName = "\"" + val + "\""
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
