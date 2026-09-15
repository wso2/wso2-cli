package env

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/cmd/common"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

// Single source of truth for environment variables the CLI recognizes.
// New global env vars should be appended here so `wso2-integration-platform env` stays
// discoverable without per-feature wiring.
var recognizedVars = []envVar{
	{
		Name:        "OUTPUT_FORMAT",
		Description: "Default value for --output on commands that support it (table, json).",
		Default:     "table",
		Sensitive:   false,
	},
	{
		Name:        "WSO2IP_PAT",
		Description: "Personal Access Token for non-interactive authentication.",
		Default:     "",
		Sensitive:   true,
	},
	{
		Name:        "WSO2IP_REGION",
		Description: "Selects the deployment region (US or EU).",
		Default:     "",
		Sensitive:   false,
	},
}

type envVar struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     string `json:"default"`
	Value       string `json:"value"`
	Set         bool   `json:"set"`
	Sensitive   bool   `json:"-"`
}

type options struct {
	outputFlag string
}

var params options

var EnvCommand = &cobra.Command{
	Use:   "env [flags]",
	Short: i18n.T("list environment variables the CLI recognizes"),
	Long: heredoc.Doc(i18n.T(`
		List every environment variable the CLI reads, along with its
		current value (if set in the calling shell), its built-in default, and
		a short description of what it controls.

		Sensitive values (such as WSO2IP_PAT) are masked; only whether they are
		set is reported.
	`)),
	Example: heredoc.Docf(i18n.T(`

		To list recognized environment variables:
			%s

		To get the output in JSON format:
			%s
	`),
		"$ wso2-integration-platform env",
		"$ wso2-integration-platform env --output=json",
	),
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(&params); err != nil {
			utils.HandleErr(err)
		}
	},
}

func init() {
	common.AddOutputFlag(EnvCommand.Flags(), &params.outputFlag)
	common.AddGenericHelper(EnvCommand)
}

func run(opts *options) error {
	format, err := common.ResolveOutputFormat(opts.outputFlag)
	if err != nil {
		return err
	}

	rows := snapshot()

	if format.IsStructured() {
		out, err := common.RenderStructured(format, rows)
		if err != nil {
			return err
		}
		fmt.Fprintln(utils.IO.Out, out)
		return nil
	}

	data := make([][]string, 0, len(rows))
	for _, r := range rows {
		data = append(data, []string{r.Name, displayValue(r), r.Default, r.Description})
	}
	fmt.Fprint(utils.IO.Out, utils.CreateTable(data,
		[]string{"NAME", "VALUE", "DEFAULT", "DESCRIPTION"}, ""))
	return nil
}

func snapshot() []envVar {
	out := make([]envVar, len(recognizedVars))
	for i, v := range recognizedVars {
		raw, ok := os.LookupEnv(v.Name)
		v.Set = ok
		if v.Sensitive {
			v.Value = ""
		} else {
			v.Value = raw
		}
		out[i] = v
	}
	return out
}

func displayValue(v envVar) string {
	if !v.Set {
		return "(unset)"
	}
	if v.Sensitive {
		return "(set)"
	}
	if v.Value == "" {
		return "(empty)"
	}
	return v.Value
}
