package contracts

type GoogleAuthRequest struct {
	Credential string `json:"credential" binding:"required"`
}

type GoogleAuthResponse struct {
	Message string `json:"message"`
	User    string `json:"user"`
	Token   string `json:"token"`
}
