package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/pflag"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

// Only table and json are supported for now
// but we could add more formats in the future. (eg: YAML)
type OutputFormat string

const (
	OutputFormatTable OutputFormat = "table"
	OutputFormatJSON  OutputFormat = "json"
)

const (
	OutputFlagName = "output"
	OutputEnvVar   = "OUTPUT_FORMAT"
)

var (
	defaultOutputFormatOnce sync.Once
	defaultOutputFormat     OutputFormat
)

// ResolveDefaultOutputFormat returns the default output format: OUTPUT_FORMAT
// if set and valid, otherwise table. Resolved once per process.
func ResolveDefaultOutputFormat() OutputFormat {
	defaultOutputFormatOnce.Do(func() {
		raw := os.Getenv(OutputEnvVar)
		if f, err := ParseOutputFormat(raw); err == nil {
			defaultOutputFormat = f
			return
		}
		defaultOutputFormat = OutputFormatTable
	})
	return defaultOutputFormat
}

// globalOutputFormat classifies the OUTPUT_FORMAT env var: the parsed format,
// whether it is set, and whether it is valid. Read directly (not via the flag
// default) since that has already coerced an invalid global to table.
func globalOutputFormat() (format OutputFormat, set, valid bool) {
	raw := os.Getenv(OutputEnvVar)
	if raw == "" {
		return "", false, false
	}
	parsed, err := ParseOutputFormat(raw)
	if err != nil {
		return "", true, false
	}
	return parsed, true, true
}

// AddOutputFlag registers the --output / -o flag on a command's flag set.
// The default value honors OUTPUT_FORMAT so users can set a global preference;
// passing -o on the command line still overrides it.
func AddOutputFlag(cmdFlags *pflag.FlagSet, bindTo *string) {
	cmdFlags.StringVarP(bindTo, OutputFlagName, "o", string(ResolveDefaultOutputFormat()),
		i18n.T("output format: table or json (defaults to $OUTPUT_FORMAT if set)"))
}

// ResolveOutputFormat resolves the format a command should use from its raw
// --output value, honoring OUTPUT_FORMAT as a fallback. It is the entry point
// commands should call. An invalid value warns and falls back where a usable
// format exists, otherwise errors.
func ResolveOutputFormat(raw string) (OutputFormat, error) {
	global, globalSet, globalValid := globalOutputFormat()

	// Resolve inline first so the invalid-global warning can name the format
	// actually applied.
	inline, inlineErr := ParseOutputFormat(raw)

	effective, err := inline, error(nil)
	switch {
	case inlineErr == nil:
		// Valid inline value (or none, which defaults) — use it.
	case globalValid:
		// Invalid inline, valid global — fall back to the global preference.
		effective = global
		fmt.Fprintf(utils.IO.ErrOut,
			i18n.T("%s Ignoring --%s=%q (valid formats: table, json); using your %s=%q instead.\n"),
			utils.CS.Yellow("!"), OutputFlagName, raw, OutputEnvVar, global)
	case globalSet:
		// Both invalid — keep the error generic rather than echoing two values.
		err = errors.New(i18n.T("invalid output format: must be one of table, json"))
	default:
		// Only inline invalid, no global — name the value the user gave.
		err = inlineErr
	}

	// Invalid global the command could still recover from: warn, naming the
	// format used in its place.
	if globalSet && !globalValid && err == nil {
		fmt.Fprintf(utils.IO.ErrOut,
			i18n.T("%s Ignoring %s=%q (valid formats: table, json); showing output as %q instead.\n"),
			utils.CS.Yellow("!"), OutputEnvVar, os.Getenv(OutputEnvVar), effective)
	}

	return effective, err
}

// Validates a raw flag value and converts it to an OutputFormat.
// The value is normalized (trimmed and lower-cased) so accidental casing or
// surrounding whitespace — e.g. "JSON" or " json" — is still accepted.
func ParseOutputFormat(raw string) (OutputFormat, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch OutputFormat(normalized) {
	case "", OutputFormatTable:
		return OutputFormatTable, nil
	case OutputFormatJSON:
		return OutputFormatJSON, nil
	default:
		return "", fmt.Errorf(
			i18n.T("invalid output format %q: must be one of table, json"), raw)
	}
}

// Check whether the format is a machine-readable one (json) or not (table),
// Commands use this to decide whether to suppress decorative output
func (f OutputFormat) IsStructured() bool {
	return f == OutputFormatJSON
}

// Turn the data into a JSON string
// MarshalIndent is used to make the output more human-readable
func RenderStructured(format OutputFormat, data any) (string, error) {
	switch format {
	case OutputFormatJSON:
		bytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf(i18n.T("failed to render JSON output: %w"), err)
		}
		return string(bytes), nil
	default:
		return "", fmt.Errorf(i18n.T("cannot render structured output for format %q"), format)
	}
}
