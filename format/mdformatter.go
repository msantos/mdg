// https://github.com/bwplotka/mdox/blob/main/LICENSE
package format

import (
	"bytes"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// https://github.com/bwplotka/mdox/blob/818772c6517630714db8b132cb94f25f72a38850/pkg/mdformatter/mdformatter.go
func FormatFrontMatter(m map[string]interface{}) ([]byte, error) {
	if len(m) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	f := sortedFrontMatter{
		m:    m,
		keys: keys,
	}

	b := bytes.NewBuffer([]byte("---\n"))
	o, err := yaml.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("marshall front matter: %w", err)
	}
	_, _ = b.Write(o)
	_, _ = b.Write([]byte("---\n\n"))
	return b.Bytes(), nil
}

var _ yaml.Marshaler = sortedFrontMatter{}

type sortedFrontMatter struct {
	m    map[string]interface{}
	keys []string
}

func (f sortedFrontMatter) MarshalYAML() (interface{}, error) {
	n := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	for _, k := range f.keys {
		n.Content = append(n.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: k})

		b, err := yaml.Marshal(f.m[k])
		if err != nil {
			return nil, fmt.Errorf("map marshal: %w", err)
		}
		v := &yaml.Node{}
		if err := yaml.Unmarshal(b, v); err != nil {
			return nil, err
		}

		// We expect a node of type document with single content containing other nodes.
		if len(v.Content) != 1 {
			return nil, fmt.Errorf("unexpected node after unmarshalling interface: %#v", v)
		}
		// TODO(bwplotka): This creates weird indentation, fix it.
		n.Content = append(n.Content, v.Content[0])
	}
	return n, nil
}
