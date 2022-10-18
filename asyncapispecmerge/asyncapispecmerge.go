package asyncapispecmerge

import (
	"context"
	"github.com/lithictech/moxpopuli/asyncapispecmerge/httpmerge"
	"github.com/lithictech/moxpopuli/asyncapispecmerge/internal"
)

var MergeHttp Merge = httpmerge.MergeHttp

type MergeInput = internal.MergeInput
type Merge func(context.Context, MergeInput) error
