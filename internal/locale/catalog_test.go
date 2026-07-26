package locale

import "testing"

func TestLoadAndFallback(t *testing.T) {
	catalog, err := Load([]byte(`{"language":"zh-TW","strings":{"continue":"繼續"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := catalog.Text("continue", "Continue"); got != "繼續" {
		t.Fatalf("got %q", got)
	}
	if got := catalog.Text("missing", "Continue"); got != "Continue" {
		t.Fatalf("fallback got %q", got)
	}
}
