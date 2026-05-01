# I/O Templates

I/O templates describe command-line values whose meaning
is an input source or an output sink.
They are designed for common CLI conventions such as:

* omitted input means standard input;
* omitted output means standard output;
* `-` means the standard stream for the role;
* any other value can be treated as a path or raw string.

The parser stores normalized string values and metadata.
It does not open files or close handles.

## Tags

* `io` declares the role.
  Use `io:"in"` for input values and `io:"out"` for output values.
* `io-kind` declares accepted value kind.
  Supported values are `auto`, `stream`, `file`, and `string`.
  The default behavior is equivalent to `auto` for I/O templates.
* `io-stream` declares the stream token used when `-`
  or an omitted positional value maps to a stream.  
  For input, the practical stream is `stdin`.  
  For output, common streams are `stdout` and `stderr`.
* `io-open` stores output open-mode metadata.
  Supported values are `truncate` and `append`.
  It is valid only with `io:"out"`.

## Kinds

* `auto` accepts stream tokens and file-like values.
  The `-` token is normalized to the configured stream.
  Other strings are kept as provided.
* `stream` accepts only stream tokens valid for the role.  
  For input, this means `stdin` or `-`.  
  For output, this means `stdout`, `stderr`, or `-`.
* `file` rejects stream tokens.
  Use it when the application must receive a path.
* `string` keeps raw values.
  Use it when the same user-facing field may contain arbitrary text,
  not only a path or stream token.

## Positional Behavior

For positional arguments, I/O templates can provide omitted-value fallbacks.

```go
type Options struct {
  IO struct {
    Input  string `io:"in" io-kind:"auto"`
    Output string `io:"out" io-kind:"auto"`
  } `positional-args:"yes"`
}
```

Behavior:

* no args stores `Input="stdin"` and `Output="stdout"`;
* `- -` stores `Input="stdin"` and `Output="stdout"`;
* `input.txt output.txt` stores both file paths.

This supports common filter-style commands:

```bash
app                  # stdin to stdout
app input.txt        # file to stdout
app input.txt out.txt
app - out.txt        # stdin to file
```

## Option Behavior

For options, I/O templates normalize only provided values.
They do not set a fallback when the flag is omitted.

```go
type Options struct {
  Input  string `long:"input" io:"in" io-kind:"auto"`
  Output string `long:"output" io:"out" io-kind:"auto" io-stream:"stderr"`
}
```

If `--input -` is passed, `Input` becomes `stdin`.  
If `--input` is omitted,
`Input` stays the zero value unless another default source sets it.

This difference is intentional.
A positional input can represent the command's primary stream contract.
An omitted option usually means the option was not requested.

## Opening Values in Application Code

The parser intentionally does not open files.
This keeps parsing side-effect-light and lets applications decide permissions,
file modes, buffering, and lifetime.

```go
type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }

func openInput(value string) (io.ReadCloser, error) {
  if value == "stdin" {
    return io.NopCloser(os.Stdin), nil
  }
  return os.Open(value)
}

func openOutput(value string) (io.WriteCloser, error) {
  switch value {
  case "stdout":
    return nopWriteCloser{os.Stdout}, nil
  case "stderr":
    return nopWriteCloser{os.Stderr}, nil
  default:
    return os.Create(value)
  }
}
```

For append mode, read `io-open` metadata from your own command configuration
or keep the mode explicit in application code.
The tag is metadata, not a file operation.

## Completion

When `completion` is not set explicitly,
`io-kind:"file"` and `io-kind:"auto"` imply file completion.

Set `completion:"none"` to disable that hint.  
Set `completion:"dir"` when a directory is expected instead of a file.

## Validation

I/O templates combine well with validation tags.
For example, use `validate-existing-file` and `validate-readable`
for a required file input.

Do not combine `validate-existing-file` with stream-capable values
unless your application never accepts `stdin` for that field.
A stream token is not a file path.

## Choosing an I/O Kind

Use I/O templates for command-line convention, not for business logic.
They answer what the value means, not how processing should be implemented.

Use `auto` for user-facing filter commands.  
Use `file` when a real path is mandatory.  
Use `stream` for explicit stream selection.  
Use `string` when the value may be arbitrary text.
