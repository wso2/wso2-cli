package commons

import (
	"fmt"

	"github.com/wso2/integration-platform-tools/internal/utils/prompt"
)

func ConfirmDeleteWithName(name string) bool {
	inputText := ""

	err := prompt.NewPromptInputMessage(
		prompt.PromptInputOpts{
			Message: fmt.Sprintf(`Type "%s" to confirm deletion`, name),
			Validate: func(s string) error {
				if s != name {
					return fmt.Errorf("Invalid input, you entered %s", s)
				}
				return nil
			},
		},
		&inputText,
	).Prompt()

	if err != nil {
		return false
	}

	return true
}
