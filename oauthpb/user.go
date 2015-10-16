package oauthpb

import (
	"fmt"
	"log"
	"net/http"

	"appengine"
	"appengine/user"
)

func TryUser(r *http.Request) (*user.User, string) {

	c := appengine.NewContext(r)
	uType := ""

	//
	//
	u := user.Current(c)
	if u == nil {
		uType += "OAuth1 fail "
	} else {
		uType += "OAuth1 succ "
	}
	uType += "\n"

	//
	//
	u2, err := user.CurrentOAuth(c, "")
	if err != nil {
		uType += fmt.Sprintf("OAuth2 fail: %v", err)
	}
	if u2 != nil {
		uType += "OAuth2 succ"
	}
	uType += "\n"

	if appengine.IsDevAppServer() {
		uType += "Logon always shines on DEV system."
	}

	// Replace
	if u == nil {
		u = u2
	}

	return u, uType
}

func Auth(r *http.Request) (bool, *user.User, string) {

	u, msg := TryUser(r)

	// var err error
	// u, err = user.Current()
	// if err != nil {
	// 	msg += "user.Current() returned error :" + err.Error()
	// 	return

	if u == nil {
		msg += "google oauth required"
		return false, nil, msg
	}

	log.Printf("is admin: %v", u.Admin)

	// if u.ID != "108853175242330402880" && u.ID != "S-1-5-21-2175189548-897864986-1736798499-1000" {
	// 	msg += "you need to be me; not " + u.ID
	// 	return false, u, msg
	// }

	return true, u, msg

}

// Show status and show login/logut url
func login(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	u, uType := TryUser(r)

	if u == nil {
		fmt.Fprintf(w, "%v<br>\n", uType)
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	} else {
		// this gets never executed on dev server
		fmt.Fprintf(w, "Hello, %v, %v, %v, %v!<br>\n", u, u.ID, u.Email, u.FederatedIdentity)
		fmt.Fprintf(w, "Login type %v<br>\n", uType)
		url2, err := user.LogoutURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprintf(w, "<a href='%v'>logout</a>", url2)
	}

}

func init() {
	http.HandleFunc("/login", login)
}
