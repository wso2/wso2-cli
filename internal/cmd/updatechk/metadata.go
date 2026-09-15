package updatechk

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wso2/integration-platform-tools/pkg/util/constants"
)

const UPDATE_METADATA_FILE = "last-update-check.json"

type UpdateMetadata struct {
	LatestVersion string    `json:"latest_version"`
	LastChecked   time.Time `json:"last_checked"`
}

func (umd *UpdateMetadata) load() error {
	usrHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fp := filepath.Join(usrHomeDir, constants.CLI_HOME_DIR, UPDATE_METADATA_FILE)

	if _, err := os.Stat(fp); os.IsNotExist(err) {
		umd.LatestVersion, err = getLatestVersion()
		if err != nil {
			return err
		}

		umd.LastChecked = time.Now()

		if err := umd.save(); err != nil {
			return err
		}
	} else {
		data, err := os.ReadFile(fp)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(data, umd); err != nil {
			return err
		}

		if time.Since(umd.LastChecked).Hours() > 24 {
			umd.LatestVersion, err = getLatestVersion()
			if err != nil {
				return err
			}

			umd.LastChecked = time.Now()

			if err := umd.save(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (umd *UpdateMetadata) save() error {
	usrHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err

	}

	fp := filepath.Join(usrHomeDir, constants.CLI_HOME_DIR, UPDATE_METADATA_FILE)

	data, err := json.Marshal(umd)
	if err != nil {
		return err
	}

	if err := os.WriteFile(fp, data, 0644); err != nil {
		return err
	}

	return nil
}

func getLatestVersion() (v string, err error) {
	req, err := http.NewRequest("GET", "https://api.github.com/repos/wso2/integration-platform-tools/releases/latest", nil)
	if err != nil {
		return
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var res = struct {
		TagName string `json:"tag_name"`
	}{}

	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return
	}

	v = res.TagName

	return
}

func NewUpdateMetadata() (*UpdateMetadata, error) {
	umd := &UpdateMetadata{}
	if err := umd.load(); err != nil {
		return nil, err
	}

	return umd, nil
}
