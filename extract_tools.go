package main

import (
	"encoding/csv"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ToolArg holds information about a tool argument
type ToolArg struct {
	Name        string
	Required    bool
	Description string
}

// ToolDef holds information about a tool
type ToolDef struct {
	VarName     string
	ToolName    string
	Description string
	Args        []ToolArg
}

func main() {
	var tools []ToolDef
	root := "internal/mcp"
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, "tools.go") {
			return nil
		}
		fileTools, err := parseToolsFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", path, err)
			return nil
		}
		tools = append(tools, fileTools...)
		return nil
	})

	writeCSV(tools, "tools_summary.csv")
	fmt.Println("Wrote tools_summary.csv")
}

func parseToolsFile(path string) ([]ToolDef, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var tools []ToolDef
	ast.Inspect(f, func(n ast.Node) bool {
		gen, ok := n.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			return true
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || len(vs.Names) == 0 || vs.Values == nil {
				continue
			}
			call, ok := vs.Values[0].(*ast.CallExpr)
			if !ok {
				continue
			}
			// Check for mcp.NewTool
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "mcp" && sel.Sel.Name == "NewTool" {
					tool := ToolDef{VarName: vs.Names[0].Name}
					if len(call.Args) > 0 {
						if nameLit, ok := call.Args[0].(*ast.BasicLit); ok {
							tool.ToolName = strings.Trim(nameLit.Value, "\"")
						}
					}
					// Parse options (description, args)
					for _, arg := range call.Args[1:] {
						if optCall, ok := arg.(*ast.CallExpr); ok {
							if sel, ok := optCall.Fun.(*ast.SelectorExpr); ok {
								if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "mcp" {
									switch sel.Sel.Name {
									case "WithDescription":
										if len(optCall.Args) > 0 {
											if descLit, ok := optCall.Args[0].(*ast.BasicLit); ok {
												tool.Description = strings.Trim(descLit.Value, "\"")
											}
										}
									case "WithString", "WithInt":
										argName := ""
										argDesc := ""
										argRequired := false
										if len(optCall.Args) > 0 {
											if nameLit, ok := optCall.Args[0].(*ast.BasicLit); ok {
												argName = strings.Trim(nameLit.Value, "\"")
											}
										}
										// Check for required and description in subcalls
										for _, sub := range optCall.Args[1:] {
											if subCall, ok := sub.(*ast.CallExpr); ok {
												if sel2, ok := subCall.Fun.(*ast.SelectorExpr); ok {
													if ident2, ok := sel2.X.(*ast.Ident); ok && ident2.Name == "mcp" {
														switch sel2.Sel.Name {
														case "Required":
															argRequired = true
														case "Description":
															if len(subCall.Args) > 0 {
																if descLit, ok := subCall.Args[0].(*ast.BasicLit); ok {
																	argDesc = strings.Trim(descLit.Value, "\"")
																}
															}
														}
													}
												}
											}
										}
										tool.Args = append(tool.Args, ToolArg{
											Name:        argName,
											Required:    argRequired,
											Description: argDesc,
										})
									}
								}
							}
						}
					}
					tools = append(tools, tool)
				}
			}
		}
		return true
	})
	return tools, nil
}

func writeCSV(tools []ToolDef, outPath string) {
	f, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{"Tool Name", "Tool Description", "Argument Name", "Required", "Argument Description"})
	for _, tool := range tools {
		if len(tool.Args) == 0 {
			w.Write([]string{tool.ToolName, tool.Description, "", "", ""})
			continue
		}
		for i, arg := range tool.Args {
			row := []string{"", "", arg.Name, map[bool]string{true: "Yes", false: "No"}[arg.Required], arg.Description}
			if i == 0 {
				row[0] = tool.ToolName
				row[1] = tool.Description
			}
			w.Write(row)
		}
	}
}
