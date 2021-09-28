package exhibitions

import (
	"context"
	"github.com/sfomuseum/go-sfomuseum-curatorial"
	"testing"
)

func TestExhibitionsLookup(t *testing.T) {

	wofid_tests := map[string]int64{
		"1845": 1746382277,
	}

	schemes := []string{
		"exhibitions://",
		"exhibitions://github",
	}

	ctx := context.Background()

	for _, s := range schemes {

		lu, err := curatorial.NewLookup(ctx, s)

		if err != nil {
			t.Fatalf("Failed to create lookup for '%s', %v", s, err)
		}

		for code, wofid := range wofid_tests {

			results, err := lu.Find(ctx, code)

			if err != nil {
				t.Fatalf("Unable to find '%s' using scheme '%s', %v", code, s, err)
			}

			if len(results) != 1 {
				t.Fatalf("Invalid results for '%s' using scheme '%s'", code, s)
			}

			a := results[0].(*Exhibition)

			if a.WhosOnFirstId != wofid {
				t.Fatalf("Invalid match for '%s', expected %d but got %d using scheme '%s'", code, wofid, a.WhosOnFirstId, s)
			}
		}
	}

}
