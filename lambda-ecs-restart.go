package main
import (
    "fmt"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ecs"
//    "github.com/aws/aws-sdk-go-v2/service/ecs"
    "github.com/aws/aws-sdk-go/aws/awserr"
    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-lambda-go/events"
    "context"
    "net/http"
    "bytes"
    "encoding/json"
    "time"
    "errors"
    "strings"
)

type SlackRequestBody struct {
    Text string `json:"text"`
}

func SendSlackNotification(webhookUrl string, msg string) error {

    slackBody, _ := json.Marshal(SlackRequestBody{Text: msg})
    req, err := http.NewRequest(http.MethodPost, webhookUrl, bytes.NewBuffer(slackBody))
    if err != nil {
        return err
    }

    req.Header.Add("Content-Type", "application/json")

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }

    buf := new(bytes.Buffer)
    buf.ReadFrom(resp.Body)
    if buf.String() != "ok" {
        return errors.New("Non-ok response returned from Slack")
    }
    return nil
}

func handleRequest(ctx context.Context, snsEvent events.SNSEvent) (string, error) {
	svc_name := ""
// Fill the Slack webhookURL 
	webhookUrl := "" 
	for _, record := range snsEvent.Records {
		snsRecord := record.SNS
	        s1:= strings.Split(snsRecord.Message, "[")
		s2:=strings.Split(s1[2],"]")
		result:=strings.Split(s2[0],"\"")

//		result:=strings.Split(strings.Split(strings.Split(msg,"[")[2],"]")[0],"\"")
		svc_name = strings.Split(strings.Split(strings.Split(snsRecord.Message,"[")[2],"]")[0],"\"")[3]
		svc_name = result[3]
		message := fmt.Sprintf("[%s %s] Message = %s", record.EventSource, snsRecord.Timestamp, svc_name)
		fmt.Println(message)
		err := SendSlackNotification(webhookUrl, message)
    		if err != nil {
        		fmt.Println(err)
    		}
	}
	
	fmt.Println(svc_name)
//	creds:= credentials.NewSharedCredentials("./secret","default")
	sess,err:= session.NewSession(&aws.Config{Region: aws.String("us-east-1"),})
	svc := ecs.New(sess)
	input := &ecs.DescribeServicesInput{
	Cluster: aws.String("green-helium-cluster"),
    	Services: []*string{
        aws.String(svc_name),
    	},
	}

	result, err := svc.DescribeServices(input)
	fmt.Println("Describe service done")
	if err != nil {
    	if aerr, ok := err.(awserr.Error); ok {
        	switch aerr.Code() {
        	case ecs.ErrCodeServerException:
            	fmt.Println(ecs.ErrCodeServerException, aerr.Error())
        	case ecs.ErrCodeClientException:
            	fmt.Println(ecs.ErrCodeClientException, aerr.Error())
        	case ecs.ErrCodeInvalidParameterException:
            	fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
        	case ecs.ErrCodeClusterNotFoundException:
            	fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
        	default:
            	fmt.Println(aerr.Error())
        	}
    	} else {
        // Print the error, cast err to awserr.Error to get the Code and
        // Message from an error.
        	fmt.Println(err.Error())
    	}
//    	return "success",err
	}

// UPDATE STARTS HERE:
	fmt.Println("Update Starts here")
	fmt.Println(result)

	s0:= fmt.Sprintf("%v",result)
	fmt.Println("String conversion")
	s1:= strings.Split(s0, "Deployments: [{")
	fmt.Println("Split deployments section")
	s2:= strings.Split(s1[1],"],")
	fmt.Println("Split deployments section")
	if strings.Contains(s2[0],"Status: \"ACTIVE\",") {
		_ = SendSlackNotification(webhookUrl, "Restart is already in progress, skipping restart")
	        fmt.Println("Restart is already in progress, skipping restart")
	}else {
		fmt.Println("initiate restart")
		_ = SendSlackNotification(webhookUrl, fmt.Sprintf("Initiating restart of %s",svc_name))
		update_in:= &ecs.UpdateServiceInput{
		Cluster: aws.String("green-helium-cluster"),
		Service: aws.String(svc_name),
		ForceNewDeployment: aws.Bool(true),
		}
		update_result, err := svc.UpdateService(update_in)
		if err != nil {
    		if aerr, ok := err.(awserr.Error); ok {
       		 	switch aerr.Code() {
        		case ecs.ErrCodeServerException:
            		fmt.Println(ecs.ErrCodeServerException, aerr.Error())
        		case ecs.ErrCodeClientException:
            		fmt.Println(ecs.ErrCodeClientException, aerr.Error())
        		case ecs.ErrCodeInvalidParameterException:
            		fmt.Println(ecs.ErrCodeInvalidParameterException, aerr.Error())
        		case ecs.ErrCodeClusterNotFoundException:
            		fmt.Println(ecs.ErrCodeClusterNotFoundException, aerr.Error())
        		default:
            		fmt.Println(aerr.Error())
        		}
    		} else {
        	// Print the error, cast err to awserr.Error to get the Code and
        	// Message from an error.
        		fmt.Println(err.Error())
    		}
    		return "success",err
		}
		fmt.Println(update_result)
	}	
	return "Restarted", nil
}

func main() {
	lambda.Start(handleRequest)
}

// Steps to deploy the script in lambda.
// 1. GOARCH=amd64 GOOS=linux go build ecs-memory.go
// <file_name> = handler name in lambda function = our script filename
// 2. zip ecs-memory.zip ecs-memory
// 3. Upload the zip file in lambda console
