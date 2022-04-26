# go-sfomuseum-writer

Common methods for writing SFO Museum (Who's On First) documents.

## Examples

_Note that error handling has been removed for the sake of brevity._

### WriteFeature

```
import (
	"context"
	"flag"
	"github.com/whosonfirst/go-writer"	
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/feature"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer"	
)

func main() {

	flag.Parse()

	ctx := context.Background()
	wr, _ := writer.NewWriter(ctx, "stdout://")
	
	for _, feature_path := range flag.Args() {
	
		fh, _ := os.Open(feature_path)
		f, _ := feature.LoadWOFFeatureFromReader(fh)

		sfom_writer.WriteFeature(ctx, wr, f)
	}
```

### WriteFeatureBytes

```
import (
	"context"
	"flag"
	"github.com/whosonfirst/go-writer"	
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer"
	"io/ioutil"
)

func main() {

	flag.Parse()

	ctx := context.Background()
	wr, _ := writer.NewWriter(ctx, "stdout://")
	
	for _, feature_path := range flag.Args() {
	
		fh, _ := os.Open(feature_path)
		body, _ := ioutil.ReadAll(fh)
		
		sfom_writer.WriteFeatureBytes(ctx, wr, body)
	}
```

## See also

* https://github.com/whosonfirst/go-writer