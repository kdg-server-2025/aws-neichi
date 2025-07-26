package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// Lambdaサービスクライアントを初期化
var client = lambda.New(session.New())

// AWS SDKを呼び出し、アカウントの情報を取得する
func callLambda() (string, error) {
	input := &lambda.GetAccountSettingsInput{}
	req, resp := client.GetAccountSettingsRequest(input)
	err := req.Send()
	if err != nil {
		return "", err
	}
	output, err := json.Marshal(resp.AccountUsage)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// レスポンスボディの構造を定義
type ResponseBody struct {
	Message         string            `json:"message"`
	CurrentTime     string            `json:"currentTime"`
	LambdaUsage     interface{}       `json:"lambdaUsage"`
	EnvironmentVars map[string]string `json:"environmentVars"`
	LambdaContext   interface{}       `json:"lambdaContext"`
}

// HTTPリクエストを処理するハンドラ関数
// 引数と戻り値をAPIGatewayProxyの型に変更
func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	requestJson, _ := json.MarshalIndent(request, "", "  ")
	log.Printf("REQUEST: %s", requestJson)

	lc, _ := lambdacontext.FromContext(ctx)
	log.Printf("REQUEST ID: %s", lc.AwsRequestID)

	// AWS SDKを呼び出し
	usageStr, err := callLambda()
	if err != nil {
		log.Printf("Error calling AWS SDK: %v", err)
		// エラーが発生したら500を返す
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error getting Lambda account usage.",
		}, err
	}

	var usageJson interface{}
	json.Unmarshal([]byte(usageStr), &usageJson)
	// environmentVars にはパブリックにしないほうが良い情報が含まれているので公開を許可する環境変数のリストを定義
	allowedEnvKeys := map[string]bool{
		"AWS_LAMBDA_FUNCTION_NAME":        true,
		"AWS_LAMBDA_INITIALIZATION_TYPE":  true,
		"AWS_REGION":                      true,
		"AWS_LAMBDA_FUNCTION_MEMORY_SIZE": true,
	}

	// 環境変数をフィルタリングして、許可されたものだけをマップに格納
	filteredEnvVars := make(map[string]string)
	for _, env := range os.Environ() {
		// "KEY=VALUE" の形式の文字列をキーと値に分割
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			key := pair[0]
			// キーがリストに含まれているかチェック
			if allowedEnvKeys[key] {
				filteredEnvVars[key] = pair[1]
			}
		}
	}

	// レスポンスボディを作成
	responseBody := ResponseBody{
		Message:         "Successfully processed request!",
		CurrentTime:     time.Now().Format(time.RFC3339),
		LambdaUsage:     usageJson,
		EnvironmentVars: filteredEnvVars,
		LambdaContext:   lc,
	}

	// レスポンスボディをJSON文字列に変換
	responseJson, err := json.Marshal(responseBody)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	// 正常なレスポンスを返す
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseJson),
	}, nil
}

func main() {
	runtime.Start(handleRequest)
}
