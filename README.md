## Build status ##

Travis: [![Build Status](https://travis-ci.org/Intelecom/smsgw-client-go.svg?branch=master)](https://travis-ci.org/Intelecom/smsgw-client-go)

## Installation ##

```sh
go get github.com/intelecom/smsgw-client-go
```

or by importing it into a project

```sh
import "github.com/intelecom/smsgw-client-go"
```

## Example usage ##

```go
// Initialize the client
gatewayClient := smsgwclient.MakeSmsGatewayClient(baseURL, serviceID, username, password)

// Single recipient, 0 NOK
message := smsgwclient.Message{Recipient: "+47XXXXXXXX",Content:"This is a test"}
messages := []smsgwclient.Message{message}

if resp, err := gatewayClient.Send(messages); err != nil {
	log.Println("error:", err)
} else {
	log.Println("resp:", resp)
}
```