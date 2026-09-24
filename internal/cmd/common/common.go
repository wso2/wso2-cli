package common

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

// LegacyAliasCommand returns a hidden command that runs target under an older
// name. `component`/`components` were renamed to `integration`/`integrations`;
// the old spellings keep working for existing scripts, print a one-line
// deprecation notice on stderr, and stay out of the help output.
func LegacyAliasCommand(target *cobra.Command, use string, aliases ...string) *cobra.Command {
	alias := &cobra.Command{
		Use:     use,
		Aliases: aliases,
		Short:   target.Short,
		Long:    target.Long,
		Example: target.Example,
		Args:    target.Args,
		Hidden:  true,
		Run:     target.Run,
		RunE:    target.RunE,
	}
	alias.PreRun = func(cmd *cobra.Command, args []string) {
		utils.PrintError("Note: \"%s\" is deprecated, use \"%s\" instead.\n", cmd.CalledAs(), target.Name())
		if target.PreRun != nil {
			target.PreRun(cmd, args)
		}
	}
	// Shares the flag pointers with target, so both spellings bind the same values.
	alias.Flags().AddFlagSet(target.Flags())
	return alias
}

func AddGenericHelper(command *cobra.Command) {
	command.SetHelpFunc(func(c *cobra.Command, s []string) {
		if c.HasSubCommands() && len(c.Flags().Args()) > 0 {
			fmt.Fprint(utils.IO.ErrOut, heredoc.Docf(`
				Error: unknown command "%s" for "%s"
				Run '%s --help' for usage.
			`, c.Flags().Args()[0], c.CommandPath(), c.CommandPath()))
		} else {
			fmt.Fprint(utils.IO.Out, heredoc.Docf(`
			%s

			%s
		`, c.Long, c.UsageString()))
		}
	})
}

func GenerateFlags(cmd *cobra.Command, flagKeys []string) string {
	var flags [][]string

	for _, key := range flagKeys {
		flagRow := []string{}
		flagShorthand := cmd.Flag(key).Shorthand
		if flagShorthand != "" {
			flagRow = append(flagRow, fmt.Sprintf(`-%s,`, flagShorthand))
		} else {
			flagRow = append(flagRow, "   ")
		}

		flagType := cmd.Flag(key).Value.Type()
		if flagType == "bool" {
			flagType = ""
		}

		flagRow = append(flagRow, fmt.Sprintf(`--%s`, cmd.Flag(key).Name))
		flagRow = append(flagRow, cmd.Flag(key).Usage)

		flags = append(flags, flagRow)
	}

	tableStr := utils.CreateTable(flags, nil, " ")
	return tableStr
}

func ValidateNameText(componentName string) error {
	if len(componentName) > 60 {
		return fmt.Errorf("name cannot have more than 60 characters")
	}

	if len(componentName) < 3 {
		return fmt.Errorf("name must have at least 3 characters")
	}

	nameRegexStartWith, err := regexp.Compile("^[A-Za-z]")
	if err != nil {
		return err
	}

	if !nameRegexStartWith.MatchString(componentName) {
		return fmt.Errorf("name must start with an alphabetic letter")
	}

	nameRegexSpecialChar, err := regexp.Compile(`^[^*|":%#!=<>[\]{}` + "`" + `\\/.()';@&$]+$`)
	if err != nil {
		return err
	}

	if !nameRegexSpecialChar.MatchString(componentName) {
		return fmt.Errorf("name cannot have any special characters")
	}

	return nil
}

// Parses strings in the format `key=value`, and returns a KeyValOpt struct
func ParseKeyValStrPair(keyValStr string) (KeyValOpt, error) {
	keyVal := KeyValOpt{}
	keyValStrSplit := strings.Split(keyValStr, "=")
	if len(keyValStrSplit) != 2 {
		return keyVal, fmt.Errorf("invalid key-value pair")
	}

	keyVal.Key = keyValStrSplit[0]
	keyVal.Val = keyValStrSplit[1]

	return keyVal, nil
}
