package forgelib

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"os"
	"strings"
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
	// Need to mess around with copying values and making pointers to them due
	// to the way in which the AWS Go SDK passes around data
	cfnSessConfig := *sess.Config
	cfnSess := session.Must(session.NewSession(&cfnSessConfig))
	if endpoint, ok := os.LookupEnv("AWS_ENDPOINT_CLOUDFORMATION"); ok {
		cfnSess.Config.Endpoint = aws.String(endpoint)
	}
	cfnClient = cloudformation.New(cfnSess, cfg...)

	stsSessConfig := *sess.Config
	stsSess := session.Must(session.NewSession(&stsSessConfig))
	if endpoint, ok := os.LookupEnv("AWS_ENDPOINT_STS"); ok {
		stsSess.Config.Endpoint = aws.String(endpoint)
	}
	stsClient = sts.New(stsSess, cfg...)
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
