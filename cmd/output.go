package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
)

// jsonOut is the root's --json flag: every command with something to say
// says it as JSON on stdout instead of drawing, and an error becomes one
// JSON object on stderr, so a script can branch on it.
var jsonOut bool

// stdout is where printJSON writes; tests point it at a buffer.
var stdout io.Writer = os.Stdout

// printJSON writes v to stdout as indented JSON.
func printJSON(v any) error {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// The error codes a script can switch on. They are stable; the messages
// are not.
const (
	codeUnknownMonster   = "unknown-monster"
	codeAmbiguousMonster = "ambiguous-monster"
	codeBadPack          = "bad-pack"
	codeBadDate          = "bad-date"
	codeBadFlag          = "bad-flag"
	codeNoTerminal       = "no-terminal"
	codeStore            = "store"
	codeUsage            = "usage" // anything else: a bad argument, an empty pack
)

// codedError is an error with a stable code and, sometimes, a hint on
// what to do next.
type codedError struct {
	Code string
	Msg  string
	Hint string
}

func (e *codedError) Error() string {
	if e.Hint != "" {
		return e.Msg + "; " + e.Hint
	}
	return e.Msg
}

// fail builds a codedError.
func fail(code, hint, format string, args ...any) error {
	return &codedError{Code: code, Msg: fmt.Sprintf(format, args...), Hint: hint}
}

// jsonError is the shape of an error under --json.
type jsonError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

// errorJSON turns any error into the object --json reports: its code when
// it has one, "usage" otherwise.
func errorJSON(err error) jsonError {
	var ce *codedError
	if errors.As(err, &ce) {
		return jsonError{Code: ce.Code, Message: ce.Msg, Hint: ce.Hint}
	}
	return jsonError{Code: codeUsage, Message: err.Error()}
}

// writeErrorJSON writes the error object to w as one line.
func writeErrorJSON(w io.Writer, err error) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(map[string]jsonError{"error": errorJSON(err)})
}

// wantsJSON reads --json from the raw arguments, for the errors cobra
// reports before it has parsed any flag (an unknown command, a bad flag).
func wantsJSON(args []string) bool {
	return jsonOut || slices.Contains(args, "--json")
}
