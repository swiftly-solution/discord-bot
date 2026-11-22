package docs

import (
	"fmt"
	"strings"
)

var panorama = strings.NewReplacer("<br>", "\n", "<b>", "**", "</b>", "**", "<pre>", "```css\n", "</pre>", "```", "&lt;", "<", "&gt;", ">")

func describeLookup(category, name string, r any) string {
	var b strings.Builder
	line := func(format string, args ...any) { fmt.Fprintf(&b, format+"\n", args...) }
	flags := func(l []any) {
		if len(l) > 0 {
			b.WriteString("\n**Flags**\n" + ticked(l, ","))
		}
	}
	entry := get(r, "entry")
	fmt.Fprintf(&b, "[SwiftlyS2 Documentation](%s)\n\n", str(get(r, "url"), ""))

	switch category {
	case "schema":
		line("```yaml")
		line("%s:", name)
		line("  project: %s", str(get(r, "project"), "server"))
		if get(r, "kind") == "class" {
			line("  kind: class")
			line("  base_classes: %s", str(get(entry, "base_classes_count"), "None"))
			line("  fields: %s", str(get(entry, "fields_count"), "None"))
			line("  size: %s", str(get(entry, "size"), "0"))
			line("  is_struct: %s", str(get(entry, "is_struct"), "false"))
			line("  has_chainer: %s", str(get(entry, "has_chainer"), "false"))
			b.WriteString("```")
			if n, _ := get(entry, "base_classes_count").(float64); n > 0 {
				b.WriteString("\n**Base Classes**\n" + ticked(list(get(entry, "base_classes")), ","))
			}
		} else {
			line("  kind: %s", str(get(r, "kind"), "enum"))
			line("  fields: %s", str(get(entry, "fields_count"), "None"))
			line("  size: %s", str(get(entry, "size"), "0"))
			b.WriteString("```")
		}
	case "entity":
		datamap := get(r, "datamap")
		line("```yaml")
		line("%s:", name)
		line("  designer_name: %s", str(get(r, "entityClass", "designer_name"), "None"))
		line("  datamap:")
		line("    data_class_name: %s", str(get(datamap, "data_class_name"), "None"))
		line("    fields:")
		line("      inputs: %d", len(list(get(datamap, "fields", "inputs"))))
		line("      outputs: %d", len(list(get(datamap, "fields", "outputs"))))
		line("      members: %d", len(list(get(datamap, "fields", "members"))))
		line("    think_functions: %d", len(list(get(datamap, "think_functions"))))
		b.WriteString("```")
		flags(list(get(r, "entityClass", "flags")))
	case "gameevent":
		line("```yaml")
		if c := get(entry, "comment"); c != "" {
			line("# %s", str(c, "No comment"))
		}
		line("%s:", name)
		line("  raw_name: %s", str(get(entry, "name"), "None"))
		line("  files:")
		for _, f := range list(get(entry, "files")) {
			line("    - %s", str(f, ""))
		}
		line("  fields: %d", len(list(get(entry, "fields"))))
		b.WriteString("```")
	case "protobuf":
		line("```yaml")
		line("%s:", name)
		line("  file: %s", str(get(r, "file"), "unknown"))
		if get(r, "kind") == "message" {
			line("  kind: message")
			line("  fields: %d", len(list(get(entry, "fields"))))
		} else {
			line("  kind: %s", str(get(r, "kind"), "enum"))
			line("  values: %d", len(list(get(entry, "values"))))
		}
		b.WriteString("```")
	case "convar":
		attr := get(entry, "attributes")
		line("```yaml")
		line("# %s", str(get(entry, "description"), "No description"))
		line("%s:", name)
		line("  module: %s", str(get(r, "module"), "unknown"))
		if get(r, "kind") == "convar" {
			line("  kind: convar")
			line("  attributes:")
			line("    has_callback: %s", str(get(attr, "has_callback"), "false"))
			line("    has_min: %s", str(get(attr, "has_min"), "false"))
			line("    has_max: %s", str(get(attr, "has_max"), "false"))
			line("    has_default: %s", str(get(attr, "has_default"), "false"))
			line("  default: %s", str(get(entry, "default"), "none"))
		} else {
			line("  kind: %s", str(get(r, "kind"), "concommand"))
			line("  attributes:")
			line("    has_callback: %s", str(get(attr, "has_callback"), "false"))
			line("    has_completion_callback: %s", str(get(attr, "has_completion_callback"), "false"))
		}
		b.WriteString("```")
		flags(list(get(entry, "flags")))
	case "panorama":
		b.WriteString(panorama.Replace(str(get(entry, "description"), "")))
	}
	return b.String()
}
