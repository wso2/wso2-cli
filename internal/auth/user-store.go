package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/util/constants"
	clikeyring "github.com/wso2/integration-platform-tools/pkg/util/keyring"
)

const (
	USR_INFO_PATH_KEY = "user_info_path"
	USR_INFO_KEY      = "user_info"
	USR_INFO_FILE     = "user_info.enc"
)

type UserStore struct {
}

func NewUserStore() *UserStore {
	return &UserStore{}
}

func (us *UserStore) StoreUserInfo(usrInfo api.UserInfo) (err error) {
	usrStr := usrInfo.String()

	key := make([]byte, 32) // AES-256

	if _, err := rand.Read(key); err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	ciphertext := make([]byte, aes.BlockSize+len(usrStr))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(usrStr))

	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, constants.CLI_HOME_DIR, constants.CLI_HOME_AUTH_DIR, USR_INFO_FILE)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(filePath), 0700)
		if err != nil {
			return err
		}
	}

	// 0600: this is encrypted credential material. The AES key lives in the
	// OS keyring, but the ciphertext must not be world-readable either.
	err = os.WriteFile(filePath, ciphertext, 0600)
	if err != nil {
		return err
	}

	// WriteFile only applies the mode when creating the file, so tighten
	// explicitly to also repair files written by older versions at 0644.
	if err = os.Chmod(filePath, 0600); err != nil {
		return err
	}

	err = clikeyring.Set(serviceName, USR_INFO_PATH_KEY, filePath)
	if err != nil {
		return err
	}

	err = clikeyring.Set(serviceName, USR_INFO_KEY, hex.EncodeToString(key))
	if err != nil {
		return err
	}

	return nil
}

func (us *UserStore) GenerateUserInfoFromToken(stsToken string, orgId string) (*api.UserInfo, error) {
	orgs, err := OrgClient.GetOrganizations()
	if err != nil {
		return nil, err
	}

	token := api.AccessToken{AccessToken: stsToken}
	claims := token.GetDecodedToken()

	// generate new user info by decoding access token and also fetching orgs or user
	userInfo := &api.UserInfo{
		IDPId:         claims.Sub,
		DisplayName:   claims.IdpClaims.Name,
		UserEmail:     claims.IdpClaims.Email,
		Organizations: orgs,
	}
	if err = usrStore.StoreUserInfo(*userInfo); err != nil { // persist user info
		return nil, err
	}
	return userInfo, nil
}

func (us *UserStore) RetrieveUserInfo() (*api.UserInfo, error) {
	filePath, err := clikeyring.Get(serviceName, USR_INFO_PATH_KEY)
	if err != nil {
		return nil, api.ErrNotLoggedIn
	}

	keyString, err := clikeyring.Get(serviceName, USR_INFO_KEY)
	if err != nil {
		return nil, api.ErrNotLoggedIn
	}

	key, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, err
	}

	ciphertext, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	usrInfo := &api.UserInfo{}
	err = json.Unmarshal(ciphertext, usrInfo)
	if err != nil {
		return nil, err
	}

	return usrInfo, nil
}

func (us *UserStore) ClearUserInfo() error {
	err := clikeyring.Delete(serviceName, USR_INFO_PATH_KEY)
	if err != nil {
		return err
	}

	err = clikeyring.Delete(serviceName, USR_INFO_KEY)
	if err != nil {
		return err
	}

	return nil
}
