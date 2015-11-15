package smsgwclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// SmsGatewayClient for sending messages through the Intelecom SMS Gateway.
type SmsGatewayClient struct {
	httpClient                  *http.Client
	serviceID                   int
	baseURL, username, password string
}

// Message that will be sent using the SmsGatewayClient.
type Message struct {
	// The MSISDN of the recipient.
	// The format should follow the ITUT E.164 standard with a + prefix.
	Recipient string `json:"recipient"`

	// The message payload to send, typically the message text.
	Content string `json:"content"`

	// The cost for the recipient to receive the message. In lowest monetary unit.
	Price int `json:"price"`

	// Arbitrary client reference ID that will be returned in the message response.
	ClientReference string `json:"clientReference,omitempty"`

	// For advanced message settings.
	Settings *Settings `json:"settings,omitempty"`
}

// Settings for a Message.
type Settings struct {
	// Uses service value unless specified.
	// Used to prioritize between messages sent from the same service.
	// 1: low (slower), 2: medium, 3: high(faster)
	Priority int `json:"priority,omitempty"`

	// Uses service value unless specified.
	// Specifies the TTL(time to live) for the message,
	// i.e.how long before the message times out in cases
	// where it cannot be delivered to a handset.
	Validity int `json:"validity,omitempty"`

	// Arbitrary string set by the client to enable grouping messages in certain statistic reports.
	Differentiator string `json:"differentiator,omitempty"`

	// Only relevant for CPA/GAS messages.
	// Defines an age limit for message content.
	// The mobile network operators enforces this.
	// IMPORTANT: If the service is a subscription service all CPA/GAS messages must have age set to 18.
	// Valid values: 0, 16 or 18.
	Age int `json:"age,omitempty"`

	// Used to start a new session.
	NewSession bool `json:"newSession,omitempty"`

	// Used to continue an existing session.
	SessionID string `json:"sessionId,omitempty"`

	// Arbitrary string set by the client to enable grouping messages on the service invoice.
	InvoiceNode string `json:"invoiceNode,omitempty"`

	// Currently not in use.
	AutoDetectEncoding bool `json:"autoDetectEncoding,omitempty"`

	// If set to true the gateway will remove or safely substitute invalid characters in the message content instead of rejecting the message.
	SafeRemoveNonGsmCharacters bool `json:"safeRemoveNonGsmCharacters,omitempty"`

	// Uses service value unless specified. Used to specify the originator.
	OriginatorSettings *OriginatorSettings `json:"originatorSettings,omitempty"`

	// Uses service value unless specified. Used if the message is a CPA Goods and Services transaction.
	GasSettings *GasSettings `json:"gasSettings,omitempty"`

	// Used if the message should be queued and sent in the future instead of immediately.
	SendWindow *SendWindow `json:"sendWindow,omitempty"`

	// Used to specify special settings including settings for binary message.
	Parameters map[string]string `json:"parameter,omitempty"`
}

// OriginatorSettings for a message.
type OriginatorSettings struct {
	//Specifies the type of originator.
	OriginatorType string `json:"originatorType"`

	// Depends on the OriginatorType. Example: +4799999999, Intelecom, 1960.
	Originator string `json:"originator"`
}

// GasSettings for a message.
type GasSettings struct {
	// Identifier for the category of Goods and services.
	ServiceCode string `json:"serviceCode"`

	// Further details of the Goods and services.
	// The description may occur on the end-user invoice(together with category)
	// for certain Mobile Network Operators.
	Description string `json:"description,omitempty"`
}

// SendWindow for a message.
type SendWindow struct {
	// The date to send the message.
	StartDate time.Time `json:"startDate"`

	// The time of day to start sending the message.
	StartTime *time.Time `json:"startTime,omitempty"`

	// The date to stop sending the message if the message is still enqueued.
	StopDate *time.Time `json:"stopDate,omitempty"`

	// The time to stop sending the message if the message is still enqueued.
	StopTime *time.Time `json:"stopTime,omitempty"`
}

// smsGatewayRequest contains common parameters.
type smsGatewayRequest struct {
	ServiceID      int       `json:"serviceId"`
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	BatchReference string    `json:"batchReference,omitempty"`
	Messages       []Message `json:"message"`
}

// SmsGatewayResponse is returned from the Send method.
type SmsGatewayResponse struct {
	// Reference ID for the request.
	// Either the value provided by the client in
	// the request or an automatically generated ID if no such value is set.
	BatchReference string `json:"batchReference"`

	// The status of each message sent to the gateway.
	MessageStatus []MessageStatus `json:"messageStatus"`
}

// MessageStatus for a message.
type MessageStatus struct {
	// Status code.
	StatusCode int `json:"statusCode"`

	// Textual information about status, e.g. which parameter failed.
	StatusMessage string `json:"statusMessage"`

	// The client reference ID if specified in the request.
	ClientReference string `json:"clientReference"`

	// The recipient.
	// NOTE: The gateway runs all numbers through a number parser so
	// the recipient in the response may not be in same format as in the request,
	// i.e. “+47 41 00 00 00” will be “+4741000000” in the response.
	// Use the ClientReference if you need to match messages in the request and response.
	Recipient string `json:"recipient"`

	// Message ID (used as reference for delivery reports).
	MessageID string `json:"messageId"`

	// Session ID for a session.
	// Only returned if NewSession parameter is set to true, or if you are specifying a session ID.
	SessionID string `json:"sessionId"`

	// The messages in the response will always be in the same order as in the request.
	// The sequence index is a convenience counter starting at 1.
	SequenceIndex int `json:"sequenceIndex"`
}

// MakeSmsGatewayClient returns an SmsGatewayClient.
func MakeSmsGatewayClient(baseURL string, serviceID int, username, password string) SmsGatewayClient {
	return SmsGatewayClient{
		httpClient: &http.Client{},
		baseURL:    baseURL,
		serviceID:  serviceID,
		username:   username,
		password:   password}
}

func createRequest(url string, buf []byte) *http.Request {
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")
	return request
}

// Send is used to send one or more messages using the GatewayClient
func (client SmsGatewayClient) Send(messages []Message) (SmsGatewayResponse, error) {
	gatewayRequest := smsGatewayRequest{
		ServiceID: client.serviceID,
		Username:  client.username,
		Password:  client.password,
		Messages:  messages}
	buffer, err := json.Marshal(gatewayRequest)
	if err != nil {
		return SmsGatewayResponse{}, err
	}
	request := createRequest(client.baseURL+"/sendMessages", buffer)
	response, err := client.httpClient.Do(request)
	defer response.Body.Close()
	if err != nil {
		return SmsGatewayResponse{}, err
	}
	if response.StatusCode != 200 {
		return SmsGatewayResponse{}, errors.New("Status code: " + strconv.Itoa(response.StatusCode))
	}
	var gatewayResponse SmsGatewayResponse
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&gatewayResponse)
	if err != nil {
		return SmsGatewayResponse{}, err
	}
	return gatewayResponse, nil
}
