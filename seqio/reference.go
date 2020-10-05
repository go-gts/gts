package seqio

import (
	"fmt"

	"github.com/go-gts/gts"
	"github.com/go-pars/pars"
)

func parseReferenceInfo(s string) pars.Parser {
	prefix := fmt.Sprintf("(%s ", s)
	parser := pars.Seq(pars.Int, " to ", pars.Int).Map(func(result *pars.Result) error {
		start := result.Children[0].Value.(int) - 1
		end := result.Children[2].Value.(int)
		result.SetValue(gts.Segment{start, end})
		return nil
	})
	return pars.Seq(
		prefix,
		parser,
		pars.Many(pars.Seq("; ", parser).Child(1)),
		')',
	).Map(func(result *pars.Result) error {
		head := result.Children[1].Value.(gts.Segment)
		tail := result.Children[2].Children
		locs := make([]gts.Segment, len(tail)+1)
		locs[0] = head
		for i, r := range tail {
			locs[i+1] = r.Value.(gts.Segment)
		}
		result.SetValue(locs)
		return nil
	})
}

// Reference represents a reference of a record.
type Reference struct {
	Number  int
	Info    string
	Authors string
	Group   string
	Title   string
	Journal string
	Xref    map[string]string
	Comment string
}
