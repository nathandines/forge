package forgelib

import "testing"

func TestGetMFASerial(t *testing.T) {
	expectedSerial := "arn:aws:iam::111111111111:mfa/nathan"

	oldIAMClient := iamClient
	defer func() { iamClient = oldIAMClient }()
	iamClient = mockIAM{
		mfaSerial: expectedSerial,
	}

	mfaSerial, err := getMFASerial()
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if e, g := expectedSerial, mfaSerial; e != g {
		t.Errorf("expected \"%s\", got \"%s\"", e, g)
	}
}

func TestGetRoleSessionName(t *testing.T) {
	oldSTSClient := stsClient
	defer func() { stsClient = oldSTSClient }()
	stsClient = mockSTS{
		callerArn: "arn:aws:iam::111111111111:user/nathan",
		accountID: "111111111111",
	}
	expectedName := "nathan"

	roleSessionName, err := getRoleSessionName()
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if e, g := expectedName, roleSessionName; e != g {
		t.Errorf("expected \"%s\", got \"%s\"", e, g)
	}
}

func TestAssumeRole(t *testing.T) {
	oldSTSClient := stsClient
	defer func() { stsClient = oldSTSClient }()
	stsClient = mockSTS{
		callerArn: "arn:aws:iam::111111111111:user/nathan",
		accountID: "111111111111",
	}

	preassumeCfnClient := cfnClient
	preassumeSTSClient := stsClient

	if err := AssumeRole("arn:aws:iam::111111111111:role/test-role"); err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if cfnClient == preassumeCfnClient {
		t.Error("expected cfnClient to have changed, no change detected")
	}
	if stsClient == preassumeSTSClient {
		t.Error("expected stsClient to have changed, no change detected")
	}

	// Cleanup
	cfnClient = preassumeCfnClient
	stsClient = preassumeSTSClient
}

func TestAssumeRoleWithMFA(t *testing.T) {
	oldSTSClient := stsClient
	defer func() { stsClient = oldSTSClient }()
	stsClient = mockSTS{
		callerArn: "arn:aws:iam::111111111111:user/nathan",
		accountID: "111111111111",
	}

	oldIAMClient := iamClient
	defer func() { iamClient = oldIAMClient }()
	iamClient = mockIAM{
		mfaSerial: "arn:aws:iam::111111111111:mfa/nathan",
	}

	preassumeCfnClient := cfnClient
	preassumeIAMClient := iamClient
	preassumeSTSClient := stsClient

	if err := AssumeRoleWithMFA("arn:aws:iam::111111111111:role/test-role", "123456", ""); err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if cfnClient == preassumeCfnClient {
		t.Error("expected cfnClient to have changed, no change detected")
	}
	if iamClient == preassumeIAMClient {
		t.Error("expected iamClient to have changed, no change detected")
	}
	if stsClient == preassumeSTSClient {
		t.Error("expected stsClient to have changed, no change detected")
	}

	// Cleanup
	cfnClient = preassumeCfnClient
	iamClient = preassumeIAMClient
	stsClient = preassumeSTSClient
}
