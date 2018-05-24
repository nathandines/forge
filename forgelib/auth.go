package forgelib

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

var originalSession *session.Session
var cfnClient cloudformationiface.CloudFormationAPI // CloudFormation Service
var iamClient iamiface.IAMAPI                       // IAM Service
var stsClient stsiface.STSAPI                       // STS Service

func init() {
	stscreds.DefaultDuration = time.Duration(60) * time.Minute
	originalSession = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
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

	iamConfig := aws.Config{}
	if endpoint, ok := os.LookupEnv("AWS_ENDPOINT_IAM"); ok {
		iamConfig.Endpoint = aws.String(endpoint)
	}
	iamConfigs := append([]*aws.Config{&generalConfig, &cfnConfig}, cfg...)
	iamClient = iam.New(sess, iamConfigs...)

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
	session, err := setupRoleSession(assumeOut)
	if err != nil {
		return err
	}
	setupClients(session)
	return nil
}

// AssumeRoleWithMFA performs the same function as AssumeRole, but accepts an
// MFA token as well. A blank value for mfaSerial will attempt to auto-detect
// the serial of the users MFA
func AssumeRoleWithMFA(roleArn, mfaToken, mfaSerial string) error {
	if mfaSerial == "" {
		var err error
		mfaSerial, err = getMFASerial()
		if err != nil {
			return err
		}
	}
	roleSessionName, err := getRoleSessionName()
	if err != nil {
		return err
	}
	assumeOut, err := stsClient.AssumeRole(&sts.AssumeRoleInput{
		DurationSeconds: aws.Int64(3600),
		RoleSessionName: aws.String(roleSessionName),
		RoleArn:         aws.String(roleArn),
		SerialNumber:    aws.String(mfaSerial),
		TokenCode:       aws.String(mfaToken),
	})
	if err != nil {
		return err
	}
	session, err := setupRoleSession(assumeOut)
	if err != nil {
		return err
	}
	setupClients(session)
	return nil
}

func setupRoleSession(assumeRoleOutput *sts.AssumeRoleOutput) (*session.Session, error) {
	session, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(
				*assumeRoleOutput.Credentials.AccessKeyId,
				*assumeRoleOutput.Credentials.SecretAccessKey,
				*assumeRoleOutput.Credentials.SessionToken,
			),
		},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}

// UnassumeAllRoles will change your credentials back to their original state
// after using AssumeRole
func UnassumeAllRoles() {
	setupClients(originalSession)
}

func getMFASerial() (string, error) {
	mfaInfo, err := iamClient.ListMFADevices(&iam.ListMFADevicesInput{})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "AccessDenied" {
				return "", fmt.Errorf("Access Denied to list available MFA devices. Please specify an MFA serial manually")
			}
			return "", err
		}
		return "", err
	}
	if len(mfaInfo.MFADevices) > 0 {
		return *mfaInfo.MFADevices[0].SerialNumber, nil
	}
	return "", fmt.Errorf("MFA device not found for the current user")
}

func getRoleSessionName() (string, error) {
	callerIdentity, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	roleSessionNameSlice := strings.Split(*callerIdentity.Arn, "/")
	return roleSessionNameSlice[len(roleSessionNameSlice)-1], nil
}
