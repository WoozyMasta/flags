// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func (p *Parser) showBuiltinHelp() error {
	var b bytes.Buffer

	p.WriteHelp(&b)
	return newError(ErrHelp, b.String())
}

func (p *Parser) showBuiltinVersion() error {
	var b bytes.Buffer

	p.WriteVersion(&b, p.versionFields)
	return newError(ErrVersion, b.String())
}

func (p *Parser) markVersionRequested() error {
	p.versionRequested = true
	return nil
}

func (p *Parser) printError(err error) error {
	if err != nil && (p.Options&PrintErrors) != None {
		writer := p.errorWriter(err)
		flagsErr, ok := err.(*Error)

		if ok && flagsErr.Type == ErrHelp {
			p.WriteHelp(writer)
			return err
		}

		if p.shouldPrintHelpOnError(err) {
			p.WriteHelp(writer)
		}

		_, _ = fmt.Fprintln(writer, p.colorizeError(err, err.Error(), writer))
	}

	return err
}

func (p *Parser) shouldPrintHelpOnError(err error) bool {
	if (p.Options & PrintHelpOnInputErrors) == None {
		return false
	}

	flagsErr, ok := err.(*Error)
	if !ok {
		return false
	}

	if flagsErr.Type == ErrHelp || flagsErr.Type == ErrVersion {
		return false
	}

	return shouldPrintHelpForErrorType(flagsErr.Type)
}

func shouldPrintHelpForErrorType(errorType ErrorType) bool {
	switch errorType {
	case ErrRequired,
		ErrCommandRequired,
		ErrUnknownFlag,
		ErrUnknownCommand,
		ErrExpectedArgument,
		ErrInvalidChoice,
		ErrNoArgumentForBool:
		return true
	default:
		return false
	}
}

func (p *Parser) errorWriter(err error) io.Writer {
	flagsErr, ok := err.(*Error)

	if ok && (flagsErr.Type == ErrHelp || flagsErr.Type == ErrVersion) {
		if (p.Options & PrintHelpOnStderr) != None {
			return os.Stderr
		}

		return os.Stdout
	}

	if (p.Options & PrintErrorsOnStdout) != None {
		return os.Stdout
	}

	return os.Stderr
}
