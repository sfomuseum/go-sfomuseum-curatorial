package publicart

import (
	"context"
	"testing"
)

func TestFindCurrentPublicArtWork(t *testing.T) {

	tests := map[string]int64{
		// "109632": 1729829257,
	}

	ctx := context.Background()

	for code, id := range tests {

		g, err := FindCurrentPublicArtWork(ctx, code)

		if err != nil {
			t.Fatalf("Failed to find current public art work for %s, %v", code, err)
		}

		if g.WhosOnFirstId != id {
			t.Fatalf("Unexpected ID for public art work %s. Got %d but expected %d", code, g.WhosOnFirstId, id)
		}
	}
}
