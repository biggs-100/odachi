package undone

import "strings"

// AddedNames is the identifiers the added side of a diff declares: a func, a
// method, a type, a var or a const written on one line. It is what the row reads
// when a reverted copy does not build -the compiler names the symbol the change
// took away, and a failure that names nothing the change added is a failure for
// another reason- (0300).
//
// The limit is written here instead of hidden: a grouped declaration -`const (`
// with its entries on their own lines- is not read, because an entry carries no
// keyword of its own and guessing from a line of a block is how any `name = value`
// later in the file becomes a declaration. A failure that names one of those stays
// unknown, which is the safe direction.
func AddedNames(diff string) []string {
	seen := map[string]bool{}
	var names []string
	for _, line := range strings.Split(diff, "\n") {
		if !strings.HasPrefix(line, "+") || strings.HasPrefix(line, "+++") {
			continue
		}
		name := declaredName(strings.TrimSpace(line[1:]))
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names
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
