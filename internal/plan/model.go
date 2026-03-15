package plan

import (
	"time"

	"gopkg.in/yaml.v3"
)

// Task represents a single task in the plan.
type Task struct {
	Text   string   `yaml:"text"`
	Status string   `yaml:"status"`            // todo, active, done, blocked
	Reason string   `yaml:"reason,omitempty"`   // only set when status=blocked
	Spec   string   `yaml:"spec,omitempty"`     // implementation instructions
	Verify []string `yaml:"verify,omitempty"`   // verification steps
	Files  []string `yaml:"files,omitempty"`    // files this task touches
}

// FileRef is a project-level file reference with its role.
type FileRef struct {
	Path string `yaml:"path"`
	Role string `yaml:"role"`
}

// Decision represents a timestamped decision.
type Decision struct {
	Time time.Time `yaml:"time"`
	Text string    `yaml:"text"`
}

// NullableString is a string that marshals to YAML ~ when empty.
type NullableString string

func (n NullableString) MarshalYAML() (interface{}, error) {
	if n == "" {
		node := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!null",
			Value: "null",
		}
		return node, nil
	}
	return string(n), nil
}

func (n *NullableString) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == "!!null" {
		*n = ""
		return nil
	}
	*n = NullableString(value.Value)
	return nil
}

// Context holds session state fields.
type Context struct {
	CurrentFile     NullableString `yaml:"current_file"`
	LastError       NullableString `yaml:"last_error"`
	TestState       NullableString `yaml:"test_state"`
	OpenQuestions   NullableString `yaml:"open_questions"`
	PendingRefactor NullableString `yaml:"pending_refactor"`
}

// Plan is the complete plan state, stored as YAML frontmatter.
type Plan struct {
	Name         string     `yaml:"name"`
	Goal         string     `yaml:"goal"`
	Branch       string     `yaml:"branch,omitempty"`        // optional — associated git branch
	Status       string     `yaml:"status"`                  // active, complete
	SessionCount int        `yaml:"session_count"`
	Created      string     `yaml:"created"`                 // date string YYYY-MM-DD
	Updated      string     `yaml:"updated"`                 // date string YYYY-MM-DD
	CurrentTask  int        `yaml:"current_task"`
	Constraints  []string   `yaml:"constraints,omitempty"`   // global rules for the agent
	PlanFiles    []FileRef  `yaml:"files,omitempty"`         // key project files
	Tasks        []Task     `yaml:"tasks"`
	Context      Context    `yaml:"context"`
	Decisions    []Decision `yaml:"decisions"`
}
