package forgelib

import "testing"

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
