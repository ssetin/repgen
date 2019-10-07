package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

const (
	generateTag = "// gen:ddd"
)

func scanFile(node *ast.File, api *entityApi) {
	scanDeclarations(node, api)
	scanMethods(node, api)
}

func scanMethods(node *ast.File, api *entityApi) {
	for _, decl := range node.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if decl.Recv == nil {
				continue
			}
			needCodegen := false
			jsonStr := ""

			if needCodegen, jsonStr = extractComments(decl.Doc, needCodegen, jsonStr); !needCodegen {
				continue
			}

			recvType := decl.Recv.List[0].Type
			recvName := recvType.(*ast.Ident).Name

			if e, found := api.Structs[recvName]; found {
				e.addMethod(decl.Name.Name, jsonStr)
			}
		}
	}
}

func extractComments(doc *ast.CommentGroup, needCodegen bool, jsonStr string) (bool, string) {
	if doc == nil {
		return false, ""
	}
	for _, comment := range doc.List {
		needCodegen = needCodegen || strings.HasPrefix(comment.Text, generateTag)
		jsonStr += comment.Text
	}
	return needCodegen, jsonStr
}

func extractStructType(spec ast.Spec) (*ast.StructType, *ast.TypeSpec) {
	currType, ok := spec.(*ast.TypeSpec)
	if !ok {
		return nil, nil
	}
	currStruct, ok := currType.Type.(*ast.StructType)
	if !ok || len(currStruct.Fields.List) == 0 {
		return nil, nil
	}
	return currStruct, currType
}

func scanDeclarations(node *ast.File, api *entityApi) {
	for _, decl := range node.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				currStruct, currType := extractStructType(spec)
				if currStruct == nil {
					continue
				}
				needCodegen := false
				jsonStr := ""

				if needCodegen, jsonStr = extractComments(decl.Doc, needCodegen, jsonStr); !needCodegen {
					continue
				}

				entityStruct := newStruct(jsonStr, currType, currStruct)
				api.Structs[entityStruct.StructName] = entityStruct
				api.Package = node.Name.Name
			}
		}
	}
}

func newStruct(jsonStr string, currType *ast.TypeSpec, currStruct *ast.StructType) *entityStruct {
	entityStruct := &entityStruct{
		StructName:     currType.Name.Name,
		RepositoryName: currType.Name.Name + "Repository",
		VarName:        unTitle(currType.Name.Name),
		Fields:         make([]structField, 0, 3),
		SearchGroups:   make(map[string][]*structField),
	}
	err := json.Unmarshal([]byte(jsonStr[len(generateTag):]), &entityStruct)
	if err != nil {
		log.Fatalf("Error parsing comment %s. %s", jsonStr[len(generateTag):], err.Error())
	}
	entityStruct.Table = "\"" + entityStruct.Table + "\""

	for _, field := range currStruct.Fields.List {
		fieldName := field.Names[0].Name
		fieldType, isPointer := getFieldType(field)
		structField := structField{
			Name:       fieldName,
			VarName:    unTitle(fieldName),
			Type:       fieldType,
			IsPointer:  isPointer,
			IsExported: field.Names[0].IsExported(),
		}
		structField.Tag = entityStruct.parseTag(&structField, field.Tag.Value)

		if structField.Tag.IsPrimary {
			entityStruct.PrimaryKey = structField
		}
		entityStruct.Fields = append(entityStruct.Fields, structField)
	}
	return entityStruct
}

func createFiles(interfaceFile, implementFile, mockFile string, node *ast.File, api *entityApi) {
	out, _ := os.Create(interfaceFile)
	defer out.Close()

	outImpl, _ := os.Create(implementFile)
	defer outImpl.Close()

	outMock, _ := os.Create(mockFile)
	defer outMock.Close()

	scanFile(node, api)

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"cat": func(a, b string) string {
			return a + b
		},
	}

	t := template.Must(template.New("implementation.gotext").Funcs(funcMap).ParseFiles("templates/implementation.gotext"))
	err := t.Execute(outImpl, api)
	if err != nil {
		log.Fatal(err)
	}

	t = template.Must(template.New("interface.gotext").Funcs(funcMap).ParseFiles("templates/interface.gotext"))
	err = t.Execute(out, api)
	if err != nil {
		log.Fatal(err)
	}

	t = template.Must(template.New("mock.gotext").Funcs(funcMap).ParseFiles("templates/mock.gotext"))
	err = t.Execute(outMock, api)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	api := newApi()
	sourceFile := os.Args[1]
	interfaceFile := os.Args[2]
	implementFile := os.Args[3]
	mockFile := os.Args[4]

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, sourceFile, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	createFiles(interfaceFile, implementFile, mockFile, node, api)
}
