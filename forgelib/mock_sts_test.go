package forgelib

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

type mockSTS struct {
	accountID       string
	callerArn       string
	fakeCredentials *sts.Credentials
	stsiface.STSAPI
}

func (m mockSTS) GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	output := sts.GetCallerIdentityOutput{
		Account: aws.String(m.accountID),
		Arn:     aws.String(m.callerArn),
	}
	return &output, nil
}

func (m mockSTS) AssumeRole(*sts.AssumeRoleInput) (*sts.AssumeRoleOutput, error) {
	output := sts.AssumeRoleOutput{Credentials: m.fakeCredentials}
	return &output, nil
}
