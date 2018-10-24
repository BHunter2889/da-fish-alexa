package main

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	alexa "github.com/mikeflynn/go-alexa/skillserver"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type AlexaResponse struct {
	Version string  `json:"version"`
	Body    ResBody `json:"response"`
}

// ResBody is the actual body of the response
type ResBody struct {
	OutputSpeech     Payload ` json:"outputSpeech,omitempty"`
	ShouldEndSession bool    `json:"shouldEndSession"`
}

// Payload ...
type Payload struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

// NewResponse builds a simple Alexa session response
func NewResponse(speech string) AlexaResponse {
	return AlexaResponse{
		Version: "1.0",
		Body: ResBody{
			OutputSpeech: Payload{
				Type: "PlainText",
				Text: speech,
			},
			ShouldEndSession: true,
		},
	}
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (Response, error) {
	var buf bytes.Buffer

	body, err := json.Marshal(map[string]interface{}{
		"message": "Go Serverless v1.0! Your function executed successfully!",
	})
	if err != nil {
		return Response{StatusCode: 404}, err
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}

	return resp, nil
}

var Applications = map[string]interface{}{
	"/echo/helloworld": alexa.EchoApplication{ // Route
		AppID: "xxxxxxxx", // Echo App ID from Amazon Dashboard
		OnIntent: EchoIntentHandler,
		OnLaunch: EchoIntentHandler,
	},
}

func EchoIntentHandler(echoReq *alexa.EchoRequest, echoResp *alexa.EchoResponse) {
	echoResp.OutputSpeech("Hello world from my new Echo test app!").Card("Hello World", "This is a test card.")
}

func main() {
	lambda.Start(Handler)
	//alexa.Run(Applications, "3000")
}
