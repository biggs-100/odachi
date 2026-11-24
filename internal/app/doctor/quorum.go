package doctor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/biggs-100/odachi/internal/adapters/reviewtoml"
	domaindoctor "github.com/biggs-100/odachi/internal/domain/doctor"
	"github.com/biggs-100/odachi/internal/domain/layout"
	"github.com/biggs-100/odachi/internal/domain/review"
)

// quorum is check 17. How many distinct models judge the lenses, and -when the
// answer is one- which model answers all of them.
//
// It is a warning and never a problem: the project decides whether it buys a second
// vendor, and the tool has no opinion about that. What it refuses is the silence
// that makes four answers look like a jury when they are one opinion with four
// names (0316).
func (r *runner) quorum() domaindoctor.Check {
	content, exists, err := r.read(layout.Review)
	if err != nil {
		return r.unreadable(domaindoctor.CheckQuorum, layout.Review, err)
	}
	if !exists {
		return domaindoctor.OK(domaindoctor.CheckQuorum,
			"no "+layout.Review+": with no review there is no quorum to count")
	}

	config, err := reviewtoml.Decode([]byte(content))
	if err != nil {
		return domaindoctor.Broken(domaindoctor.CheckQuorum, oneLine(err.Error()),
			"fix the row the message names in "+layout.Review+", or delete it when this project runs no review")
	}

	distinct, shared := review.Quorum(config.Lenses)
	lenses := len(config.Lenses)
	switch {
	case lenses == 0:
		// A review with no lens table runs every lens with the model the reviewer
		// command names, and this check does not invent a quorum for a table the
		// project never wrote.
		return domaindoctor.OK(domaindoctor.CheckQuorum,
			"no lens table in "+layout.Review+": the reviewer runs with the session's model and no lens has its own")
	case distinct == 1:
		return domaindoctor.Warning(domaindoctor.CheckQuorum, fmt.Sprintf(
			"%s: the quorum is 1 of %d, one opinion in %d costumes; one model is enough to review; a second is optional, put it in a lens row of %s to buy independence",
			shared, lenses, lenses, layout.Review))
	default:
		return domaindoctor.OK(domaindoctor.CheckQuorum,
			fmt.Sprintf("the quorum is %d of %d: %s judge the lenses", distinct, lenses, strings.Join(judgedBy(config.Lenses), ", ")))
	}
}

// judgedBy lists the models the table names, once each and in a fixed order: the
// line names who judges instead of only how many there are.
func judgedBy(models map[string]string) []string {
	seen := make(map[string]bool, len(models))
	names := make([]string, 0, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" || seen[model] {
			continue
		}
		seen[model] = true
		names = append(names, model)
	}
	sort.Strings(names)
	return names
}
