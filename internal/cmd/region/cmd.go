package region

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/internal/utils"
)

var RegionCommand = &cobra.Command{
	Use:   "region [region-name]",
	Short: i18n.T("Set or view the region"),
	Long: heredoc.Doc(i18n.T(`
		Set or view the region.
		
		If no region is specified, displays the current region.
		If a region is specified, sets the current region to the specified value.
		
		Valid regions are:
		- US (United States)
		- EU (European Union)
		
		Note: Changing regions may require you to log in again if you don't have
		a valid session for the new region.
	`)),
	Example: heredoc.Docf(i18n.T(`
		# View current region
		$ wso2-integration-platform region

		# Switch to EU region
		$ wso2-integration-platform region EU

		# Switch to US region
		$ wso2-integration-platform region US
	`)),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Show current region
			region := region.GetCurrentRegion()
			fmt.Printf("Current region: %s\n", region)
			return
		}

		newRegion := args[0]
		err := region.SetCurrentRegion(newRegion)
		if err != nil {
			utils.HandleErr(err)
			return
		}

		fmt.Printf("Successfully switched to %s region\n", newRegion)
	},
}
