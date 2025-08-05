package migrator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var GoalieAnalyzer = &analysis.Analyzer{
	Name: "goalieanalyzer",
	Doc:  "checks for missed errors in `defer`'d functions",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: run,
}

const goalieImportPath = "github.com/ras0q/goalie"

var errorType = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)

type FuncInfo struct {
	ReturnsError     bool
	NamedErrorVar    string
	IsAlreadyPatched bool
}

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.DeferStmt)(nil)}
	funcCache := make(map[*ast.FuncDecl]*FuncInfo)

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		deferStmt := n.(*ast.DeferStmt)
		deferredSig, ok := pass.TypesInfo.TypeOf(deferStmt.Call.Fun).(*types.Signature)
		if !ok || !returnsError(deferredSig.Results()) {
			return
		}

		enclosingFunc := findEnclosingFunc(pass, n.Pos())
		if enclosingFunc == nil || enclosingFunc.Body == nil {
			return
		}

		funcInfo, ok := funcCache[enclosingFunc]
		if !ok {
			funcInfo = analyzeFunc(pass, enclosingFunc)
			funcCache[enclosingFunc] = funcInfo
		}

		if !funcInfo.ReturnsError {
			pass.Report(analysis.Diagnostic{
				Pos: deferStmt.Pos(),
				Message: fmt.Sprintf(
					"missed error in defer statement, but cannot autofix because enclosing function %s does not return an error: %s",
					render(pass, deferStmt.Call),
					enclosingFunc.Name.Name,
				),
			})
			return
		}

		pass.Report(analysis.Diagnostic{
			Pos:     n.Pos(),
			Message: fmt.Sprintf("missed error in defer statement: %s", render(pass, deferStmt.Call)),
			SuggestedFixes: []analysis.SuggestedFix{
				createFix(pass, deferStmt, enclosingFunc, deferredSig, funcInfo),
			},
		})
	})

	return nil, nil
}

// analyzeFunc analyzes a function declaration and returns all relevant info.
// This is the single source of truth for function analysis.
func analyzeFunc(pass *analysis.Pass, funcDecl *ast.FuncDecl) *FuncInfo {
	info := &FuncInfo{}

	// --- Part 1: Analyze signature for error return ---
	results := funcDecl.Type.Results
	if results != nil {
		for _, field := range results.List {
			if field.Type != nil && types.Implements(pass.TypesInfo.TypeOf(field.Type), errorType) {
				info.ReturnsError = true
				if len(field.Names) > 0 {
					info.NamedErrorVar = field.Names[0].Name
				}
				break // Found the first error return
			}
		}
	}

	// --- Part 2: Check for existing goalie patch (idempotency) ---
	if !info.ReturnsError || info.NamedErrorVar == "" {
		// If there's no named error return, it can't be patched.
		info.IsAlreadyPatched = false
		return info
	}
	if funcDecl.Body == nil || len(funcDecl.Body.List) < 2 {
		info.IsAlreadyPatched = false
		return info
	}

	// Check for: g := goalie.New()
	assignStmt, ok := funcDecl.Body.List[0].(*ast.AssignStmt)
	if !ok || assignStmt.Tok != token.DEFINE || len(assignStmt.Lhs) != 1 || len(assignStmt.Rhs) != 1 {
		info.IsAlreadyPatched = false
		return info
	}
	if lhs, ok := assignStmt.Lhs[0].(*ast.Ident); !ok || lhs.Name != "g" {
		info.IsAlreadyPatched = false
		return info
	}
	callExpr, ok := assignStmt.Rhs[0].(*ast.CallExpr)
	if !ok {
		info.IsAlreadyPatched = false
		return info
	}
	selector, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		info.IsAlreadyPatched = false
		return info
	}
	if x, ok := selector.X.(*ast.Ident); !ok || x.Name != "goalie" || selector.Sel.Name != "New" {
		info.IsAlreadyPatched = false
		return info
	}

	// Check for: defer g.Collect(&err)
	deferStmt, ok := funcDecl.Body.List[1].(*ast.DeferStmt)
	if !ok {
		info.IsAlreadyPatched = false
		return info
	}
	selector, ok = deferStmt.Call.Fun.(*ast.SelectorExpr)
	if !ok {
		info.IsAlreadyPatched = false
		return info
	}
	if x, ok := selector.X.(*ast.Ident); !ok || x.Name != "g" || selector.Sel.Name != "Collect" {
		info.IsAlreadyPatched = false
		return info
	}
	if len(deferStmt.Call.Args) != 1 {
		info.IsAlreadyPatched = false
		return info
	}
	unaryExpr, ok := deferStmt.Call.Args[0].(*ast.UnaryExpr)
	if !ok || unaryExpr.Op != token.AND {
		info.IsAlreadyPatched = false
		return info
	}
	if ident, ok := unaryExpr.X.(*ast.Ident); !ok || ident.Name != info.NamedErrorVar {
		info.IsAlreadyPatched = false
		return info
	}

	// All checks passed.
	info.IsAlreadyPatched = true
	return info
}

func returnsError(results *types.Tuple) bool {
	if results == nil || results.Len() == 0 {
		return false
	}
	for i := 0; i < results.Len(); i++ {
		if types.Implements(results.At(i).Type(), errorType) {
			return true
		}
	}
	return false
}

func findEnclosingFunc(pass *analysis.Pass, pos token.Pos) *ast.FuncDecl {
	for _, file := range pass.Files {
		if file.Pos() <= pos && pos < file.End() {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					if fn.Pos() <= pos && pos < fn.End() {
						return fn
					}
				}
			}
		}
	}
	return nil
}

func createFix(
	pass *analysis.Pass,
	deferStmt *ast.DeferStmt,
	enclosingFunc *ast.FuncDecl,
	sig *types.Signature,
	funcInfo *FuncInfo,
) analysis.SuggestedFix {
	textEdits := []analysis.TextEdit{
		buildDeferEdit(pass, deferStmt, sig),
	}

	if !funcInfo.IsAlreadyPatched {
		textEdits = append(textEdits, buildFuncSetupEdits(pass, enclosingFunc)...)

		if importEdit := buildImportEdit(pass, enclosingFunc.Pos()); importEdit != nil {
			textEdits = append(textEdits, *importEdit)
		}
	}

	return analysis.SuggestedFix{
		Message:   "Handle defer with Goalie",
		TextEdits: textEdits,
	}
}

func buildFuncSetupEdits(pass *analysis.Pass, funcDecl *ast.FuncDecl) []analysis.TextEdit {
	results := funcDecl.Type.Results
	var newParams []string
	var errVarName = "err"

	if results == nil || results.List == nil {
		return []analysis.TextEdit{}
	}

	for _, field := range results.List {
		typeName := render(pass, field.Type)
		isErrorType := types.Implements(pass.TypesInfo.TypeOf(field.Type), errorType)
		if isErrorType {
			if len(field.Names) > 0 {
				errVarName = field.Names[0].Name
				newParams = append(newParams, fmt.Sprintf("%s %s", errVarName, typeName))
			} else {
				newParams = append(newParams, fmt.Sprintf("%s %s", errVarName, typeName))
			}
		} else {
			if len(field.Names) > 0 {
				newParams = append(newParams, fmt.Sprintf("%s %s", field.Names[0].Name, typeName))
			} else {
				newParams = append(newParams, fmt.Sprintf("_ %s", typeName))
			}
		}
	}

	newSignature := fmt.Sprintf("(%s)", strings.Join(newParams, ", "))
	sigEditPos, sigEditEnd := results.Pos(), results.End()
	sigEdit := analysis.TextEdit{
		Pos:     sigEditPos,
		End:     sigEditEnd,
		NewText: []byte(newSignature),
	}

	bodyInsertionPos := funcDecl.Body.Lbrace + 1
	bodySetupText := fmt.Sprintf(
		"\n\tg := goalie.New()\n"+
			"\tdefer g.Collect(&%s)\n\n",
		errVarName,
	)
	bodyEdit := analysis.TextEdit{
		Pos:     bodyInsertionPos,
		End:     bodyInsertionPos,
		NewText: []byte(bodySetupText),
	}

	return []analysis.TextEdit{
		sigEdit,
		bodyEdit,
	}
}

func buildDeferEdit(pass *analysis.Pass, deferStmt *ast.DeferStmt, sig *types.Signature) analysis.TextEdit {
	var newDeferText string

	// check if the function is `func () error`
	isFuncWithErrorOnly := sig.Params().Len() == 0 &&
		sig.Results().Len() == 1 &&
		types.Implements(sig.Results().At(0).Type(), errorType)
	if isFuncWithErrorOnly {
		funcExprStr := render(pass, deferStmt.Call.Fun) // f.Close
		newDeferText = fmt.Sprintf("defer g.Guard(%s)", funcExprStr)
	} else {
		callExprStr := render(pass, deferStmt.Call) // f.Close(params)
		newDeferText = fmt.Sprintf(
			"defer g.Guard(func () error {\n"+
				"\treturn %s\n"+
				"})",
			callExprStr,
		)
	}

	return analysis.TextEdit{
		Pos:     deferStmt.Pos(),
		End:     deferStmt.End(),
		NewText: []byte(newDeferText),
	}
}

func buildImportEdit(pass *analysis.Pass, pos token.Pos) *analysis.TextEdit {
	var currentFile *ast.File
	for _, file := range pass.Files {
		if file.Pos() <= pos && pos <= file.End() {
			currentFile = file
			break
		}
	}
	if currentFile == nil {
		return nil
	}
	for _, spec := range currentFile.Imports {
		if spec.Path.Value == goalieImportPath {
			return nil
		}
	}

	insertPos := currentFile.Name.End()
	newTextFormat := "import %q"
	if len(currentFile.Imports) > 0 {
		insertPos = currentFile.Imports[len(currentFile.Imports)-1].End()
		newTextFormat = "\n\t%q"
	}

	return &analysis.TextEdit{
		Pos:     insertPos,
		End:     insertPos,
		NewText: fmt.Appendf([]byte{}, newTextFormat, goalieImportPath),
	}
}

func render(pass *analysis.Pass, node ast.Node) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, pass.Fset, node); err != nil {
		return ""
	}
	return buf.String()
}
