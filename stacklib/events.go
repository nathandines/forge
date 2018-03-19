package stacklib

import (
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type byTime []*cloudformation.StackEvent

func (t byTime) Len() int           { return len(t) }
func (t byTime) Less(i, j int) bool { return t[i].Timestamp.UnixNano() < t[j].Timestamp.UnixNano() }
func (t byTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

// Events will get all events for a stack and sort them in chronological order
// within a time range
func (s *Stack) Events(after *time.Time) (events []*cloudformation.StackEvent, err error) {
	var stackName *string
	if s.StackID != "" {
		stackName = &s.StackID
	} else {
		stackName = &s.StackName
	}
	if err := cfn.DescribeStackEventsPages(
		&cloudformation.DescribeStackEventsInput{
			StackName: stackName,
		}, func(page *cloudformation.DescribeStackEventsOutput, lastPage bool) bool {
			for _, e := range page.StackEvents {
				if e.Timestamp.UnixNano() > after.UnixNano() {
					events = append(events, e)
				}
			}
			// Continue reading all pages
			return true
		},
	); err != nil {
		return events, err
	}

	sort.Sort(byTime(events))

	return
}

// GetLastEventTime will get the time of the last event for the stack
func (s *Stack) GetLastEventTime() (*time.Time, error) {
	epoch := time.Unix(0, 0)
	events, err := s.Events(&epoch)
	if err != nil {
		return new(time.Time), err
	}
	return events[len(events)-1].Timestamp, err
}
