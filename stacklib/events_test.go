package stacklib

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type mockEvents struct {
	cloudformationiface.CloudFormationAPI
	StackEventsOutput *cloudformation.DescribeStackEventsOutput
}

func (m mockEvents) DescribeStackEventsPages(input *cloudformation.DescribeStackEventsInput, function func(*cloudformation.DescribeStackEventsOutput, bool) bool) error {
	function(m.StackEventsOutput, true)
	return nil
}

func TestGetLastEventTime(t *testing.T) {
	cases := []struct {
		resp     []time.Time
		expected time.Time
	}{
		{
			resp: []time.Time{
				time.Unix(100, 0),
				time.Unix(300, 0),
				time.Unix(200, 0),
			},
			expected: time.Unix(300, 0),
		},
	}
	for i, c := range cases {
		var stackEvents []*cloudformation.StackEvent
		for i := range c.resp {
			event := &cloudformation.StackEvent{
				Timestamp: &c.resp[i],
			}
			stackEvents = append(stackEvents, event)
		}
		cfn = mockEvents{
			StackEventsOutput: &cloudformation.DescribeStackEventsOutput{
				StackEvents: stackEvents,
			},
		}

		s := Stack{}
		result, err := s.GetLastEventTime()
		if err != nil {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}
		if *result != c.expected {
			t.Errorf("%d, expected %v time, got %v", i, c.expected, result)
		}
	}
}
