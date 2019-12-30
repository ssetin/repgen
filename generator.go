package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	generateTag = "// gen:ddd"
)

func (api *entities) scanFile(pkgFile packageFile) {
	api.scanDeclarations(pkgFile)
	api.scanMethods(pkgFile)

	err := api.ShapeNestedObjects()
	if err != nil {
		log.Fatal(err)
	}
}

func (api *entities) scanMethods(packageFile packageFile) {
	for _, decl := range packageFile.Decls {
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

			if e, found := api.Structs[packageFile.fileName][recvName]; found {
				e.addMethod(decl.Name.Name, jsonStr)
			}
			if e, found := api.NestedStructs[recvName]; found {
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

func extractTypeSpec(spec ast.Spec) (*ast.StructType, *ast.TypeSpec) {
	currType, ok := spec.(*ast.TypeSpec)
	if !ok {
		return nil, nil
	}
	currStruct, ok := currType.Type.(*ast.StructType)
	if !ok || len(currStruct.Fields.List) == 0 {
		return nil, currType
	}
	return currStruct, currType
}

func (api *entities) preScanAllDeclarations(packageFile packageFile) {
	for _, decl := range packageFile.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				_, currType := extractTypeSpec(spec)
				if currType != nil && currType.Name != nil {
					api.Types[currType.Name.Name] = struct{}{}
				}
			}
		}
	}
}

func (api *entities) scanDeclarations(packageFile packageFile) {
	api.Structs[packageFile.fileName] = make(map[string]*entityStruct)
	for _, decl := range packageFile.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				currStruct, currType := extractTypeSpec(spec)
				if currStruct == nil {
					continue
				}
				needCodegen := false
				commentStr := ""

				if needCodegen, commentStr = extractComments(decl.Doc, needCodegen, commentStr); !needCodegen {
					continue
				}

				api.Package = packageFile.Name.Name
				api.addStruct(commentStr, packageFile.fileName, currType, currStruct)
			}
		}
	}
}

func (api *entities) createFiles() {
	packFiles := api.buildAST(api.sourcePath)
	for _, node := range packFiles {
		api.preScanAllDeclarations(node)
	}

	for _, node := range packFiles {
		api.scanFile(node)
	}

	for fileName, _ := range api.Structs {
		fileApi := api.Copy(fileName)
		base := strings.Split(fileName, ".")
		if len(base) > 0 && fileApi != nil {
			newFileName := strings.ToLower(base[0]) + ".gen.go"
			interfaceFile := filepath.Join(api.sourcePath, newFileName)
			implementFile := filepath.Join(api.implementPath, newFileName)
			mockFile := filepath.Join(api.mockPath, newFileName)
			fileApi.createFile(api.homeDir, interfaceFile, implementFile, mockFile)
		}
	}
}

func (api *fileEntities) createFile(homeDir string, interfaceFile string, implementFile string, mockFile string) {
	out, _ := os.Create(interfaceFile)
	defer out.Close()
	outImpl, _ := os.Create(implementFile)
	defer outImpl.Close()
	outMock, _ := os.Create(mockFile)
	defer outMock.Close()

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"cat": func(a, b string) string {
			return a + b
		},
	}

	t := template.Must(template.New("implementation.gotext").Funcs(funcMap).ParseFiles(homeDir + "/templates/implementation.gotext"))
	err := t.Execute(outImpl, api)
	if err != nil {
		log.Fatal(err)
	}

	t = template.Must(template.New("interface.gotext").Funcs(funcMap).ParseFiles(homeDir + "/templates/interface.gotext"))
	err = t.Execute(out, api)
	if err != nil {
		log.Fatal(err)
	}

	t = template.Must(template.New("mock.gotext").Funcs(funcMap).ParseFiles(homeDir + "/templates/mock.gotext"))
	err = t.Execute(outMock, api)
	if err != nil {
		log.Fatal(err)
	}
}

func (api *entities) buildAST(sourcePath string) []packageFile {
	sourceDir, err := os.Open(sourcePath)
	if err != nil {
		log.Fatal(err)
	}
	defer sourceDir.Close()

	sourceFiles, err := sourceDir.Readdir(0)
	nodes := make([]packageFile, 0, len(sourceFiles))

	for _, sourceFile := range sourceFiles {
		if sourceFile.IsDir() || strings.HasSuffix(sourceFile.Name(), ".gen.go") || !strings.HasSuffix(sourceFile.Name(), ".go") {
			continue
		}

		fileSet := token.NewFileSet()
		node, err := parser.ParseFile(fileSet, filepath.Join(sourcePath, sourceFile.Name()), nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		nodes = append(nodes, packageFile{fileName: sourceFile.Name(), File: node})
	}
	return nodes
}
