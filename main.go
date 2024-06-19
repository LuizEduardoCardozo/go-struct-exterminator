package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

type StructUsageChecker struct {
	structs map[string]bool
}

func NewStructUsageChecker() *StructUsageChecker {
	return &StructUsageChecker{
		structs: make(map[string]bool),
	}
}

func (s *StructUsageChecker) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.TypeSpec:
		if _, ok := n.Type.(*ast.StructType); ok {
			if exists := s.structs[n.Name.Name]; !exists {
				fmt.Println("New struct found: ", n.Name.Name)
				s.structs[n.Name.Name] = false
			}
		}
		break
	case *ast.SelectorExpr:
		if _, exists := s.structs[n.Sel.Name]; exists {
			s.structs[n.Sel.Name] = true
		}
		break
	}
	return s
}

func (s *StructUsageChecker) CheckFileStructs(filename string) error {
	fs := token.NewFileSet()

	node, err := parser.ParseFile(fs, filename, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	ast.Walk(s, node)
	return nil
}

func (s *StructUsageChecker) CheckFileUsages(filename string) error {
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, filename, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	ast.Walk(s, node)
	return nil
}

func (s *StructUsageChecker) CheckDir(dirname string) error {
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			if err := s.CheckFileStructs(path); err != nil {
				fmt.Printf("Error while analyzing the file %s: %v\n", path, err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			if err := s.CheckFileUsages(path); err != nil {
				fmt.Printf("Error while analyzing the file %s: %v\n", path, err)
			}
		}
		return nil
	})
	return err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <path>", os.Args[0])
		return
	}

	path := os.Args[1]
	checker := NewStructUsageChecker()

	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error while reading the specified path: %v\n", err)
		return
	}

	if info.IsDir() {
		err = checker.CheckDir(path)
	} else {
		err = checker.CheckFileStructs(path)
		if err == nil {
			err = checker.CheckFileUsages(path)
		}
	}

	if err != nil {
		fmt.Printf("Error while analyzing the specified path: %v\n", err)
		return
	}

	fmt.Println("Non used structs found:")
	for structName, used := range checker.structs {
		if !used {
			fmt.Println(structName)
		}
	}
}
