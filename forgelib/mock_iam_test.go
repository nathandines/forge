package forgelib

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

type mockIAM struct {
	mfaSerial string
	iamiface.IAMAPI
}

func (m mockIAM) ListMFADevices(*iam.ListMFADevicesInput) (*iam.ListMFADevicesOutput, error) {
	mfaDevice := iam.MFADevice{
		SerialNumber: aws.String(m.mfaSerial),
	}
	output := iam.ListMFADevicesOutput{
		MFADevices: []*iam.MFADevice{
			&mfaDevice,
		},
	}
	return &output, nil
}
