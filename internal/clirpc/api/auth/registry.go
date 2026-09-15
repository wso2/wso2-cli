package auth

import "github.com/wso2/integration-platform-tools/internal/clirpc/server"

func init() {
	server.RegisterHandler("auth/getUserInfo", GetUserInfo)
	server.RegisterHandler("auth/signOut", SignOutUser)
	server.RegisterHandler("auth/getCurrentRegion", GetCurrentRegion)
	server.RegisterHandler("auth/getSignInUrl", GetSignInAuthUrl)
	server.RegisterHandler("auth/getDevantSignInUrl", GetDevantSignInAuthUrl)
	server.RegisterHandler("auth/signInWithAuthCode", SignInWithAuthCode)
	server.RegisterHandler("auth/signInDevantWithAuthCode", SignInDevantWithAuthCode)
	server.RegisterHandler("auth/getCurrentOrg", GetCurrentOrg)
	server.RegisterHandler("auth/changeOrg", ChangeOrg)
	server.RegisterHandler("auth/getSubscriptions", GetSubscriptions)
	server.RegisterHandler("auth/getStsToken", GetStsToken)
	server.RegisterHandler("auth/getConfigs", getConfigs)
}
