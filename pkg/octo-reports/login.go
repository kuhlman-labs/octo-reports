package octoreports

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cli/oauth/device"
)

func RequestCode(url, clientID string) (string, error) {

	scopes := []string{"repo", "read:org", "read:user", "read:enterprise"}

	httpClient := http.DefaultClient

	code, err := device.RequestCode(httpClient, url+"/login/device/code", clientID, scopes)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Copy code: %s\n", code.UserCode)
	fmt.Printf("then open: %s\n", code.VerificationURI)

	accessToken, err := device.Wait(context.TODO(), httpClient, url+"/login/oauth/access_token", device.WaitOptions{
		ClientID:   clientID,
		DeviceCode: code,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Access token: %s\n", accessToken.Token)

	return accessToken.Token, nil
}
