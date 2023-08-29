package log

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	logRepo LogRepository
	csvRepo CSVRepository
}

func New(logRepo LogRepository, csvRepo CSVRepository) Service {
	return Service{logRepo: logRepo, csvRepo: csvRepo}
}

func (s Service) GenerateCSV(ctx context.Context, request GetCSVRequest) (string, error) {
	logs, err := s.logRepo.Get(ctx, request.UserID, request.From, request.To)
	if err != nil {
		return "", fmt.Errorf("error while getting logs: %w", err)
	}

	var csv []string
	header := strings.Join([]string{"logId", "userId", "segmentId", "operation", "insertTime"}, request.Separator)
	csv = append(csv, header)
	for _, log := range logs {
		row := strings.Join(
			[]string{
				strconv.FormatInt(log.ID, 10),
				strconv.FormatInt(log.UserID, 10),
				log.SegmentID,
				log.Operation,
				log.InsertTime.Format(time.RFC3339),
			},
			request.Separator,
		)
		csv = append(csv, row)
	}

	URI, err := s.csvRepo.Save(strings.Join(csv, "\n"))
	if err != nil {
		return "", fmt.Errorf("error while saving csv: %w", err)
	}

	return URI, nil
}
