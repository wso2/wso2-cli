package auth

import "fmt"

func SignOut() error {
	// get current user info
	userInfo, _ := usrStore.RetrieveUserInfo()
	ClearAuthStores()
	if userInfo == nil {
		return fmt.Errorf("no user info found")
	}

	return nil
}
