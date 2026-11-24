package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opencairn-org/cairn-core/internal/domain"
	"github.com/spf13/cobra"
)

func newExtractPolylineCmd() *cobra.Command {
	var activityId string
	var startTime string
	var endTime string
	cmd := &cobra.Command{
		Use:   "extract-polyline",
		Short: "Extract polyline from an activity to create a segment",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runExtractPolyline(cmd.Context(), activityId, startTime, endTime)
		},
	}
	cmd.Flags().StringVarP(&activityId, "activity-id", "a", "", "UUID of the activity")
	cmd.Flags().StringVarP(&startTime, "start", "s", "", "start time for the segment, ex: 2026-01-02T15:04:00-03:00 or 2026-11-22T18:00:00Z for UTC (which is what we have in the GPX files)")
	cmd.Flags().StringVarP(&endTime, "end", "e", "", "end time for the segment")
	cmd.MarkFlagRequired("activity-id")
	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")
	return cmd
}

func runExtractPolyline(context context.Context, activityId, startTime, endTime string) error {
	start, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return fmt.Errorf("invalid start time (want RFC3339): %w", err)
	}
	end, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		return fmt.Errorf("invalid end time (want RFC3339): %w", err)
	}
	fmt.Printf("Will extract segment from activity %s, from %s to %s\n", activityId, start, end)

	runtime, err := newRuntime(context)
	if err != nil {
		return fmt.Errorf("runtime: %w", err)
	}
	defer runtime.close()

	basicApp, err := newBasicApp(runtime)

	if err != nil {
		return fmt.Errorf("basic app: %w", err)
	}

	err = loadSegment(context, activityId, basicApp, start, end)
	if err != nil {
		return err
	}

	//activity.StartTime

	return nil
}

func loadSegment(context context.Context, activityId string, basicApp *BasicApp, start time.Time, end time.Time) error {
	id, err := uuid.Parse(activityId)
	if err != nil {
		return fmt.Errorf("invalid activity id: %w", err)
	}
	activity, err := basicApp.Activities.GetActivity(context, domain.ActivityID(id))
	if err != nil {
		return fmt.Errorf("error loading activity id: %w", err)
	}

	fmt.Printf("Activity started at %s and finished at %s\n", activity.StartTime, activity.EndTime)

	// go run ./cmd/server extract-polyline -a 01a202ff-f0e1-7a72-8ba2-b9d318fe9d79 -s 2026-01-02T15:04:00-03:00 -e 2026-01-02T15:14:00-03:00

	if activity.EndTime.Compare(start) < 0 {
		return fmt.Errorf("activity end time (%s) before segment start time (%s)", activity.EndTime, start)
	}
	if activity.EndTime.Compare(end) < 0 {
		return fmt.Errorf("activity end time (%s) before the segment end time (%s)", activity.EndTime, end)
	}

	if activity.StartTime.Compare(start) > 0 {
		return fmt.Errorf("activity start time (%s) after the segment start time (%s)", activity.StartTime, start)
	}
	if activity.StartTime.Compare(end) > 0 {
		return fmt.Errorf("activity start time (%s) after the segment end time (%s)", activity.StartTime, end)
	}

	query := domain.StreamQuery{
		ActivitySourceID: *activity.PrimaryStreamSourceID,
		Channels:         []domain.StreamChannel{domain.StreamChannelLatitude, domain.StreamChannelLongitude},
		StartTime:        start,
		EndTime:          end,
		Resolution:       domain.StreamResolutionRaw,
	}

	stream, err := basicApp.Streams.QueryStream(context, query)
	if err != nil {
		return fmt.Errorf("error loading stream: %w", err)
	}

	fmt.Printf("loaded stream %d", len(stream.Samples))

	return nil
}
