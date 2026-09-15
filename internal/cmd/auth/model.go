package auth

type SignInOpt struct {
	withToken bool
	token     string

	// TODO: Remove once retrospect ep is given
	userEmail string
	orgHandle string
}
