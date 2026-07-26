// SPDX-FileType: SOURCE
// SPDX-License-Identifier: BSD-3-Clause
// Project: https://github.com/woozymasta/flags

package flags

import "testing"

func TestGroupXorAllowsZeroOrOne(t *testing.T) {
	var opts struct {
		Postgres struct {
			Host string `long:"pg-host"`
			Port string `long:"pg-port"`
		} `group:"Postgres" group-xor:"backend"`
		Sqlite struct {
			Path string `long:"sqlite-path"`
		} `group:"SQLite" group-xor:"backend"`
	}

	assertParseSuccess(t, &opts)
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"pg-host", "localhost")
	assertParseFail(
		t,
		ErrOptionConflict,
		"option groups Postgres and SQLite are mutually exclusive",
		&opts,
		defaultLongOptDelimiter+"pg-host", "localhost",
		defaultLongOptDelimiter+"sqlite-path", "db.sqlite",
	)
}

func TestGroupXorRequiredExactlyOne(t *testing.T) {
	var opts struct {
		Postgres struct {
			Host string `long:"pg-host"`
		} `group:"Postgres" group-xor:"backend" required:"true"`
		Sqlite struct {
			Path string `long:"sqlite-path"`
		} `group:"SQLite" group-xor:"backend"`
	}

	assertParseFail(
		t,
		ErrOptionRequirement,
		"one of option groups Postgres or SQLite must be specified",
		&opts,
	)
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"sqlite-path", "db.sqlite")
	assertParseFail(
		t,
		ErrOptionConflict,
		"option groups Postgres and SQLite are mutually exclusive",
		&opts,
		defaultLongOptDelimiter+"pg-host", "localhost",
		defaultLongOptDelimiter+"sqlite-path", "db.sqlite",
	)
}

func TestGroupAndAllOrNone(t *testing.T) {
	var opts struct {
		Auth struct {
			User string `long:"user"`
		} `group:"Auth" group-and:"creds"`
		Session struct {
			Token string `long:"session-token"`
		} `group:"Session" group-and:"creds"`
	}

	assertParseSuccess(t, &opts)
	assertParseFail(
		t,
		ErrOptionRequirement,
		"option groups Auth and Session must be specified together",
		&opts,
		defaultLongOptDelimiter+"user", "admin",
	)
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"user", "admin",
		defaultLongOptDelimiter+"session-token", "abc",
	)
}

func TestGroupAndRequiredAllMandatory(t *testing.T) {
	var opts struct {
		Auth struct {
			User string `long:"user"`
		} `group:"Auth" group-and:"creds" required:"true"`
		Session struct {
			Token string `long:"session-token"`
		} `group:"Session" group-and:"creds"`
	}

	assertParseFail(
		t,
		ErrOptionRequirement,
		"option groups Auth and Session must be specified together",
		&opts,
	)
	assertParseFail(
		t,
		ErrOptionRequirement,
		"option groups Auth and Session must be specified together",
		&opts,
		defaultLongOptDelimiter+"user", "admin",
	)
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"user", "admin",
		defaultLongOptDelimiter+"session-token", "abc",
	)
}

func TestGroupOrAllowsAnyNonEmptySubset(t *testing.T) {
	var opts struct {
		Email struct {
			Address string `long:"email"`
		} `group:"Email" group-or:"contact"`
		Phone struct {
			Number string `long:"phone"`
		} `group:"Phone" group-or:"contact"`
	}

	assertParseFail(
		t,
		ErrOptionRequirement,
		"one of option groups Email or Phone must be specified",
		&opts,
	)
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"email", "a@b.com")
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"email", "a@b.com",
		defaultLongOptDelimiter+"phone", "123",
	)
}

func TestGroupNandForbidsAllTogether(t *testing.T) {
	var opts struct {
		Fast struct {
			Flag bool `long:"fast-flag"`
		} `group:"Fast" group-nand:"mode"`
		Thorough struct {
			Flag bool `long:"thorough-flag"`
		} `group:"Thorough" group-nand:"mode"`
	}

	assertParseSuccess(t, &opts)
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"fast-flag")
	assertParseFail(
		t,
		ErrOptionConflict,
		"option groups Fast and Thorough cannot all be specified together",
		&opts,
		defaultLongOptDelimiter+"fast-flag",
		defaultLongOptDelimiter+"thorough-flag",
	)
}

func TestGroupRequiresSatisfiedByAnyProvider(t *testing.T) {
	var opts struct {
		TLS struct {
			Cert string `long:"cert"`
		} `group:"TLS" group-requires:"ca"`
		SystemCA struct {
			Path string `long:"system-ca"`
		} `group:"System CA" group-provides:"ca"`
		CustomCA struct {
			Path string `long:"custom-ca"`
		} `group:"Custom CA" group-provides:"ca"`
	}

	assertParseSuccess(t, &opts)
	assertParseFail(
		t,
		ErrOptionRequirement,
		"option group TLS requires one of option groups Custom CA or System CA",
		&opts,
		defaultLongOptDelimiter+"cert", "cert.pem",
	)
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"cert", "cert.pem",
		defaultLongOptDelimiter+"system-ca", "ca.pem",
	)
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"cert", "cert.pem",
		defaultLongOptDelimiter+"custom-ca", "ca.pem",
	)
}

func TestGroupRequiresWithoutProvidesFails(t *testing.T) {
	// A dangling group-requires token is a structural configuration mistake,
	// caught by Parser.Validate (and the first ParseArgs call) regardless of
	// which CLI arguments are passed.
	var opts struct {
		Sub struct {
			TLS struct {
				Cert string `long:"cert"`
			} `group:"TLS" group-requires:"dangling-token"`
		} `command:"sub"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	err := parser.Validate()
	assertError(
		t,
		err,
		ErrInvalidTag,
		"group-requires token `dangling-token` has no matching `group-provides` group in command `sub`",
	)

	assertParseFail(
		t,
		ErrInvalidTag,
		"group-requires token `dangling-token` has no matching `group-provides` group in command `sub`",
		&opts,
	)
}

func TestGroupRequiredWithoutXorOrAndFailsFast(t *testing.T) {
	var opts struct {
		Postgres struct {
			Host string `long:"pg-host"`
		} `group:"Postgres" required:"true"`
	}

	assertParseFail(
		t,
		ErrInvalidTag,
		"group `Postgres` uses `required:\"true\"` without `group-xor` or `group-and`; "+
			"`required` on a group only modifies group-xor/group-and semantics",
		&opts,
	)
}

func TestGroupRequiredFalseWithoutXorOrAndIsNotAnError(t *testing.T) {
	var opts struct {
		Postgres struct {
			Host string `long:"pg-host"`
		} `group:"Postgres" required:"false"`
	}

	assertParseSuccess(t, &opts)
}

func TestGroupRelationsAreCommandLocal(t *testing.T) {
	var opts struct {
		RootA struct {
			Flag string `long:"root-a"`
		} `group:"Root A" group-xor:"backend"`
		RootB struct {
			Flag string `long:"root-b"`
		} `group:"Root B" group-xor:"backend"`
		Cmd struct {
			SubA struct {
				Flag string `long:"sub-a"`
			} `group:"Sub A" group-xor:"backend"`
			SubB struct {
				Flag string `long:"sub-b"`
			} `group:"Sub B" group-xor:"backend"`
		} `command:"login"`
	}

	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"root-a", "1",
		"login",
		defaultLongOptDelimiter+"sub-a", "2",
	)
	assertParseFail(
		t,
		ErrOptionConflict,
		"option groups Sub A and Sub B are mutually exclusive",
		&opts,
		"login",
		defaultLongOptDelimiter+"sub-a", "2",
		defaultLongOptDelimiter+"sub-b", "3",
	)
}

func TestGroupRelationTokensAreIndependentOfOptionTokens(t *testing.T) {
	var opts struct {
		TokenA string `long:"token-a" xor:"shared"`
		TokenB string `long:"token-b" xor:"shared"`
		GroupA struct {
			Flag string `long:"group-a-flag"`
		} `group:"Group A" group-xor:"shared"`
		GroupB struct {
			Flag string `long:"group-b-flag"`
		} `group:"Group B" group-xor:"shared"`
	}

	// Both an option-level xor pair and a group-level group-xor pair share
	// the literal token "shared". They live in separate maps and must not
	// interact: setting one member from each pair simultaneously conflicts
	// with neither relation.
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"token-a", "1",
		defaultLongOptDelimiter+"group-a-flag", "x",
	)
	assertParseFail(
		t,
		ErrOptionConflict,
		"flags "+defaultLongOptDelimiter+"token-a and "+defaultLongOptDelimiter+"token-b are mutually exclusive",
		&opts,
		defaultLongOptDelimiter+"token-a", "1",
		defaultLongOptDelimiter+"token-b", "2",
	)
	assertParseFail(
		t,
		ErrOptionConflict,
		"option groups Group A and Group B are mutually exclusive",
		&opts,
		defaultLongOptDelimiter+"group-a-flag", "x",
		defaultLongOptDelimiter+"group-b-flag", "y",
	)
}

func TestGroupRelationAncestorDescendantTokenSharingLimitation(t *testing.T) {
	// Known, documented limitation: applying the same group-relation token to
	// a group and its own nested subgroup is nonsensical, since groupSatisfied
	// is recursive - the descendant's state leaks into the ancestor's. This
	// pins the current, unguarded behavior rather than letting it change
	// silently later. Do not do this in real configurations.
	var opts struct {
		Outer struct {
			Inner struct {
				Flag string `long:"inner-flag"`
			} `group:"Inner" group-xor:"backend"`
		} `group:"Outer" group-xor:"backend"`
		Other struct {
			Flag string `long:"other-flag"`
		} `group:"Other" group-xor:"backend"`
	}

	assertParseFail(
		t,
		ErrOptionConflict,
		"option groups Inner, Other and Outer are mutually exclusive",
		&opts,
		defaultLongOptDelimiter+"inner-flag", "x",
		defaultLongOptDelimiter+"other-flag", "y",
	)
}
