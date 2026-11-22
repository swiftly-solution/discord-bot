package docs

import (
	"fmt"
	"strings"
)

func describeSearch(category, name string, r any) string {
	var b strings.Builder
	link := func(e any, label, suffix string) {
		fmt.Fprintf(&b, "[%s](%s)%s\n", label, str(get(e, "url"), ""), suffix)
	}
	results := func(suffix func(e any) string) {
		b.WriteString("**Results**\n")
		for _, e := range first(list(r), 25) {
			link(e, str(get(e, "name"), ""), suffix(e))
		}
	}
	kind := func(e any) string { return " - " + capitalize(str(get(e, "kind"), "")) }
	fmt.Fprintf(&b, "Search query: `%s`\n\n", name)

	switch category {
	case "schema":
		byModule(&b, first(list(get(r, "classes")), 25), "project", kind)
	case "entity":
		if classes := first(list(get(r, "classes")), 25); len(classes) > 0 {
			b.WriteString("**Classes**\n")
			for _, e := range classes {
				link(e, str(get(e, "name"), ""), "")
			}
			b.WriteString("\n")
		}
		if fields := first(list(get(r, "fields")), 25); len(fields) > 0 {
			b.WriteString("**Fields**\n")
			for _, e := range fields {
				link(e, str(get(e, "className"), "")+"::"+str(get(e, "fieldName"), ""), "")
			}
		}
	case "gameevent":
		results(func(e any) string { return " - " + ticked(list(get(e, "files")), ", ") })
	case "protobuf":
		results(kind)
	case "convar":
		byModule(&b, first(list(get(r, "items")), 25), "module", kind)
	case "panorama":
		results(func(e any) string { return " - " + str(get(e, "snippet"), "") })
	}
	return b.String()
}

// byModule lists items under a header per distinct key, in first-seen order.
func byModule(b *strings.Builder, items []any, key string, suffix func(e any) string) {
	var order []string
	groups := map[string][]any{}
	for _, e := range items {
		k := str(get(e, key), "undefined")
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], e)
	}
	for _, k := range order {
		fmt.Fprintf(b, "**Module - %s**\n", k)
		for _, e := range groups[k] {
			fmt.Fprintf(b, "[%s](%s)%s\n", str(get(e, "name"), ""), str(get(e, "url"), ""), suffix(e))
		}
	}
}
