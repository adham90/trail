package plan

import (
	"fmt"
	"strings"
	"unicode"
)

// GenerateTemplate creates a new plan Markdown file with the given name and goal.
func GenerateTemplate(name, goal string) []byte {
	return []byte(fmt.Sprintf(`# %s

%s

## Acceptance Criteria

- [ ] Define acceptance criteria

## Tasks

- [ ] **1.** Define first task
- [ ] **2.** Define second task
- [ ] **3.** Define third task

## Decisions

## Notes
`, name, goal))
}

// SlugToTitle converts a filename slug to a title.
// "deploy-pipeline" → "Deploy Pipeline"
func SlugToTitle(slug string) string {
	words := strings.Split(slug, "-")
	for i, w := range words {
		if len(w) > 0 {
			runes := []rune(w)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}
