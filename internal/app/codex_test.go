package app

import "testing"

func TestExplainPromptIsStable(t *testing.T) {
	if ExplainPrompt == "" {
		t.Fatal("ExplainPrompt must not be empty")
	}
}
