package forgelib

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

	"github.com/aws/aws-sdk-go/service/sts"
)

func valueToString(v interface{}, dest *string, allowSlices bool, allowCommas bool) error {
	if allowSlices && reflect.TypeOf(v).Kind() == reflect.Slice {
		var buffer bytes.Buffer
		vv := reflect.ValueOf(v)
		for i := 0; i < vv.Len(); i++ {
			var thisOutput string
			if err := valueToString(vv.Index(i).Interface(), &thisOutput, false, false); err != nil {
				return err
			}
			if buffer.Len() > 0 {
				buffer.WriteByte(',')
			}
			buffer.WriteString(thisOutput)
		}
		*dest = buffer.String()
		return nil
	}

	switch vv := v.(type) {
	case string:
		for _, b := range vv {
			if !allowCommas && b == ',' {
				return fmt.Errorf("Commas not allowed in list values")
			}
		}
		*dest = vv
	// float64 also captures integers from YAML/JSON
	case float64:
		*dest = strconv.FormatFloat(vv, 'f', -1, 64)
	case bool:
		*dest = strconv.FormatBool(vv)
	default:
		return fmt.Errorf("Field of type %s is not allowed", reflect.TypeOf(vv).Kind().String())
	}
	return nil
}

func roleARNFromName(roleName string) (output string, err error) {
	callerIdentity, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return output, err
	}
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", *callerIdentity.Account, roleName), err
}
