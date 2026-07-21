// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// JSONKeyFunc transforms a long flag name to a JSON object key.
// The input is the option's long tag value, which by convention is kebab-case.
type JSONKeyFunc func(longName string) string

// JSONKeyLong returns the long name unchanged.
//
//	"my-long-flag" -> "my-long-flag"
func JSONKeyLong(s string) string { return s }

// JSONKeyCamel converts a kebab-case long name to camelCase.
//
//	"my-long-flag" -> "myLongFlag"
func JSONKeyCamel(s string) string {
	parts := strings.Split(s, "-")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return strings.Join(parts, "")
}

// JSONKeySnake converts a kebab-case long name to snake_case.
//
//	"my-long-flag" -> "my_long_flag"
func JSONKeySnake(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

// JSONKeyPascal converts a kebab-case long name to PascalCase.
//
//	"my-long-flag" -> "MyLongFlag"
func JSONKeyPascal(s string) string {
	parts := strings.Split(s, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}

	return strings.Join(parts, "")
}

// JSONParser reads and writes flag values in JSON format. It mirrors the
// IniParser API and can be used before Parse() to load configuration files.
type JSONParser struct {
	parser *Parser

	// tx tracks per-option pre-apply state during ParseMap so the whole
	// apply pass can be rolled back if a later key fails to convert.
	tx *configApplyTransaction

	// KeyName transforms a long flag name to a JSON key when no standard Go
	// json struct tag is present on the field. Defaults to JSONKeyLong.
	KeyName JSONKeyFunc

	// ParseAsDefaults treats JSON values as default values that can be
	// overridden by command-line flags. When false (the default), JSON values
	// are applied as if the user had set them explicitly.
	ParseAsDefaults bool
}

// SetJSONKeyName sets the key naming function used by the built-in config
// command when rendering JSON output. This lets the application control how
// long flag names are mapped to JSON keys without exposing that choice as a
// user-facing CLI flag.
//
// Example: p.SetJSONKeyName(flags.JSONKeyPascal)
func (p *Parser) SetJSONKeyName(fn JSONKeyFunc) {
	p.jsonKeyName = fn
}

// NewJSONParser creates a new JSON parser backed by the given Parser.
func NewJSONParser(p *Parser) *JSONParser {
	return &JSONParser{
		parser:  p,
		KeyName: JSONKeyLong,
	}
}

// ParseFile reads flag values from a JSON file.
func (j *JSONParser) ParseFile(filename string) error {
	f, err := os.Open(filename) //nolint:gosec // filename is a user-supplied config file path.
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return j.Parse(f)
}

// Parse reads flag values from a JSON reader.
func (j *JSONParser) Parse(reader io.Reader) error {
	var data map[string]any

	if err := json.NewDecoder(reader).Decode(&data); err != nil {
		return err
	}

	return j.ParseMap(data)
}

// ParseMap applies a pre-decoded map to the parser's flags. This is the core
// implementation used by Parse and ParseFile.
//
// It also acts as a bridge for other configuration formats: decode YAML, TOML,
// or any map-based format with an external library and pass the result here.
//
// Application is transactional: if any key fails to convert, every option
// touched earlier in this call is rolled back to its pre-apply state before
// the error is returned, so a partially-invalid config file never leaves the
// bound struct partially modified.
func (j *JSONParser) ParseMap(data map[string]any) error {
	j.parser.eachOption(func(_ *Command, _ *Group, option *Option) {
		option.clearReferenceBeforeSet = true
	})

	j.tx = newConfigApplyTransaction()
	defer func() { j.tx = nil }()

	if err := j.applyCommand(j.parser.Command, data); err != nil {
		j.tx.rollback()
		return err
	}

	return nil
}

// applyCommand applies JSON data to a command's groups and subcommands.
// Subcommands are represented as nested objects keyed by command name.
func (j *JSONParser) applyCommand(cmd *Command, data map[string]any) error {
	if err := j.applyGroup(cmd.Group, data); err != nil {
		return err
	}

	for _, subcmd := range cmd.commands {
		val, exists := data[subcmd.Name]
		if !exists {
			continue
		}

		sub, ok := val.(map[string]any)
		if !ok {
			continue
		}

		if err := j.applyCommand(subcmd, sub); err != nil {
			return err
		}
	}

	return nil
}

// WriteFile writes the current flag values as a JSON file.
func (j *JSONParser) WriteFile(filename string) error {
	return writeFileAtomically(filename, j.Write)
}

// Write encodes the current flag values as indented JSON to writer.
func (j *JSONParser) Write(writer io.Writer) error {
	data := j.buildCommandMap(j.parser.Command)
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")

	return enc.Encode(data)
}

// jsonKeyForOption resolves the JSON key for an option.
//
// Priority:
//  1. Standard Go json struct tag - json:"name,..."; json:"-" means excluded.
//  2. Long flag name passed through j.KeyName (defaults to identity).
//  3. Struct field name as last resort.
//
// Returns ("", false) when the field is excluded (json:"-").
func (j *JSONParser) jsonKeyForOption(opt *Option) (string, bool) {
	if tag := opt.field.Tag.Get("json"); tag != "" {
		name, _, _ := strings.Cut(tag, ",")
		if name == "-" {
			return "", false
		}

		if name != "" {
			return name, true
		}
	}

	if opt.LongName != "" {
		keyFn := j.KeyName
		if keyFn == nil {
			keyFn = JSONKeyLong
		}

		return keyFn(opt.LongName), true
	}

	return opt.field.Name, true
}

// jsonGroupKey resolves the JSON object key for a group.
// Returns "" when the group has no explicit namespace or ini-name, meaning its
// options surface at the parent level (anonymous group in JSON context).
func jsonGroupKey(g *Group) string {
	if g.Namespace != "" {
		return g.Namespace
	}

	if g.IniName != "" {
		return g.IniName
	}

	return ""
}

// applyGroup applies JSON data to a group and its subgroups recursively.
func (j *JSONParser) applyGroup(group *Group, data map[string]any) error {
	for _, opt := range group.options {
		if opt.isFunc() || opt.Hidden {
			continue
		}

		key, ok := j.jsonKeyForOption(opt)
		if !ok {
			continue
		}

		val, exists := data[key]
		if !exists || val == nil {
			continue
		}

		if j.ParseAsDefaults && opt.preventDefault {
			continue
		}

		if err := j.applyValue(opt, val); err != nil {
			return err
		}
	}

	for _, sg := range group.groups {
		gkey := jsonGroupKey(sg)

		if gkey == "" {
			if err := j.applyGroup(sg, data); err != nil {
				return err
			}

			continue
		}

		val, exists := data[gkey]
		if !exists {
			continue
		}

		sub, ok := val.(map[string]any)
		if !ok {
			return newErrorf(ErrUnknown,
				"JSON key %q: expected an object for group namespace, got %T",
				gkey, val,
			)
		}

		if err := j.applyGroup(sg, sub); err != nil {
			return err
		}
	}

	return nil
}

// applyValue dispatches JSON value handling based on the option's Go kind.
func (j *JSONParser) applyValue(opt *Option, val any) error {
	kind := opt.value.Type().Kind()

	switch kind {
	case reflect.Slice:
		arr, ok := val.([]any)
		if !ok {
			return j.applyScalar(opt, val)
		}

		for _, item := range arr {
			if err := j.applyScalar(opt, item); err != nil {
				return err
			}
		}

	case reflect.Map:
		sub, ok := val.(map[string]any)
		if !ok {
			return nil
		}

		delim := opt.tag.Get(FlagTagKeyValueDelimiter)
		if delim == "" {
			delim = ":"
		}

		for k, v := range sub {
			entry := k + delim + jsonScalarToString(v)

			if err := j.applySetValue(opt, &entry); err != nil {
				return err
			}
		}

	default:
		return j.applyScalar(opt, val)
	}

	return nil
}

// applyScalar converts a scalar JSON value to string and applies it.
func (j *JSONParser) applyScalar(opt *Option, val any) error {
	s := jsonScalarToString(val)

	return j.applySetValue(opt, &s)
}

// applySetValue calls Set or setDefault depending on ParseAsDefaults.
func (j *JSONParser) applySetValue(opt *Option, val *string) error {
	j.tx.touch(opt)

	if !opt.canArgument() && (val == nil || *val == "") {
		val = nil
	}

	if j.ParseAsDefaults {
		if err := opt.setDefault(val); err != nil {
			return err
		}
		// Prevent clearDefault from overwriting the config value with the
		// struct-tag default during ParseArgs initialization.
		// CLI flags still win because Set() always sets preventDefault = true.
		opt.preventDefault = true

		return nil
	}

	return opt.Set(val)
}

// jsonScalarToString converts a scalar JSON value to its string representation.
func jsonScalarToString(val any) string {
	switch v := val.(type) {
	case string:
		return v

	case bool:
		if v {
			return "true"
		}
		return "false"

	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)

	case nil:
		return ""

	default:
		return fmt.Sprintf("%v", v)
	}
}

// buildCommandMap builds a JSON-serializable map for a command, its groups,
// and all subcommands (represented as nested objects keyed by command name).
func (j *JSONParser) buildCommandMap(cmd *Command) map[string]any {
	out := j.buildGroupMap(cmd.Group)

	for _, subcmd := range cmd.commands {
		if subcmd.Hidden {
			continue
		}

		sub := j.buildCommandMap(subcmd)
		if len(sub) > 0 {
			out[subcmd.Name] = sub
		}
	}

	return out
}

// buildGroupMap builds a JSON-serializable map for a group and its subgroups.
func (j *JSONParser) buildGroupMap(group *Group) map[string]any {
	out := make(map[string]any)

	for _, opt := range group.options {
		if opt.isFunc() || opt.Hidden {
			continue
		}

		key, ok := j.jsonKeyForOption(opt)
		if !ok {
			continue
		}

		if j.jsonOmit(opt) {
			continue
		}

		out[key] = j.optionJSONValue(opt)
	}

	for _, sg := range group.groups {
		gkey := jsonGroupKey(sg)
		sub := j.buildGroupMap(sg)

		if gkey == "" {
			maps.Copy(out, sub)
		} else if len(sub) > 0 {
			out[gkey] = sub
		}
	}

	return out
}

// jsonOmit reports whether the option should be omitted from JSON output
// based on omitempty or omitzero in its json struct tag.
func (j *JSONParser) jsonOmit(opt *Option) bool {
	tag := opt.field.Tag.Get("json")
	if tag == "" {
		return false
	}

	_, tagOpts, _ := strings.Cut(tag, ",")
	if tagOpts == "" {
		return false
	}

	for p := range strings.SplitSeq(tagOpts, ",") {
		if p == "omitempty" || p == "omitzero" {
			return opt.value.IsZero()
		}
	}

	return false
}

// optionJSONValue returns a JSON-serializable value for the option.
// Booleans and numbers are returned as native Go types; everything else as string.
func (j *JSONParser) optionJSONValue(opt *Option) any {
	val := opt.value
	kind := val.Type().Kind()

	switch kind {
	case reflect.Bool:
		return val.Bool()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int()

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint()

	case reflect.Float32, reflect.Float64:
		return val.Float()

	case reflect.String:
		return val.String()

	case reflect.Slice:
		result := make([]any, val.Len())
		for i := range val.Len() {
			s, _ := convertToString(val.Index(i), opt.tag)
			result[i] = s
		}
		return result

	case reflect.Map:
		result := make(map[string]any)
		for _, k := range val.MapKeys() {
			ks, _ := convertToString(k, opt.tag)
			vs, _ := convertToString(val.MapIndex(k), opt.tag)
			result[ks] = vs
		}
		return result

	default:
		s, _ := convertToString(val, opt.tag)
		return s
	}
}
