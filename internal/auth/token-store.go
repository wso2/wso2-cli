package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/wso2/integration-platform-tools/internal/region"
	"github.com/wso2/integration-platform-tools/pkg/api"
	"github.com/wso2/integration-platform-tools/pkg/util/constants"
	clikeyring "github.com/wso2/integration-platform-tools/pkg/util/keyring"
)

type OrgTokenStore struct {
	orgStore *OrgStore
}

const (
	FileSuffix    = "_token.enc"
	PathKeySuffix = "_path"
	EncKeySuffix  = "_key"
	serviceName   = "wso2-integration-platform"
)

func NewOrgTokenStore(orgStore OrgStore) *OrgTokenStore {
	return &OrgTokenStore{
		orgStore: &orgStore,
	}
}

// GetToken gets the token for the current region and default org
func (ots *OrgTokenStore) GetToken() (*api.AccessToken, error) {
	sOrg, err := ots.orgStore.GetDefaultOrg()

	if err != nil {
		return nil, err
	}

	if sOrg == nil {
		return nil, fmt.Errorf("no organization info found")
	}

	region := region.GetCurrentRegion()
	token, err := ots.RetrieveToken(region, sOrg.ID)

	if err != nil {
		return nil, fmt.Errorf("error retrieving token: %v", err)
	}

	return token, nil
}

// StoreToken stores a token for a specific region and org
func (ots *OrgTokenStore) StoreToken(region string, orgId string, token *api.AccessToken) error {
	tknStr, err := token.String()

	if err != nil {
		return err
	}

	// Generate a new AES key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return err
	}

	// Encrypt the token
	block, err := aes.NewCipher(key)

	if err != nil {
		return err
	}

	ciphertext := make([]byte, aes.BlockSize+len(tknStr))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(tknStr))

	// Store the encrypted token with region prefix
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, constants.CLI_HOME_DIR, constants.CLI_HOME_AUTH_DIR, region+"_"+orgId+FileSuffix)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(filePath), 0700)
		if err != nil {
			return err
		}
	}

	// Write atomically: a plain WriteFile truncates in place, so a concurrent
	// reader — another CLI process sharing this store — can observe a partial or
	// empty file. Write a sibling temp file and rename, which is atomic on POSIX.
	//
	// 0600: this is encrypted credential material. The AES key lives in the OS
	// keyring, but the ciphertext must not be world-readable either.
	tmpFile, err := os.CreateTemp(filepath.Dir(filePath), filepath.Base(filePath)+".tmp*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // no-op once the rename succeeds

	if err = tmpFile.Chmod(0600); err != nil {
		tmpFile.Close()
		return err
	}
	if _, err = tmpFile.Write(ciphertext); err != nil {
		tmpFile.Close()
		return err
	}
	if err = tmpFile.Close(); err != nil {
		return err
	}
	if err = os.Rename(tmpPath, filePath); err != nil {
		return err
	}

	// Store the file path and the encryption key in the keyring with region prefix
	keyringKey := fmt.Sprintf("%s_%s", region, orgId)
	err = clikeyring.Set(serviceName, keyringKey+PathKeySuffix, filePath)
	if err != nil {
		return err
	}

	err = clikeyring.Set(serviceName, keyringKey+EncKeySuffix, hex.EncodeToString(key))
	if err != nil {
		return err
	}

	return nil
}

// DeleteToken deletes a token for a specific region and org
func (ots *OrgTokenStore) DeleteToken(region string, orgId string) error {
	keyringKey := fmt.Sprintf("%s_%s", region, orgId)
	filePath, err := clikeyring.Get(serviceName, keyringKey+PathKeySuffix)
	if err != nil {
		return err
	}

	// Remove the file
	err = os.Remove(filePath)
	if err != nil {
		return err
	}

	// Remove the keyring entries
	err = clikeyring.Delete(serviceName, keyringKey+PathKeySuffix)
	if err != nil {
		return fmt.Errorf("error deleting file path from keyring: %v", err)
	}

	err = clikeyring.Delete(serviceName, keyringKey+EncKeySuffix)
	if err != nil {
		return fmt.Errorf("error deleting key from keyring: %v", err)
	}

	return nil
}

// RetrieveToken retrieves a token for a specific region and org
func (ots *OrgTokenStore) RetrieveToken(region string, orgId string) (*api.AccessToken, error) {
	keyringKey := fmt.Sprintf("%s_%s", region, orgId)

	// Retrieve the file path and the encryption key from the keyring
	filePath, err := clikeyring.Get(serviceName, keyringKey+PathKeySuffix)
	if err != nil {
		return nil, fmt.Errorf("%w, region=%s, orgId=%s", api.ErrNoTokenFoundForOrg, region, orgId)
	}

	keyString, err := clikeyring.Get(serviceName, keyringKey+EncKeySuffix)
	if err != nil {
		return nil, fmt.Errorf("no key found for organization ID: %s in region: %s", orgId, region)
	}

	key, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, err
	}

	// Read the encrypted token from the file
	ciphertext, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Decrypt the token
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// A concurrent writer can leave the file short or empty (os.WriteFile is not
	// atomic), and slicing an IV off that panics rather than erroring. Callers
	// treat a returned error as "no usable token" and re-authenticate, which is
	// recoverable; a panic is not.
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("stored token for org %s in region %s is truncated (%d bytes); re-authentication required",
			orgId, region, len(ciphertext))
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	token := &api.AccessToken{}
	err = json.Unmarshal(ciphertext, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// HasValidToken checks if there's a valid token for the given region and org
func (ots *OrgTokenStore) HasValidToken(region string, orgId string) bool {
	token, err := ots.RetrieveToken(region, orgId)
	if err != nil {
		return false
	}
	return token != nil && !token.IsExpired()
}
