package media

import (
	"github.com/geego/gean/app/docshelper"
)

// This is is just some helpers used to create some JSON used in the Hugo docs.
func init() {
	docsProvider := func() map[string]interface{} {
		docs := make(map[string]interface{})

		docs["types"] = DefaultTypes
		return docs
	}

	docshelper.AddDocProvider("media", docsProvider)
}
