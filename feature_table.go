package gts

import (
	"io"
	"sort"
	"strings"

	pars "gopkg.in/ktnyt/pars.v2"
)

// FeatureTable represents an INSDC feature table. The features are sorted by
// Location in ascending order.
type FeatureTable []Feature

// Len is the number of elements in the feature table.
func (ft FeatureTable) Len() int { return len(ft) }

// Less reports whether the element with index i should sort before the element
// with index j.
func (ft FeatureTable) Less(i, j int) bool {
	a, b := ft[i], ft[j]
	if a.Key == "source" && b.Key != "source" {
		return true
	}
	if b.Key == "source" && a.Key != "source" {
		return false
	}
	return LocationLess(ft[i].Location, ft[j].Location)
}

// Swap the elements with indices i and j.
func (ft FeatureTable) Swap(i, j int) {
	ft[i], ft[j] = ft[j], ft[i]
}

// Format creates a FeatureFormatter object for the qualifier with the given
// prefix and depth.
func (ft FeatureTable) Format(prefix string, depth int) FeatureTableFormatter {
	return FeatureTableFormatter{ft, prefix, depth}
}

// Insert the feature to the feature table at the given position. Note that
// inserting a feature that disrupts the sortedness of the features will
// inevitably lead to predictable yet unconventional behavior when the Add
// method is called later. Use Add instead if this is not desired.
func (ft *FeatureTable) Insert(i int, f Feature) {
	features := append(*ft, Feature{})
	copy(features[i+1:], features[i:])
	features[i] = f
	*ft = features
}

// Add the feature to the feature table. The feature will be inserted in the
// sorted position with the exception of sources.
func (ft *FeatureTable) Add(f Feature) {
	n := 0
	for n < len(*ft) && (*ft)[n].Key == "source" {
		n++
	}

	switch f.Key {
	case "source":
		ft.Insert(n, f)
	default:
		i := sort.Search(len((*ft)[n:]), func(i int) bool {
			return LocationLess(f.Location, (*ft)[n+i].Location)
		})
		ft.Insert(n+i, f)
	}
}

// FeatureTableFormatter will format a FeatureTable object with the given
// prefix and depth.
type FeatureTableFormatter struct {
	FeatureTable FeatureTable
	Prefix       string
	Depth        int
}

// String satisfies the fmt.Stringer interface.
func (ff FeatureTableFormatter) String() string {
	b := strings.Builder{}
	for _, f := range ff.FeatureTable {
		f.Format(ff.Prefix, ff.Depth).WriteTo(&b)
		b.WriteByte('\n')
	}
	return b.String()
}

// WriteTo satisfies the io.WriterTo interface.
func (ff FeatureTableFormatter) WriteTo(w io.Writer) (int, error) {
	return w.Write([]byte(ff.String()))
}

// FeatureTableParser will attempt to match an INSDC feature table.
func FeatureTableParser(prefix string) pars.Parser {
	return pars.Many(FeatureParser(prefix)).Map(func(result *pars.Result) error {
		features := make([]Feature, len(result.Children))
		for i, child := range result.Children {
			features[i] = child.Value.(Feature)
		}
		result.SetValue(FeatureTable(features))
		return nil
	})
}
