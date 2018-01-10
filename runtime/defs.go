package runtime

type tokenKind string
type exprKind string
type declKind string

type Case struct {
	expr Expr
	dcls []Declaration
}

type Declaration struct {
	kind  declKind
	key   Token
	value Expr
}

type Expr struct {
	kind  exprKind
	value Token
	args  []Token
}

type Token struct {
	kind   tokenKind
	lexeme string
}

const (
	caseToken       tokenKind = "casetok"
	blockOpenToken  tokenKind = "blockotok"
	identifierToken tokenKind = "idtok"
	defPathToken    tokenKind = "defptok"
	defEqToken      tokenKind = "defeqtok"

	call exprKind = "call"
	expr exprKind = "expr"

	path declKind = "path"
)
