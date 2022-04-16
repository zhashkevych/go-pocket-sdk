# GetPocket API Golang SDK
## https://getpocket.com/developer/

### Example usage:

```go
package main

import (
	"context"
	"fmt"
	pocket "github.com/zhashkevych/go-pocket-sdk"
	"log"
)

func main()  {
	ctx := context.Background()

	client, err := pocket.NewClient("<your-consumer-key>") // you can generate key at https://getpocket.com/developer/apps/
	if err != nil {
		log.Fatalf(`Couldn't create new client: %v`, err)
	}

	requestToken, err := client.GetRequestToken(ctx, "http://example.com/")
	if err != nil {
		log.Fatalf("failed to get request token: %s", err.Error())
	}

	url, err := client.GetAuthorizationURL(requestToken, "http://example.com/")
	if err != nil {
		log.Fatalf(`Couldn't get authorization url: %v`, err)
	}
	fmt.Println(url)

	authResp, err := client.Authorize(ctx, requestToken)
	if err != nil {
		log.Fatalf("failed to authorize: %s", err)
	}

	err = client.Add(ctx, pocket.AddInput{
		URL: "https://github.com/zhashkevych/go-pocket-sdk",
		AccessToken: authResp.AccessToken,
	})
	if err != nil {
		log.Fatalf("failed to add item: %s", err)
	}
}
```

