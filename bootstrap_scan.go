// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

// bootstrapScanArgs performs a lightweight pre-parse pass over args to extract the value/presence
// of a small set of known "bootstrap" options (--config, --env-file, --no-env, --env-override)
// before the main parser has fully set itself up (EnsureBuiltinOptions/the main tokenizer loop).
//
// It reuses the same primitives as the main tokenizer (argumentIsOption, stripOptionPrefix, splitOption),
// so it recognizes the same option styles the main parser does (POSIX --long/-s, Windows /long:value),
// honors any long/short aliases registered on the target options,
// and stops scanning at a literal "--" exactly when the main parser would (PassDoubleDash).
//
// valueTargets maps a value-carrying option to the string pointer that receives its value if present;
// boolTargets maps a boolean-flag option to the bool pointer that is set true if the flag is present.
// Options with a nil map entry are ignored. Short-option clusters
// (e.g. -xc combining an unrelated flag with a bootstrap short flag) are not decoded here,
// same as before this scanner existed:
// bootstrap flags are expected to appear on their own (-c, -c=value, -c value).
func (p *Parser) bootstrapScanArgs(args []string, valueTargets map[*Option]*string, boolTargets map[*Option]*bool) {
	longIndex := make(map[string]*Option, len(valueTargets)+len(boolTargets))
	shortIndex := make(map[string]*Option, len(valueTargets)+len(boolTargets))

	register := func(opt *Option) {
		if opt == nil {
			return
		}

		if opt.LongName != "" {
			longIndex[opt.LongName] = opt
		}

		for _, alias := range opt.LongAliases {
			longIndex[alias] = opt
		}

		if opt.ShortName != 0 {
			shortIndex[string(opt.ShortName)] = opt
		}

		for _, alias := range opt.ShortAliases {
			shortIndex[string(alias)] = opt
		}
	}

	for opt := range valueTargets {
		register(opt)
	}

	for opt := range boolTargets {
		register(opt)
	}

	if len(longIndex) == 0 && len(shortIndex) == 0 {
		return
	}

	passDoubleDash := (p.Options & PassDoubleDash) != None

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if passDoubleDash && arg == "--" {
			break
		}

		if !argumentIsOption(arg) {
			continue
		}

		prefix, optname, islong := stripOptionPrefix(arg)
		name, _, value, hasValue := splitOption(prefix, optname, islong)

		var opt *Option
		if islong {
			opt = longIndex[name]
		} else {
			opt = shortIndex[name]
		}

		if opt == nil {
			continue
		}

		if target, ok := boolTargets[opt]; ok {
			*target = true
			continue
		}

		target, ok := valueTargets[opt]
		if !ok {
			continue
		}

		if hasValue {
			*target = value
			continue
		}

		if i+1 < len(args) && !argumentIsOption(args[i+1]) {
			i++
			*target = args[i]
		}
	}
}
