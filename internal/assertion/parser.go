package assertion

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// Parser represents a source file parser.
type Parser struct {
	m sync.Mutex

	// Excluded call exprs should be excluded when finding assignments.
	excluded []*ast.CallExpr
}

// Info represents code analysis information of an assertion function.
type Info struct {
	Source string   // Source code of the caller.
	Args   []string // Selected arguments.

	// The last assignments related to Args.
	// The len(Assignments) is guaranteed to be the same as len(Args).
	Assignments [][]string

	// RelatedVars is the list of variables which are referenced in argument expression
	// or in the assignment statements containing arguments.
	// For instance, consider following code.
	//
	//     for i, c := range cases {
	//         a := 1
	//         Assert(t, a+1 == c.Value)
	//     }
	//
	// After parsing `Assert`, RelatedVars contains `a`, `c.Value` and `i`.
	// Note that, `i` is listed in related vars because of the value of `i` is assigned in
	// `i, c := range cases` in which `c` is also assigned.
	RelatedVars []string
}

// Func represents AST information of an assertion function.
type Func struct {
	FileSet *token.FileSet
	Func    *ast.FuncDecl
	Caller  *ast.CallExpr
	Args    []ast.Expr

	Filename string
	Line     int
}

// ParseArgs parses caller's source code, finds out the right call expression by name
// and returns the argument source AST.
//
// Skip is the stack frame calling an assert function. If skip is 0, the stack frame for
// ParseArgs is selected.
// In most cases, caller should set skip to 1 to skip ParseArgs itself.
func (p *Parser) ParseArgs(name string, skip int, argIndex []int) (f *Func, err error) {
	if len(argIndex) == 0 {
		err = fmt.Errorf("missing argIndex")
		return
	}

	filename, line, err := findCaller(skip + 1)

	if err != nil {
		return
	}

	dotIdx := strings.LastIndex(name, ".")

	if dotIdx >= 0 {
		name = name[dotIdx+1:]
	}

	fset, parsedAst, err := parseFile(filename)
	filename = path.Base(filename)

	if err != nil {
		return
	}

	var funcDecl *ast.FuncDecl
	var caller *ast.CallExpr
	argExprs := make([]ast.Expr, 0, len(argIndex))
	maxArgIdx := 0

	for _, idx := range argIndex {
		if idx > maxArgIdx {
			maxArgIdx = idx
		}
	}

	// Inspect AST and find target function at target line.
	done := false
	ast.Inspect(parsedAst, func(node ast.Node) bool {
		if node == nil || done {
			return false
		}

		if decl, ok := node.(*ast.FuncDecl); ok {
			funcDecl = decl
			return true
		}

		call, ok := node.(*ast.CallExpr)

		if !ok {
			return true
		}

		var fn string
		switch expr := call.Fun.(type) {
		case *ast.Ident:
			fn = expr.Name
		case *ast.SelectorExpr:
			fn = expr.Sel.Name
		}

		if fn != name {
			return true
		}

		pos := fset.Position(call.Pos())
		posEnd := fset.Position(call.End())

		if line < pos.Line || line > posEnd.Line {
			return true
		}

		caller = call

		for _, idx := range argIndex {
			if idx < 0 {
				idx += len(call.Args)
			}

			if idx < 0 || idx >= len(call.Args) {
				// Ignore invalid idx.
				argExprs = append(argExprs, nil)
				continue
			}

			arg := call.Args[idx]
			argExprs = append(argExprs, arg)
		}

		done = true
		return false
	})

	f = &Func{
		FileSet: fset,
		Func:    funcDecl,
		Caller:  caller,
		Args:    argExprs,

		Filename: filename,
		Line:     line,
	}
	return
}

// ParseInfo returns more context related information about this f.
// See document of Info for details.
func (p *Parser) ParseInfo(f *Func) (info *Info) {
	fset := f.FileSet
	args := make([]string, 0, len(f.Args))
	assignments := make([][]string, 0, len(f.Args))
	relatedVars := make(map[string]struct{})

	// If args contains any arg which is an ident, find out where it's assigned.
	for _, arg := range f.Args {
		assigns, related := findAssignments(fset, f.Func, f.Line, arg, p.excluded)
		args = append(args, formatNode(fset, arg))
		assignments = append(assignments, assigns)

		for v := range related {
			relatedVars[v] = struct{}{}
		}
	}

	vars := make([]string, 0, len(relatedVars))

	for v := range relatedVars {
		vars = append(vars, v)
	}

	sort.Strings(vars)
	info = &Info{
		Source:      formatNode(fset, f.Caller),
		Args:        args,
		Assignments: assignments,
		RelatedVars: vars,
	}
	return
}

func findCaller(skip int) (filename string, line int, err error) {
	const minimumSkip = 2 // Skip 2 frames running runtime functions.

	pc := make([]uintptr, 1)
	n := runtime.Callers(skip+minimumSkip, pc)

	if n == 0 {
		err = fmt.Errorf("fail to read call stack")
		return
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	frame, _ := frames.Next()
	filename = frame.File
	line = frame.Line

	if filename == "" || line == 0 {
		err = fmt.Errorf("fail to read source code information")
	}

	return
}

type fileAST struct {
	FileSet *token.FileSet
	File    *ast.File
}

var (
	fileCacheLock sync.Mutex
	fileCache     = map[string]*fileAST{}
)

func parseFile(filename string) (fset *token.FileSet, f *ast.File, err error) {
	fileCacheLock.Lock()
	fa, ok := fileCache[filename]
	fileCacheLock.Unlock()

	if ok {
		fset = fa.FileSet
		f = fa.File
		return
	}

	file, err := os.Open(filename)

	if err != nil {
		return
	}

	defer file.Close()
	fset = token.NewFileSet()
	f, err = parser.ParseFile(fset, filename, file, 0)

	fileCacheLock.Lock()
	fileCache[filename] = &fileAST{
		FileSet: fset,
		File:    f,
	}
	fileCacheLock.Unlock()
	return
}

func formatNode(fset *token.FileSet, node ast.Node) string {
	if node == nil {
		return ""
	}

	buf := &bytes.Buffer{}
	config := &printer.Config{
		Mode:     printer.UseSpaces,
		Tabwidth: 4,
	}
	config.Fprint(buf, fset, node)
	return buf.String()
}

func findAssignments(fset *token.FileSet, decl *ast.FuncDecl, line int, arg ast.Expr, excluded []*ast.CallExpr) (assignments []string, relatedVars map[string]struct{}) {
	if decl == nil || arg == nil {
		return
	}

	src := formatNode(fset, arg)
	exprs := findRelatedExprs(fset, arg)

	if len(exprs) == 0 {
		return
	}

	assignmentStmts := make(map[ast.Stmt]struct{})

	for _, expr := range exprs {
		// Find the last assignment for ident.
		var stmt, lastStmt ast.Stmt
		done := false
		ast.Inspect(decl, func(n ast.Node) bool {
			if n == nil || done {
				return false
			}

			if pos := fset.Position(n.Pos()); pos.Line >= line {
				done = true
				return false
			}

			if node, ok := n.(ast.Stmt); ok {
				stmt = node
			}

			switch node := n.(type) {
			case *ast.AssignStmt:
				for _, left := range node.Lhs {
					switch n := left.(type) {
					case *ast.Ident:
						if isRelated(fset, expr, n) {
							lastStmt = stmt
							return true
						}
					}
				}
			case *ast.RangeStmt:
				if node.Key == nil {
					return true
				}

				switch n := node.Key.(type) {
				case *ast.Ident:
					if isRelated(fset, expr, n) {
						lastStmt = stmt
						return true
					}
				}

				if node.Value == nil {
					return true
				}

				switch n := node.Value.(type) {
				case *ast.Ident:
					if isRelated(fset, expr, n) {
						lastStmt = stmt
						return true
					}
				}
			case *ast.CallExpr:
				for _, call := range excluded {
					if node.Pos() == call.Pos() {
						return false
					}
				}

				for _, arg := range node.Args {
					switch n := arg.(type) {
					case *ast.UnaryExpr:
						// Treat `&a` as a kind of assignment to `a`.
						if n.Op == token.AND && isRelated(fset, expr, n.X) {
							lastStmt = stmt
							return true
						}
					}
				}
			}

			return true
		})

		if lastStmt != nil {
			assignmentStmts[lastStmt] = struct{}{}
		}
	}

	// Collect all stmts and exprs to find out related vars.
	stmts := make([]ast.Stmt, 0, len(assignmentStmts))
	relatedExprs := make([]ast.Expr, 0, 4*len(assignmentStmts))
	relatedExprs = append(relatedExprs, arg)

	for s := range assignmentStmts {
		switch assign := s.(type) {
		case *ast.AssignStmt:
			stmts = append(stmts, assign)
			relatedExprs = append(relatedExprs, assign.Lhs...)
			relatedExprs = append(relatedExprs, assign.Rhs...)
		case *ast.RangeStmt:
			// For RangeStmt, only use the code like `k, v := range arr`.
			stmt := *assign
			body := *assign.Body
			body.List = nil
			stmt.Body = &body

			stmts = append(stmts, &stmt)
			relatedExprs = append(relatedExprs, assign.Key)

			if assign.Value != nil {
				relatedExprs = append(relatedExprs, assign.Value)
			}
		default:
			stmts = append(stmts, s)
		}
	}

	// Find out related vars referenced in expr.
	// If there is a SelectorExpr, use the longest possible selector.
	// E.g., for expr `a.b.c.d()`, only the longest selector expr `a.b.c` is returned.
	relatedVars = make(map[string]struct{})

	for _, expr := range relatedExprs {
		related := findRelatedExprs(fset, expr)

		for _, n := range related {
			relatedVars[formatNode(fset, n)] = struct{}{}
		}
	}

	// Remove arg itself.
	delete(relatedVars, src)

	// Format source code of all assignments.
	sort.Sort(sortByStmts(stmts))
	for _, stmt := range stmts {
		code := formatNode(fset, stmt)

		// Remove keyword `for` and `{}` in RangeStmt.
		if rng, ok := stmt.(*ast.RangeStmt); ok {
			start := rng.Pos()
			code = code[rng.Key.Pos()-start : rng.X.End()-start]
		}

		assignments = append(assignments, code)
	}

	return
}

type sortByStmts []ast.Stmt

func (stmts sortByStmts) Len() int           { return len(stmts) }
func (stmts sortByStmts) Less(i, j int) bool { return stmts[i].Pos() < stmts[j].Pos() }
func (stmts sortByStmts) Swap(i, j int)      { stmts[i], stmts[j] = stmts[j], stmts[i] }

// findIdents parses arg to find all referenced idents in arg.
func findIdents(fset *token.FileSet, arg ast.Expr) (idents []string) {
	names := make(map[string]struct{})
	related := findRelatedExprs(fset, arg)

	for _, expr := range related {
		names[formatNode(fset, expr)] = struct{}{}
	}

	for name := range names {
		idents = append(idents, name)
	}

	return
}

type exprVisitor struct {
	Related map[ast.Expr]struct{}
}

func newExprVisitor() *exprVisitor {
	return &exprVisitor{
		Related: make(map[ast.Expr]struct{}),
	}
}

func (v *exprVisitor) Visit(n ast.Node) (w ast.Visitor) {
	if n == nil {
		return nil
	}

	switch node := n.(type) {
	case *ast.SelectorExpr:
		// Find out the longest selector expr.
		// For code `a.b.c().d.e.f()`, only `a.b` is considered as related var.
		if IsVar(node.X) {
			v.Related[node] = struct{}{}

			// No need to inspect this node any more.
			return nil
		}

		// Never walk node.Sel.
		ast.Walk(v, node.X)
		return nil
	case *ast.Ident:
		v.Related[node] = struct{}{}
		return nil
	}

	return v
}

func findRelatedExprs(fset *token.FileSet, arg ast.Expr) (related []ast.Expr) {
	v := newExprVisitor()
	ast.Walk(v, arg)

	related = make([]ast.Expr, 0, len(v.Related))

	for expr := range v.Related {
		related = append(related, expr)
	}

	return
}

// IsVar returns true if expr is an ident or a selector expr like `a.b`.
func IsVar(expr ast.Expr) bool {
	switch n := expr.(type) {
	case *ast.Ident:
		return true
	case *ast.SelectorExpr:
		x := n.X

		for {
			if sel, ok := x.(*ast.SelectorExpr); ok {
				x = sel
			} else {
				break
			}
		}

		if _, ok := x.(*ast.Ident); ok {
			return true
		}
	}

	return false
}

// isRelated returns true, if target is the same as expr or "parent" of expr.
func isRelated(fset *token.FileSet, expr, target ast.Expr) bool {
	if expr == target {
		return true
	}

	if !IsVar(target) {
		return false
	}

	// target must be a selector or ident.
	switch n := target.(type) {
	case *ast.SelectorExpr:
		if _, ok := expr.(*ast.Ident); ok {
			return false
		}

		return IsIncluded(formatNode(fset, n), formatNode(fset, expr))
	case *ast.Ident:
		return IsIncluded(n.Name, formatNode(fset, expr))
	}

	return false
}

// IsIncluded checks whether child var is a children of parent var.
// Regarding the child var `a.b.c`, it's the children of `a`, `a.b` and `a.b.c`.
func IsIncluded(parent, child string) bool {
	if len(child) < len(parent) {
		return false
	}

	if parent == child {
		return true
	}

	if strings.HasPrefix(child, parent) && child[len(parent)] == '.' {
		return true
	}

	return false
}

// AddExcluded adds an expr to excluded expr list so that
// this expr will not be inspected when finding related assignments.
func (p *Parser) AddExcluded(expr *ast.CallExpr) {
	p.m.Lock()
	defer p.m.Unlock()
	p.excluded = append(p.excluded, expr)
}
