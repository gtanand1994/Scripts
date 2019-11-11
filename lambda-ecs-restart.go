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
	webhookUrl := "https://hooks.slack.com/services/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"
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
        	
    		}
	}
	
	fmt.Println(svc_name)
//	creds:= credentials.NewSharedCredentials("./secret","default")
	sess,err:= session.NewSession(&aws.Config{Region: aws.String("us-east-1"),})
	svc := ecs.New(sess)
	input := &ecs.DescribeServicesInput{
	Cluster: aws.String("green-helium-cluster"),
    	Services: []*string{
        aws.String("svc_name"),
    	},
	}

	result, err := svc.DescribeServices(input)
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

	s0:= result.String()
	s1:= strings.Split(s0, "\"deployments\":")
	s2:= strings.Split(s1[1],"],")
	if strings.Contains(s2[0],"\"status\": \"ACTIVE\",") {
		message:= "Restart is already in progress, skipping restart"
		_ = SendSlackNotification(webhookUrl, message)
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
