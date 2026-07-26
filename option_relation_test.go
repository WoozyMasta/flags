package flags

import "testing"

func TestOptionXorAllowsZeroOrOne(t *testing.T) {
	var opts struct {
		Token string `long:"token" xor:"auth"`
		User  string `long:"user" xor:"auth"`
	}

	assertParseSuccess(t, &opts)
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"token", "secret")

	assertParseFail(
		t,
		ErrOptionConflict,
		"flags "+defaultLongOptDelimiter+"token and "+defaultLongOptDelimiter+"user are mutually exclusive",
		&opts,
		defaultLongOptDelimiter+"token", "secret",
		defaultLongOptDelimiter+"user", "admin",
	)
}

func TestOptionXorRequiredRequiresExactlyOne(t *testing.T) {
	var opts struct {
		Token string `long:"token" xor:"auth" required:"true"`
		User  string `long:"user" xor:"auth"`
	}

	assertParseFail(
		t,
		ErrOptionRequirement,
		"one of flags "+defaultLongOptDelimiter+"token or "+defaultLongOptDelimiter+"user must be specified",
		&opts,
	)
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"user", "admin")
	assertParseFail(
		t,
		ErrOptionConflict,
		"flags "+defaultLongOptDelimiter+"token and "+defaultLongOptDelimiter+"user are mutually exclusive",
		&opts,
		defaultLongOptDelimiter+"token", "secret",
		defaultLongOptDelimiter+"user", "admin",
	)
}

func TestOptionAndRequiresAllWhenAnyIsSet(t *testing.T) {
	var opts struct {
		User string `long:"user" and:"basic-auth"`
		Pass string `long:"pass" and:"basic-auth"`
	}

	assertParseSuccess(t, &opts)
	assertParseFail(
		t,
		ErrOptionRequirement,
		"flags "+defaultLongOptDelimiter+"pass and "+defaultLongOptDelimiter+"user must be specified together",
		&opts,
		defaultLongOptDelimiter+"user", "admin",
	)
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"user", "admin",
		defaultLongOptDelimiter+"pass", "secret",
	)
}

func TestOptionAndRequiredRequiresAll(t *testing.T) {
	var opts struct {
		User string `long:"user" and:"basic-auth" required:"true"`
		Pass string `long:"pass" and:"basic-auth"`
	}

	assertParseFail(
		t,
		ErrOptionRequirement,
		"flags "+defaultLongOptDelimiter+"pass and "+defaultLongOptDelimiter+"user must be specified together",
		&opts,
	)
	assertParseFail(
		t,
		ErrOptionRequirement,
		"flags "+defaultLongOptDelimiter+"pass and "+defaultLongOptDelimiter+"user must be specified together",
		&opts,
		defaultLongOptDelimiter+"user", "admin",
	)
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"user", "admin",
		defaultLongOptDelimiter+"pass", "secret",
	)
}

func TestOptionRelationsUseTagListDelimiter(t *testing.T) {
	var opts struct {
		Token string `long:"token" xor:"auth,login"`
		User  string `long:"user" xor:"auth"`
		Pass  string `long:"pass" xor:"login"`
	}

	p := NewParser(&opts, Default&^PrintErrors)
	if err := p.SetTagListDelimiter(','); err != nil {
		t.Fatal(err)
	}

	_, err := p.ParseArgs([]string{
		defaultLongOptDelimiter + "token", "secret",
		defaultLongOptDelimiter + "pass", "secret",
	})
	assertError(
		t,
		err,
		ErrOptionConflict,
		"flags "+defaultLongOptDelimiter+"pass and "+defaultLongOptDelimiter+"token are mutually exclusive",
	)
}

func TestOptionRelationsAreCommandLocal(t *testing.T) {
	var opts struct {
		RootToken string `long:"root-token" xor:"auth"`
		Cmd       struct {
			User string `long:"user" xor:"auth"`
			Pass string `long:"pass" xor:"auth"`
		} `command:"login"`
	}

	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"root-token", "secret",
		"login",
		defaultLongOptDelimiter+"user", "admin",
	)
	assertParseFail(
		t,
		ErrOptionConflict,
		"flags "+defaultLongOptDelimiter+"pass and "+defaultLongOptDelimiter+"user are mutually exclusive",
		&opts,
		"login",
		defaultLongOptDelimiter+"user", "admin",
		defaultLongOptDelimiter+"pass", "secret",
	)
}

func TestOptionAndMultipleGroups(t *testing.T) {
	var opts struct {
		User  string `long:"user" and:"auth"`
		Pass  string `long:"pass" and:"auth"`
		Host  string `long:"host" and:"db"`
		Port  string `long:"port" and:"db"`
		Token string `long:"token"`
	}

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errType ErrorType
		errMsg  string
	}{
		{
			name: "no and groups set",
			args: nil,
		},
		{
			name: "auth complete only",
			args: []string{
				defaultLongOptDelimiter + "user", "admin",
				defaultLongOptDelimiter + "pass", "secret",
			},
		},
		{
			name: "db complete only",
			args: []string{
				defaultLongOptDelimiter + "host", "localhost",
				defaultLongOptDelimiter + "port", "5432",
			},
		},
		{
			name: "both groups complete",
			args: []string{
				defaultLongOptDelimiter + "user", "admin",
				defaultLongOptDelimiter + "pass", "secret",
				defaultLongOptDelimiter + "host", "localhost",
				defaultLongOptDelimiter + "port", "5432",
			},
		},
		{
			name: "auth partial fails",
			args: []string{
				defaultLongOptDelimiter + "user", "admin",
			},
			wantErr: true,
			errType: ErrOptionRequirement,
			errMsg:  "flags " + defaultLongOptDelimiter + "pass and " + defaultLongOptDelimiter + "user must be specified together",
		},
		{
			name: "db partial fails",
			args: []string{
				defaultLongOptDelimiter + "host", "localhost",
			},
			wantErr: true,
			errType: ErrOptionRequirement,
			errMsg:  "flags " + defaultLongOptDelimiter + "host and " + defaultLongOptDelimiter + "port must be specified together",
		},
		{
			name: "independent groups still independent with unrelated flag",
			args: []string{
				defaultLongOptDelimiter + "token", "abc",
				defaultLongOptDelimiter + "user", "admin",
			},
			wantErr: true,
			errType: ErrOptionRequirement,
			errMsg:  "flags " + defaultLongOptDelimiter + "pass and " + defaultLongOptDelimiter + "user must be specified together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assertParseFail(t, tt.errType, tt.errMsg, &opts, tt.args...)
				return
			}

			assertParseSuccess(t, &opts, tt.args...)
		})
	}
}

func TestOptionAndGroupNamesAreNotOptionReferences(t *testing.T) {
	var opts struct {
		Alpha string `long:"alpha" and:"beta"`
		Beta  string `long:"beta" and:"gamma"`
		Gamma string `long:"gamma" and:"alpha"`
	}

	// Each option forms its own single-member and-group:
	// group "b" => {--a}, group "c" => {--b}, group "a" => {--c}.
	// This is valid and does not create transitive A-B-C relation.
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"alpha", "1",
		defaultLongOptDelimiter+"beta", "2",
	)
}

func TestOptionOrAllowsAnyNonEmptySubset(t *testing.T) {
	var opts struct {
		Host string `long:"host" or:"target"`
		Port string `long:"port" or:"target"`
	}

	assertParseFail(
		t,
		ErrOptionRequirement,
		"one of flags "+defaultLongOptDelimiter+"host or "+defaultLongOptDelimiter+"port must be specified",
		&opts,
	)
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"host", "example.com")
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"port", "8080")
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"host", "example.com",
		defaultLongOptDelimiter+"port", "8080",
	)
}

func TestOptionNandForbidsAllSetTogether(t *testing.T) {
	var opts struct {
		Fast     bool `long:"fast" nand:"mode"`
		Thorough bool `long:"thorough" nand:"mode"`
	}

	assertParseSuccess(t, &opts)
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"fast")
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"thorough")
	assertParseFail(
		t,
		ErrOptionConflict,
		"flags "+defaultLongOptDelimiter+"fast and "+defaultLongOptDelimiter+"thorough cannot all be specified together",
		&opts,
		defaultLongOptDelimiter+"fast",
		defaultLongOptDelimiter+"thorough",
	)
}

func TestOptionNandSingleMemberIsNoOp(t *testing.T) {
	// A single-member nand group can never have "all members set", so it is
	// a permanent no-op, the same way a single-member xor group is harmless.
	var opts struct {
		Fast bool `long:"fast" nand:"solo"`
	}

	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"fast")
}

func TestOptionRequiresSatisfiedByAnyProvider(t *testing.T) {
	var opts struct {
		Cert string `long:"cert" requires:"tls-key"`
		Key  string `long:"key" provides:"tls-key"`
	}

	assertParseSuccess(t, &opts)
	assertParseFail(
		t,
		ErrOptionRequirement,
		"flag "+defaultLongOptDelimiter+"cert requires one of flags "+defaultLongOptDelimiter+"key",
		&opts,
		defaultLongOptDelimiter+"cert", "cert.pem",
	)
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"cert", "cert.pem",
		defaultLongOptDelimiter+"key", "key.pem",
	)
}

func TestOptionRequiresMultipleProviders(t *testing.T) {
	// requires/provides matching is many-to-many: any live provider satisfies
	// any live requirer sharing the same token.
	var opts struct {
		Primary   string `long:"primary" requires:"backend"`
		Secondary string `long:"secondary" requires:"backend"`
		Redis     string `long:"redis" provides:"backend"`
		Memcached string `long:"memcached" provides:"backend"`
	}

	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"primary", "a", defaultLongOptDelimiter+"redis", "r")
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"secondary", "b", defaultLongOptDelimiter+"memcached", "m")
	assertParseFail(
		t,
		ErrOptionRequirement,
		"flag "+defaultLongOptDelimiter+"primary and "+defaultLongOptDelimiter+"secondary"+
			" requires one of flags "+defaultLongOptDelimiter+"memcached or "+defaultLongOptDelimiter+"redis",
		&opts,
		defaultLongOptDelimiter+"primary", "a",
		defaultLongOptDelimiter+"secondary", "b",
	)
}

func TestOptionProvidesWithoutRequiresIsNotAnError(t *testing.T) {
	// An unused provider is harmless, the same way a size-1 xor group is.
	var opts struct {
		Key string `long:"key" provides:"tls-key"`
	}

	assertParseSuccess(t, &opts)
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"key", "key.pem")
}

func TestOptionSelfReferentialRequiresProvides(t *testing.T) {
	// An option that both requires and provides the same token trivially
	// satisfies itself once set; this is valid, not a foot-gun to guard
	// against.
	var opts struct {
		Flag string `long:"flag" requires:"self" provides:"self"`
	}

	assertParseSuccess(t, &opts)
	assertParseSuccess(t, &opts, defaultLongOptDelimiter+"flag", "value")
}

func TestOptionRequiresWithoutProvidesFailsAtValidate(t *testing.T) {
	// A dangling requires token is a structural configuration mistake, so it
	// is caught by Parser.Validate (and, transitively, by the very first
	// ParseArgs call) regardless of which CLI arguments are passed - not
	// deferred until someone happens to set the requiring flag.
	var opts struct {
		Sub struct {
			Cert string `long:"cert" requires:"dangling-token"`
		} `command:"sub"`
	}

	parser := NewParser(&opts, Default&^PrintErrors)
	err := parser.Validate()
	assertError(
		t,
		err,
		ErrInvalidTag,
		"requires token `dangling-token` has no matching `provides` option in command `sub`",
	)

	assertParseFail(
		t,
		ErrInvalidTag,
		"requires token `dangling-token` has no matching `provides` option in command `sub`",
		&opts,
	)
}

func TestOptionRequiredIndependentOfOrNand(t *testing.T) {
	// required keeps its normal, independent, unconditional meaning when
	// combined with or/nand/requires/provides - unlike its modifier role for
	// xor/and. A required member of an `or` group is unconditionally
	// mandatory in addition to (not instead of) the group's own "at least
	// one" rule, which the tag itself already enforces unconditionally.
	var opts struct {
		Host string `long:"host" or:"target" required:"true"`
		Port string `long:"port" or:"target"`
	}

	assertParseFail(
		t,
		ErrRequired,
		"the required flag `"+defaultLongOptDelimiter+"host` was not specified",
		&opts,
		defaultLongOptDelimiter+"port", "8080",
	)
	assertParseSuccess(
		t,
		&opts,
		defaultLongOptDelimiter+"host", "example.com",
		defaultLongOptDelimiter+"port", "8080",
	)
}
