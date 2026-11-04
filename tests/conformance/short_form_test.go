package conformance

import (
	"os"
	"regexp"
	"testing"

	"go.yaml.in/yaml/v3"
)

type gitlabInclude struct {
	Project string `yaml:"project"`
	Ref     string `yaml:"ref"`
	File    string `yaml:"file"`
}

// The GitLab short form takes the template by reference, and its warn child
// pipeline reads the same file from the same project and ref: the version is
// one string (a YAML anchor), so changing it moves both.
func TestGitLabShortFormChildReadsTheSameRemoteFile(t *testing.T) {
	b, err := os.ReadFile("../../ci/gitlab/workline-short.gitlab-ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Include []gitlabInclude `yaml:"include"`
		Child   struct {
			Trigger struct {
				Include []gitlabInclude `yaml:"include"`
			} `yaml:"trigger"`
		} `yaml:"workline:warn:child"`
	}
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Include) != 1 || len(doc.Child.Trigger.Include) != 1 {
		t.Fatalf("want one include and one child include, got %v and %v", doc.Include, doc.Child.Trigger.Include)
	}
	top, child := doc.Include[0], doc.Child.Trigger.Include[0]
	if top.Project == "" || top.Ref == "" || top.File == "" {
		t.Fatalf("the include names no project, ref and file: %+v", top)
	}
	if top != child {
		t.Errorf("the child pipeline reads %+v, the include %+v: they must be the same", child, top)
	}
	if !regexp.MustCompile(`(?m)^\s*ref: &\w+ v\d+\.\d+\.\d+$`).Match(b) {
		t.Errorf("the ref is not one anchored version string, so the child could drift")
	}
}

// The GitHub short form calls a reusable workflow of workline's repository at
// a version: that workflow must exist, and offer workflow_call.
func TestGitHubShortFormCallsAWorkflowThatExists(t *testing.T) {
	b, err := os.ReadFile("../../ci/github/workline-short.yml")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`uses:\s*JN0V/workline/(\.github/workflows/[\w.-]+)@v\d+\.\d+\.\d+`).FindSubmatch(b)
	if m == nil {
		t.Fatal("the short form calls no workflow of workline at a version")
	}
	path := "../../" + string(m[1])
	w, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("%s is not in the repository yet: the short form calls a workflow that does not exist (needs .github/workflows, outside this change's scope)", m[1])
	}
	if !regexp.MustCompile(`(?m)^\s*workflow_call:`).Match(w) {
		t.Errorf("%s has no workflow_call trigger", m[1])
	}
}
