package profile

// Change is one field a profile change rewrites, in the order
// .odachi/profile.toml uses. The human reads this: the field name is the TOML
// key, because the file is the thing being changed.
type Change struct{ Field, From, To string }

// Evidence is what the project shows about the human's progress. VisionAnswers
// counts the decisions only the human could make: the questions they answered
// and the requirements they changed. Craft decisions are counted because the
// tutor reports them, but they never buy a graduation: the graduation is about
// vision decided, not work delivered.
type Evidence struct{ Features, VisionAnswers, CraftDecisions int }

// nextProfile is the only profile the tool ever proposes by itself.
const nextProfile = "senior"

// The graduation thresholds of Contract IV, in one place.
const (
	featuresForSenior = 3
	answersForSenior  = 5
)

// Next is the graduation proposal: offered, never taken. It proposes senior
// only from novato, and only with three features that carry a done step and
// five answered vision questions. Anything else proposes nothing, so a senior
// is never told to become a senior.
func Next(current string, e Evidence) (string, bool) {
	if current != "novato" {
		return "", false
	}
	if e.Features < featuresForSenior || e.VisionAnswers < answersForSenior {
		return "", false
	}
	return nextProfile, true
}

// changeField is one row of the profile, read for comparison.
type changeField struct {
	field string
	value func(Profile) string
}

// changeFields is the eight fields of Contract A, in file order.
var changeFields = []changeField{
	{"name", func(p Profile) string { return p.Name }},
	{"extends", func(p Profile) string { return p.Extends }},
	{"brief_source", func(p Profile) string { return p.BriefSource }},
	{"questions", func(p Profile) string { return p.Questions }},
	{"assumptions", func(p Profile) string { return p.Assumptions }},
	{"output", func(p Profile) string { return p.Output }},
	{"proof", func(p Profile) string { return p.Proof }},
	{"explain", func(p Profile) string { return p.Explain }},
}

// Apply returns the profile the human accepted and the fields that changed,
// so the tool can say exactly what it did instead of "listo".
func Apply(current, to Profile) (Profile, []Change) {
	changes := make([]Change, 0, len(changeFields))
	for _, f := range changeFields {
		from, target := f.value(current), f.value(to)
		if from == target {
			continue
		}
		changes = append(changes, Change{Field: f.field, From: from, To: target})
	}
	return to, changes
}
