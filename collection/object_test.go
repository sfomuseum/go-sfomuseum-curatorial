package collection

import (
	"context"
	"testing"
)

func TestFindCurrentObjects(t *testing.T) {

	tests := map[string]int64{
		// "123": 1226605575,
	}

	ctx := context.Background()

	for code, id := range tests {

		g, err := FindCurrentObject(ctx, code)

		if err != nil {
			t.Fatalf("Failed to find current object for %s, %v", code, err)
		}

		if g.WhosOnFirstId != id {
			t.Fatalf("Unexpected ID for object %s. Got %d but expected %d", code, g.WhosOnFirstId, id)
		}
	}
}
