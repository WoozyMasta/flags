// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type valueValidationConfig struct {
	regex   *regexp.Regexp
	pattern string

	minLen *int
	maxLen *int
	min    *validationNumber
	max    *validationNumber

	existingFile bool
	existingDir  bool
	readable     bool
	writable     bool
	nonEmpty     bool
	pathAbs      bool
}

type validationNumber struct {
	floating float64
	signed   int64
	unsigned uint64
	kind     reflect.Kind
}

func parseValueValidationConfig(tag multiTag, fieldName string, value reflect.Value) (valueValidationConfig, error) {
	cfg := valueValidationConfig{}

	for _, spec := range []struct {
		tagName string
		target  *bool
	}{
		{FlagTagValidateExistingFile, &cfg.existingFile},
		{FlagTagValidateExistingDir, &cfg.existingDir},
		{FlagTagValidateReadable, &cfg.readable},
		{FlagTagValidateWritable, &cfg.writable},
		{FlagTagValidateNonEmpty, &cfg.nonEmpty},
		{FlagTagValidatePathAbs, &cfg.pathAbs},
	} {
		enabled, _, err := parseStructBoolTag(tag, spec.tagName, fieldName)
		if err != nil {
			return cfg, err
		}
		*spec.target = enabled
	}

	if raw := tag.Get(FlagTagValidateRegex); raw != "" {
		re, err := regexp.Compile(raw)
		if err != nil {
			return cfg, newErrorf(ErrInvalidTag,
				"invalid regex value `%s' for tag `%s' on field `%s': %v",
				raw, FlagTagValidateRegex, fieldName, err)
		}
		cfg.regex = re
		cfg.pattern = raw
	}

	var err error
	cfg.minLen, err = parseNonNegativeIntValidationTag(tag, FlagTagValidateMinLen, fieldName)
	if err != nil {
		return cfg, err
	}
	cfg.maxLen, err = parseNonNegativeIntValidationTag(tag, FlagTagValidateMaxLen, fieldName)
	if err != nil {
		return cfg, err
	}
	if cfg.minLen != nil && cfg.maxLen != nil && *cfg.minLen > *cfg.maxLen {
		return cfg, newErrorf(ErrInvalidTag,
			"tag `%s' value %d must be <= tag `%s' value %d on field `%s'",
			FlagTagValidateMinLen, *cfg.minLen, FlagTagValidateMaxLen, *cfg.maxLen, fieldName)
	}

	valueKind := validationValueKind(value.Type())

	cfg.min, err = parseNumericValidationTag(tag, FlagTagValidateMin, fieldName, valueKind)
	if err != nil {
		return cfg, err
	}
	cfg.max, err = parseNumericValidationTag(tag, FlagTagValidateMax, fieldName, valueKind)
	if err != nil {
		return cfg, err
	}
	if cfg.min != nil && cfg.max != nil && compareValidationNumbers(valueKind, *cfg.min, *cfg.max) > 0 {
		return cfg, newErrorf(ErrInvalidTag,
			"tag `%s' value %s must be <= tag `%s' value %s on field `%s'",
			FlagTagValidateMin, tag.Get(FlagTagValidateMin),
			FlagTagValidateMax, tag.Get(FlagTagValidateMax), fieldName)
	}

	if err := validateValidationTargetType(cfg, valueKind, fieldName); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func parseNonNegativeIntValidationTag(tag multiTag, tagName string, fieldName string) (*int, error) {
	raw := tag.Get(tagName)
	if raw == "" {
		return nil, nil
	}

	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return nil, newErrorf(ErrInvalidTag,
			"tag `%s' on field `%s' must be a non-negative integer",
			tagName, fieldName)
	}

	return &n, nil
}

func parseNumericValidationTag(
	tag multiTag,
	tagName string,
	fieldName string,
	kind reflect.Kind,
) (*validationNumber, error) {
	raw := tag.Get(tagName)
	if raw == "" {
		return nil, nil
	}

	n := validationNumber{kind: kind}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		bits := kindBits(kind)
		v, err := strconv.ParseInt(raw, 10, bits)
		if err != nil {
			return nil, invalidNumericValidationTag(tagName, fieldName, raw, kind)
		}
		n.signed = v

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bits := kindBits(kind)
		v, err := strconv.ParseUint(raw, 10, bits)
		if err != nil {
			return nil, invalidNumericValidationTag(tagName, fieldName, raw, kind)
		}
		n.unsigned = v

	case reflect.Float32, reflect.Float64:
		bits := kindBits(kind)
		v, err := strconv.ParseFloat(raw, bits)
		if err != nil {
			return nil, invalidNumericValidationTag(tagName, fieldName, raw, kind)
		}
		n.floating = v

	default:
		return nil, newErrorf(ErrInvalidTag,
			"tag `%s' requires numeric field type on field `%s'",
			tagName, fieldName)
	}

	return &n, nil
}

func invalidNumericValidationTag(tagName string, fieldName string, raw string, kind reflect.Kind) error {
	return newErrorf(ErrInvalidTag,
		"invalid numeric value `%s' for tag `%s' on field `%s' with type `%s'",
		raw, tagName, fieldName, kind)
}

func validateValidationTargetType(cfg valueValidationConfig, kind reflect.Kind, fieldName string) error {
	usesString := cfg.existingFile || cfg.existingDir || cfg.readable || cfg.writable ||
		cfg.nonEmpty || cfg.pathAbs || cfg.regex != nil || cfg.minLen != nil || cfg.maxLen != nil
	if usesString && kind != reflect.String {
		return newErrorf(ErrInvalidTag,
			"string validation tags on field `%s' require string or []string type",
			fieldName)
	}

	usesNumeric := cfg.min != nil || cfg.max != nil
	if usesNumeric && !isNumericKind(kind) {
		return newErrorf(ErrInvalidTag,
			"numeric validation tags on field `%s' require numeric or numeric slice type",
			fieldName)
	}

	return nil
}

func validationValueKind(tp reflect.Type) reflect.Kind {
	if tp.Kind() == reflect.Slice {
		return tp.Elem().Kind()
	}

	return tp.Kind()
}

func kindBits(kind reflect.Kind) int {
	switch kind {
	case reflect.Int8, reflect.Uint8:
		return 8
	case reflect.Int16, reflect.Uint16:
		return 16
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		return 32
	case reflect.Int64, reflect.Uint64, reflect.Float64:
		return 64
	case reflect.Int:
		return strconv.IntSize
	case reflect.Uint:
		return strconv.IntSize
	default:
		return 0
	}
}

func isNumericKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func compareValidationNumbers(kind reflect.Kind, left validationNumber, right validationNumber) int {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cmpOrdered(left.signed, right.signed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cmpOrdered(left.unsigned, right.unsigned)
	case reflect.Float32, reflect.Float64:
		return cmpOrdered(left.floating, right.floating)
	default:
		return 0
	}
}

func cmpOrdered[T ~int64 | ~uint64 | ~float64](left T, right T) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func (p *parseState) checkValueValidators(parser *Parser) error {
	for c := parser.Command; c != nil; c = c.Active {
		if err := p.checkCommandValueValidators(parser, c); err != nil {
			p.err = err
			return err
		}
	}

	return nil
}

func (p *parseState) checkCommandValueValidators(parser *Parser, command *Command) error {
	var err error

	command.eachGroup(func(g *Group) {
		if err != nil {
			return
		}

		for _, option := range g.options {
			if shouldSkipOptionValidation(option) {
				continue
			}
			err = validateReflectValue(parser, option.validation, option.value, option.String())
			if err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	for _, arg := range command.args {
		if arg.isEmpty() {
			continue
		}
		if err := validateReflectValue(parser, arg.validation, arg.value, arg.localizedName()); err != nil {
			return err
		}
	}

	return nil
}

func shouldSkipOptionValidation(option *Option) bool {
	if option.isSet || len(option.Default) > 0 {
		return false
	}
	if envKey := option.EnvKeyWithNamespace(); envKey != "" {
		if _, ok := os.LookupEnv(envKey); ok {
			return false
		}
	}

	return option.isEmpty()
}

func validateReflectValue(
	parser *Parser,
	cfg valueValidationConfig,
	value reflect.Value,
	name string,
) error {
	if !cfg.hasRules() {
		return nil
	}

	if value.Kind() == reflect.Slice {
		for i := 0; i < value.Len(); i++ {
			if err := validateSingleValue(parser, cfg, value.Index(i), name); err != nil {
				return err
			}
		}
		return nil
	}

	return validateSingleValue(parser, cfg, value, name)
}

func (cfg valueValidationConfig) hasRules() bool {
	return cfg.existingFile || cfg.existingDir || cfg.readable || cfg.writable ||
		cfg.nonEmpty || cfg.pathAbs || cfg.regex != nil || cfg.minLen != nil ||
		cfg.maxLen != nil || cfg.min != nil || cfg.max != nil
}

func validateSingleValue(
	parser *Parser,
	cfg valueValidationConfig,
	value reflect.Value,
	name string,
) error {
	switch value.Kind() {
	case reflect.String:
		return validateStringValue(parser, cfg, name, value.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return validateSignedValue(parser, cfg, name, value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return validateUnsignedValue(parser, cfg, name, value.Uint())
	case reflect.Float32, reflect.Float64:
		return validateFloatValue(parser, cfg, name, value.Float())
	default:
		return nil
	}
}

func validateStringValue(parser *Parser, cfg valueValidationConfig, name string, value string) error {
	if cfg.nonEmpty && strings.TrimSpace(value) == "" {
		return validationError(parser, "err.validation.non_empty",
			"value `{value}` for `{name}` must not be empty", name, value)
	}
	if cfg.regex != nil && !regexpMatchesFullValue(cfg.regex, value) {
		return validationErrorWithVars(parser, "err.validation.regex",
			"value `{value}` for `{name}` does not match pattern `{pattern}`",
			map[string]string{
				"name":    name,
				"value":   value,
				"pattern": cfg.pattern,
			})
	}
	if cfg.minLen != nil && utf8.RuneCountInString(value) < *cfg.minLen {
		return validationError(parser, "err.validation.min_len",
			"value `{value}` for `{name}` is shorter than required length", name, value)
	}
	if cfg.maxLen != nil && utf8.RuneCountInString(value) > *cfg.maxLen {
		return validationError(parser, "err.validation.max_len",
			"value `{value}` for `{name}` is longer than allowed length", name, value)
	}
	if cfg.pathAbs && !filepath.IsAbs(value) {
		return validationError(parser, "err.validation.path_abs",
			"path `{value}` for `{name}` must be absolute", name, value)
	}
	if cfg.existingFile {
		if err := validateExistingFile(value); err != nil {
			return validationError(parser, "err.validation.existing_file",
				"path `{value}` for `{name}` must be an existing file", name, value)
		}
	}
	if cfg.existingDir {
		if err := validateExistingDir(value); err != nil {
			return validationError(parser, "err.validation.existing_dir",
				"path `{value}` for `{name}` must be an existing directory", name, value)
		}
	}
	if cfg.readable {
		if err := validateReadable(value); err != nil {
			return validationError(parser, "err.validation.readable",
				"path `{value}` for `{name}` must be readable", name, value)
		}
	}
	if cfg.writable {
		if err := validateWritable(value); err != nil {
			return validationError(parser, "err.validation.writable",
				"path `{value}` for `{name}` must be writable", name, value)
		}
	}

	return nil
}

func regexpMatchesFullValue(re *regexp.Regexp, value string) bool {
	match := re.FindStringIndex(value)
	return match != nil && match[0] == 0 && match[1] == len(value)
}

func validateExistingFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", path)
	}

	return nil
}

func validateExistingDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	return nil
}

func validateReadable(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	return file.Close()
}

func validateWritable(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return validateDirWritable(path)
		}

		file, openErr := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
		if openErr != nil {
			return openErr
		}

		return file.Close()
	}
	if !os.IsNotExist(err) {
		return err
	}

	parent := filepath.Dir(path)
	if parent == "." || parent == "" {
		parent = "."
	}

	return validateDirWritable(parent)
}

func validateDirWritable(dir string) error {
	file, err := os.CreateTemp(dir, ".flags-write-test-*")
	if err != nil {
		return err
	}

	name := file.Name()
	closeErr := file.Close()
	removeErr := os.Remove(name)
	if closeErr != nil {
		return closeErr
	}

	return removeErr
}

func validateSignedValue(parser *Parser, cfg valueValidationConfig, name string, value int64) error {
	if cfg.min != nil && value < cfg.min.signed {
		return validationError(parser, "err.validation.min",
			"value `{value}` for `{name}` is lower than allowed minimum", name, strconv.FormatInt(value, 10))
	}
	if cfg.max != nil && value > cfg.max.signed {
		return validationError(parser, "err.validation.max",
			"value `{value}` for `{name}` is greater than allowed maximum", name, strconv.FormatInt(value, 10))
	}

	return nil
}

func validateUnsignedValue(parser *Parser, cfg valueValidationConfig, name string, value uint64) error {
	if cfg.min != nil && value < cfg.min.unsigned {
		return validationError(parser, "err.validation.min",
			"value `{value}` for `{name}` is lower than allowed minimum", name, strconv.FormatUint(value, 10))
	}
	if cfg.max != nil && value > cfg.max.unsigned {
		return validationError(parser, "err.validation.max",
			"value `{value}` for `{name}` is greater than allowed maximum", name, strconv.FormatUint(value, 10))
	}

	return nil
}

func validateFloatValue(parser *Parser, cfg valueValidationConfig, name string, value float64) error {
	if math.IsNaN(value) {
		return validationError(parser, "err.validation.number",
			"value `{value}` for `{name}` must be a valid number", name, strconv.FormatFloat(value, 'g', -1, 64))
	}
	if cfg.min != nil && value < cfg.min.floating {
		return validationError(parser, "err.validation.min",
			"value `{value}` for `{name}` is lower than allowed minimum", name, strconv.FormatFloat(value, 'g', -1, 64))
	}
	if cfg.max != nil && value > cfg.max.floating {
		return validationError(parser, "err.validation.max",
			"value `{value}` for `{name}` is greater than allowed maximum", name, strconv.FormatFloat(value, 'g', -1, 64))
	}

	return nil
}

func validationError(
	parser *Parser,
	key string,
	source string,
	name string,
	value string,
) error {
	return validationErrorWithVars(parser, key, source, map[string]string{
		"name":  name,
		"value": value,
	})
}

func validationErrorWithVars(
	parser *Parser,
	key string,
	source string,
	vars map[string]string,
) error {
	return newError(ErrValidation, parser.i18nTextf(key, source, vars))
}
