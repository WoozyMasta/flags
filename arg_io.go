// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"fmt"
	"strings"
)

type argIOConfig struct {
	role   string
	kind   string
	stream string
	open   string
}

const (
	argIORoleIn  = "in"
	argIORoleOut = "out"

	argIOKindAuto   = "auto"
	argIOKindStream = "stream"
	argIOKindFile   = "file"
	argIOKindString = "string"

	argIOStreamStdin  = "stdin"
	argIOStreamStdout = "stdout"
	argIOStreamStderr = "stderr"

	argIOOpenTruncate = "truncate"
	argIOOpenAppend   = "append"
)

func parseFieldIOConfig(tag multiTag, fieldName string, valueType string, target string) (argIOConfig, error) {
	cfg := argIOConfig{
		role:   strings.TrimSpace(tag.Get(FlagTagIO)),
		kind:   strings.TrimSpace(tag.Get(FlagTagIOKind)),
		stream: strings.TrimSpace(tag.Get(FlagTagIOStream)),
		open:   strings.TrimSpace(tag.Get(FlagTagIOOpen)),
	}

	if cfg.role == "" && cfg.kind == "" && cfg.stream == "" && cfg.open == "" {
		return argIOConfig{}, nil
	}
	if cfg.role == "" {
		return argIOConfig{}, newErrorf(
			ErrInvalidTag,
			"field `%s' uses `%s' tags and must define `%s'",
			fieldName,
			"io-*",
			FlagTagIO,
		)
	}
	if valueType != "string" {
		return argIOConfig{}, newErrorf(
			ErrInvalidTag,
			"field `%s' with tag `%s' must be a string %s",
			fieldName,
			FlagTagIO,
			target,
		)
	}

	switch cfg.role {
	case argIORoleIn, argIORoleOut:
	default:
		return argIOConfig{}, newErrorf(
			ErrInvalidTag,
			"invalid value `%s' for tag `%s' on field `%s' (expected in or out)",
			cfg.role,
			FlagTagIO,
			fieldName,
		)
	}

	if cfg.kind == "" {
		cfg.kind = argIOKindAuto
	}
	switch cfg.kind {
	case argIOKindAuto, argIOKindStream, argIOKindFile, argIOKindString:
	default:
		return argIOConfig{}, newErrorf(
			ErrInvalidTag,
			"invalid value `%s' for tag `%s' on field `%s' (expected auto, stream, file, or string)",
			cfg.kind,
			FlagTagIOKind,
			fieldName,
		)
	}

	if cfg.open != "" {
		if cfg.role != argIORoleOut {
			return argIOConfig{}, newErrorf(
				ErrInvalidTag,
				"tag `%s' on field `%s' requires `%s:\"out\"`",
				FlagTagIOOpen,
				fieldName,
				FlagTagIO,
			)
		}
		switch cfg.open {
		case argIOOpenTruncate, argIOOpenAppend:
		default:
			return argIOConfig{}, newErrorf(
				ErrInvalidTag,
				"invalid value `%s' for tag `%s' on field `%s' (expected truncate or append)",
				cfg.open,
				FlagTagIOOpen,
				fieldName,
			)
		}
	}

	if cfg.stream == "" {
		cfg.stream = defaultStreamForRole(cfg.role)
	} else {
		switch cfg.stream {
		case argIOStreamStdin, argIOStreamStdout, argIOStreamStderr:
		default:
			return argIOConfig{}, newErrorf(
				ErrInvalidTag,
				"invalid value `%s' for tag `%s' on field `%s' (expected stdin, stdout, or stderr)",
				cfg.stream,
				FlagTagIOStream,
				fieldName,
			)
		}
		if cfg.role == argIORoleIn && cfg.stream != argIOStreamStdin {
			return argIOConfig{}, newErrorf(
				ErrInvalidTag,
				"tag `%s' on field `%s' with `%s:\"in\"` supports only stdin",
				FlagTagIOStream,
				fieldName,
				FlagTagIO,
			)
		}
		if cfg.role == argIORoleOut && cfg.stream == argIOStreamStdin {
			return argIOConfig{}, newErrorf(
				ErrInvalidTag,
				"tag `%s' on field `%s' with `%s:\"out\"` supports only stdout or stderr",
				FlagTagIOStream,
				fieldName,
				FlagTagIO,
			)
		}
	}

	return cfg, nil
}

func defaultStreamForRole(role string) string {
	if role == argIORoleOut {
		return argIOStreamStdout
	}

	return argIOStreamStdin
}

func isStreamToken(value string) bool {
	switch value {
	case "-", argIOStreamStdin, argIOStreamStdout, argIOStreamStderr:
		return true

	default:
		return false
	}
}

func normalizeIOValue(cfg argIOConfig, raw string) (string, error) {
	switch cfg.kind {
	case argIOKindString:
		return raw, nil

	case argIOKindFile:
		if isStreamToken(raw) {
			return "", fmt.Errorf("io-kind `%s` does not allow stream token `%s`", cfg.kind, raw)
		}
		return raw, nil

	case argIOKindStream:
		return normalizeStreamValue(cfg, raw)

	case argIOKindAuto:
		if isStreamToken(raw) {
			return normalizeStreamValue(cfg, raw)
		}
		return raw, nil

	default:
		return "", ErrInvalidTag
	}
}

func (a *Arg) normalizeIOValue(raw string) (string, error) {
	return normalizeIOValue(a.io, raw)
}

func normalizeStreamValue(cfg argIOConfig, raw string) (string, error) {
	if raw == "-" {
		return cfg.stream, nil
	}

	switch cfg.role {
	case argIORoleIn:
		if raw != argIOStreamStdin {
			return "", fmt.Errorf("io `%s` accepts only `%s` or `-`", cfg.role, argIOStreamStdin)
		}

	case argIORoleOut:
		if raw != argIOStreamStdout && raw != argIOStreamStderr {
			return "", fmt.Errorf("io `%s` accepts only `%s`, `%s`, or `-`", cfg.role, argIOStreamStdout, argIOStreamStderr)
		}
	}

	return raw, nil
}

func (a *Arg) applyIOFallback() (bool, error) {
	cfg := a.io
	if cfg.role == "" {
		return false, nil
	}

	switch cfg.kind {
	case argIOKindStream, argIOKindAuto:
		stream := cfg.stream
		if stream == "" {
			stream = defaultStreamForRole(cfg.role)
		}
		if err := convert(stream, a.value, a.tag); err != nil {
			return false, err
		}
		return true, nil

	default:
		return false, nil
	}
}

func completionHintFromIO(cfg argIOConfig) (completionHint, bool) {
	if cfg.role == "" {
		return completionHintAuto, false
	}

	switch cfg.kind {
	case argIOKindFile, argIOKindAuto:
		return completionHintFile, true
	default:
		return completionHintAuto, false
	}
}
