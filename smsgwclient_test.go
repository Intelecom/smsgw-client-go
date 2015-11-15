package smsgwclient_test

import (
	"testing"

	"github.com/intelecom/smsgw-client-go"
	"github.com/jarcoal/httpmock"
)

func TestErrorIsNilWhenOkResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://www.dummy-address.com/sendMessages",
		httpmock.NewStringResponder(200, `{"batchReference":"123abc","messageStatus":[]}`))

	gatewayClient := smsgwclient.MakeSmsGatewayClient("https://www.dummy-address.com", 0, "", "")
	messages := []smsgwclient.Message{smsgwclient.Message{}}
	if _, err := gatewayClient.Send(messages); err != nil {
		t.FailNow()
	}
}

func TestErrorIsNotNilWhenErrorResponse(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://www.dummy-address.com/sendMessages",
		httpmock.NewStringResponder(500, ""))

	gatewayClient := smsgwclient.MakeSmsGatewayClient("https://www.dummy-address.com", 0, "", "")
	messages := []smsgwclient.Message{smsgwclient.Message{}}
	if _, err := gatewayClient.Send(messages); err == nil {
		t.FailNow()
	}
}
