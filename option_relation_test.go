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
