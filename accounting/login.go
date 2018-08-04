package accounting

import "log"

type loginRequest struct {
	username string
	password string
}

var loginChannel = make(chan *loginRequest)

// Login requests an account login
func Login(username, password string) {
	if loginChannel != nil {
		loginChannel <- &loginRequest{
			username,
			password,
		}
	}
}

func doLogin(r *loginRequest) {
	log.Println("Login request for", r.username)
}
