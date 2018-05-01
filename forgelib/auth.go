package forgelib

import (
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

var originalSession *session.Session
var cfnClient cloudformationiface.CloudFormationAPI // CloudFormation service
var stsClient stsiface.STSAPI                       // STS Service

func init() {
	originalSession = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	setupClients(originalSession)
}

func setupClients(sess *session.Session, cfg ...*aws.Config) {
	generalConfig := aws.Config{
		MaxRetries: aws.Int(10),
	}
	// Need to mess around with copying values and making pointers to them due
	// to the way in which the AWS Go SDK passes around data
	cfnConfig := aws.Config{}
	if endpoint, ok := os.LookupEnv("AWS_ENDPOINT_CLOUDFORMATION"); ok {
		cfnConfig.Endpoint = aws.String(endpoint)
	}
	cfnConfigs := append([]*aws.Config{&generalConfig, &cfnConfig}, cfg...)
	cfnClient = cloudformation.New(sess, cfnConfigs...)

	stsConfig := aws.Config{}
	if endpoint, ok := os.LookupEnv("AWS_ENDPOINT_STS"); ok {
		stsConfig.Endpoint = aws.String(endpoint)
	}
	stsConfigs := append([]*aws.Config{&generalConfig, &stsConfig}, cfg...)
	stsClient = sts.New(sess, stsConfigs...)
}

// AssumeRole will change your credentials for Forge to those of an assumed role
// as specific by the ARN specified in the arguments to AssumeRole
func AssumeRole(roleArn string) error {
	roleSessionName, err := getRoleSessionName()
	if err != nil {
		return err
	}
	assumeOut, err := stsClient.AssumeRole(&sts.AssumeRoleInput{
		DurationSeconds: aws.Int64(900),
		RoleSessionName: aws.String(roleSessionName),
		RoleArn:         aws.String(roleArn),
	})
	if err != nil {
		return err
	}
	session, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(
				*assumeOut.Credentials.AccessKeyId,
				*assumeOut.Credentials.SecretAccessKey,
				*assumeOut.Credentials.SessionToken,
			),
		},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return err
	}
	setupClients(session)
	return nil
}

// UnassumeAllRoles will change your credentials back to their original state
// after using AssumeRole
func UnassumeAllRoles() {
	setupClients(originalSession)
}

func getRoleSessionName() (string, error) {
	callerIdentity, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	roleSessionNameSlice := strings.Split(*callerIdentity.Arn, "/")
	return roleSessionNameSlice[len(roleSessionNameSlice)-1], nil
}
