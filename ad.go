package main

import (
	"fmt"
	"os"

	auth "github.com/korylprince/go-ad-auth/v3"
)

type ADUserInfo struct {
	ID   string
	Name string
}

func GetUserInfo(key string) ADUserInfo {
	rwm.RLock()
	defer rwm.RUnlock()
	return authenticatedUsers[key]
}

func SetUserInfo(key string, value ADUserInfo) {
	rwm.Lock()
	defer rwm.Unlock()
	authenticatedUsers[key] = value
}

// CheckUsernameAndPassword - authenticates user
func CheckUsernameAndPassword(username, password string) bool {

	adServer := os.Getenv("FORM503ADSERVER")
	adBaseDN := os.Getenv("FORM503ADBASEDN")

	config := &auth.Config{
		Server:   adServer,
		Port:     389,
		BaseDN:   adBaseDN,
		Security: auth.SecurityInsecureStartTLS,
	}
	c, err := config.Connect()
	if err != nil {
		fmt.Println("Error connecting:", err)
		return false
	}
	defer c.Conn.Close()

	upn, err := config.UPN(username)
	if err != nil {
		fmt.Println("Error UPNing:", err)
		return false
	}
	//fmt.Println("UPN:", upn)

	status, err := c.Bind(upn, password)
	if err != nil {
		fmt.Println("Error binding:", err)
		return false
	}
	if !status {
		//handle failed authentication
		fmt.Println("status:", status)
		return false
	}

	entry, err := c.GetAttributes("userPrincipalName", upn, []string{"employeeNumber", "cn"})
	if err != nil {
		fmt.Println("Attribute Error! " + username)
		fmt.Println("Success authenticating, error getting attributes:", err)
		return false
	}

	//fmt.Println("Success! " + username + " = " + entry.GetAttributeValue("employeeNumber"))
	userInfo := ADUserInfo{
		ID:   entry.GetAttributeValue("employeeNumber"),
		Name: entry.GetAttributeValue("cn"),
	}
	SetUserInfo(username, userInfo)

	return true
}
