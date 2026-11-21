package undone

import "strings"

// AddedNames is the identifiers the added side of a diff declares: a func, a
// method, a type, a var or a const written on one line, and a field of a struct
// when the diff says the line is inside one. It is what the row reads when a
// reverted copy does not build -the compiler names the symbol the change took
// away, and a failure that names nothing the change added is a failure for
// another reason- (0300, 0341).
//
// The limit is written here instead of hidden: a grouped declaration -`const (`
// with its entries on their own lines- is not read, because an entry carries no
// keyword of its own and guessing from a line of a block is how any `name = value`
// later in the file becomes a declaration. A field is read only inside a struct
// scope, which git writes in the hunk header (`@@ ... @@ type State struct {`) or
// as an added `type ... struct {`; a `Name Type` line in any other hunk is a guess
// and names nobody. A qualified field (`io.Reader`) and a line carrying `=` are
// left out too. A failure that names one of those stays unknown, which is the
// safe direction.
func AddedNames(diff string) []string {
	seen := map[string]bool{}
	var names []string
	inStruct := false
	add := func(name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		names = append(names, name)
	}
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "@@") {
			inStruct = structScope(hunkLabel(line))
			continue
		}
		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			continue
		}
		content := line[1:]
		text := strings.TrimSpace(content)
		if strings.HasPrefix(text, "type ") && strings.Contains(text, "struct {") {
			inStruct = true
		}
		if inStruct && content == "}" {
			inStruct = false
		}
		add(declaredName(text))
		if inStruct {
			add(structFieldName(text))
		}
	}
	return names
}

// hunkLabel is the declaration git writes after the second `@@` of a hunk
// header: the nearest enclosing func or type, and the only scope the diff names.
func hunkLabel(header string) string {
	at := strings.LastIndex(header, "@@")
	if at < 0 {
		return ""
	}
	return strings.TrimSpace(header[at+2:])
}

// structScope reports whether a hunk label opens a struct, and nothing else: a
// func label is not a struct, and a comment that mentions one is not a scope.
func structScope(label string) bool {
	return strings.HasPrefix(label, "type ") && strings.Contains(label, "struct {")
}

// structFieldName is the name of a field line inside a struct scope. A leading
// `Name Type` is a field only there; the line is refused when a type follows a
// dot (an embedded `io.Reader`), when `=` makes it an assignment, or when the
// text before the type is itself a keyword. The empty rest is an embedded field,
// whose name is the leading identifier.
func structFieldName(line string) string {
	if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
		return ""
	}
	name := identifier(line)
	if name == "" {
		return ""
	}
	rest := line[len(name):]
	if strings.HasPrefix(rest, ".") {
		return ""
	}
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return name
	}
	if strings.Contains(rest, "=") {
		return ""
	}
	switch rest[0] {
	case '=', ':', '+', '-', '.', ',', '(', '{', '}':
		return ""
	}
	return name
}

// declaredName reads the identifier a Go declaration line introduces.
func declaredName(line string) string {
	switch {
	case strings.HasPrefix(line, "func "):
		rest := strings.TrimSpace(strings.TrimPrefix(line, "func "))
		if strings.HasPrefix(rest, "(") {
			close := strings.Index(rest, ")")
			if close < 0 {
				return ""
			}
			rest = strings.TrimSpace(rest[close+1:])
		}
		return identifier(rest)
	case strings.HasPrefix(line, "type "):
		return identifier(strings.TrimSpace(strings.TrimPrefix(line, "type ")))
	case strings.HasPrefix(line, "const "):
		return identifier(strings.TrimSpace(strings.TrimPrefix(line, "const ")))
	case strings.HasPrefix(line, "var "):
		return identifier(strings.TrimSpace(strings.TrimPrefix(line, "var ")))
	}
	return ""
}

// identifier is the leading Go identifier of text, if there is one. A declaration
// that opens a group -"const ("- names nothing, and neither does a keyword.
func identifier(text string) string {
	end := 0
	for end < len(text) {
		c := text[end]
		letter := c == '_' || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
		digit := end > 0 && '0' <= c && c <= '9'
		if !letter && !digit {
			break
		}
		end++
	}
	name := text[:end]
	if name == "" || goKeywords[name] {
		return ""
	}
	return name
}

// goKeywords are the words that open a line but name no symbol of their own.
var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}
