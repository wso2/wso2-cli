package updatechk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc"
	i18n "github.com/wso2/integration-platform-tools/i18n/impl"
	"github.com/wso2/integration-platform-tools/internal/utils"
	"github.com/wso2/integration-platform-tools/internal/utils/textstyle"
)

type UpdateChecker struct {
	metadata *UpdateMetadata
	version  string
}

func (uc *UpdateChecker) Run() {
	if uc.version == "DEV" {
		return
	}

	if uc.metadata == nil {
		return
	}

	if !isCurrentVersionOlder(uc.version, uc.metadata.LatestVersion) {
		return
	}

	if vseg := strings.Split(uc.version, "-"); len(vseg) >= 1 && vseg[0] != uc.metadata.LatestVersion {
		// show update message

		updateCommand := fmt.Sprintf("https://github.com/wso2/integration-platform-tools/releases/tag/%s", uc.metadata.LatestVersion)

		oldTxtStyle := textstyle.CreateTextStyle().SetTextColor(textstyle.BLACK).AddTextStyle(textstyle.BRIGHT)
		newTxtStyle := textstyle.CreateTextStyle().SetTextColor(textstyle.GREEN).AddTextStyle(textstyle.BRIGHT)
		cmdTxtStyle := textstyle.CreateTextStyle().SetTextColor(textstyle.BLUE).AddTextStyle(textstyle.BRIGHT)

		boxView := utils.CreateNotificationBox(utils.Yellow)

		fmt.Fprintln(utils.IO.ErrOut, boxView.Render(heredoc.Docf(i18n.T(`
			Update available: %s -> %s

			Download the latest release:
				%s
		`), oldTxtStyle.Text(uc.version), newTxtStyle.Text(uc.metadata.LatestVersion), cmdTxtStyle.Text(updateCommand))))
	}

}

func isCurrentVersionOlder(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	currentSegs := strings.Split(current, ".")
	latestSegs := strings.Split(latest, ".")

	if len(currentSegs) >= 3 && strings.Contains(currentSegs[2], "-") {
		currentSegs[2] = strings.Split(currentSegs[2], "-")[0]
	}

	if len(latestSegs) >= 3 && strings.Contains(latestSegs[2], "-") {
		latestSegs[2] = strings.Split(latestSegs[2], "-")[0]
	}

	for i := 0; i < 3; i++ {
		if i >= len(latestSegs) {
			break
		}

		currentSeg, _ := strconv.Atoi(currentSegs[i])
		latestSeg, _ := strconv.Atoi(latestSegs[i])

		if currentSeg < latestSeg {
			return true
		}

		if currentSeg > latestSeg {
			return false
		}
	}

	return false
}

func NewUpdateChecker(version string) *UpdateChecker {
	umd, _ := NewUpdateMetadata()

	return &UpdateChecker{
		metadata: umd,
		version:  version,
	}
}
