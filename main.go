package main

import (
	"encoding/json"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/fictionbase/fictionbase"
	"github.com/fictionbase/fictionbase/type/fbhttp"
	"github.com/fictionbase/fictionbase/type/fbresource"
	"go.uber.org/zap"
)

func init() {
	fictionbase.SetViperConfig()
}

// Resources struct
type typeChecker struct {
	TypeKey    string `json:"type_key"`
	StorageKey string `json:"storage_key"`
	TimeKey    string `json:"time_key"`
}

var (
	sq       *fictionbase.Sqs
	cw       *fictionbase.Cw
	logger   *zap.Logger
	wg       sync.WaitGroup
	messages []*sqs.Message
	err      error
)

// @TODO more Elegant...
func main() {
	sq = fictionbase.NewSqs()
	cw = fictionbase.NewCw()
	var typeChecker typeChecker

	logger := fictionbase.GetLogger()

	for {
		messages, err = sq.GetFictionbaseMessage()
		if err != nil {
			logger.Error(err.Error())
			return
		}
		// Get All SQS Data
		if len(messages) == 0 {
			return
		}
		for _, m := range messages {
			wg.Add(1)
			go func(m *sqs.Message) {
				defer wg.Done()
				err = json.Unmarshal([]byte(*m.Body), &typeChecker)
				if err != nil {
					logger.Error(err.Error())
					return
				}
				if typeChecker.TypeKey == "fbresource.Resources" {
					SetFbresource(m)
					return
				}
				if typeChecker.TypeKey == "fbhttp.HTTP" {
					SetFbHTTP(m)
					return
				}
			}(m)
		}
		wg.Wait()
	}
}

// SetFbresource Set For fbresource
func SetFbresource(message *sqs.Message) {
	var sqsData fbresource.Resources
	err := json.Unmarshal([]byte(*message.Body), &sqsData)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	// @TODO OtherResources
	dimensionParam := &cloudwatch.Dimension{
		Name:  aws.String("Hostname"),
		Value: aws.String(sqsData.Host.Hostname),
	}
	metricDataParam := &cloudwatch.MetricDatum{
		Dimensions:        []*cloudwatch.Dimension{dimensionParam},
		MetricName:        aws.String("DiskUsage"),
		Timestamp:         &sqsData.TimeKey,
		Unit:              aws.String("Percent"),
		Value:             aws.Float64(sqsData.Disk.UsedPercent),
		StorageResolution: aws.Int64(1),
	}
	putMetricDataInput := &cloudwatch.PutMetricDataInput{
		MetricData: []*cloudwatch.MetricDatum{metricDataParam},
		Namespace:  aws.String("EC2"),
	}
	err = cw.SendCloudWatch(putMetricDataInput)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	sq.DeleteFictionbaseMessage(message)
	if err != nil {
		logger.Error(err.Error())
		return
	}
}

// SetFbHTTP Set For fbhttp
func SetFbHTTP(message *sqs.Message) {
	var sqsData fbhttp.HTTP
	err := json.Unmarshal([]byte(*message.Body), &sqsData)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	dimensionParam := &cloudwatch.Dimension{
		Name:  aws.String("MonitorHTTP"),
		Value: aws.String(sqsData.MonitorHTTP),
	}
	metricDataParam := &cloudwatch.MetricDatum{}
	// statuscode Error
	if sqsData.Status != 0 {
		metricDataParam = &cloudwatch.MetricDatum{
			Dimensions:        []*cloudwatch.Dimension{dimensionParam},
			MetricName:        aws.String("Status"),
			Timestamp:         &sqsData.TimeKey,
			Unit:              aws.String("StatusCode"),
			Value:             aws.Float64(sqsData.Status),
			StorageResolution: aws.Int64(1),
		}
		// ResponseTime Too long
	} else if sqsData.ResponseTime != 0 {
		metricDataParam = &cloudwatch.MetricDatum{
			Dimensions:        []*cloudwatch.Dimension{dimensionParam},
			MetricName:        aws.String("ResponseTime"),
			Timestamp:         &sqsData.TimeKey,
			Unit:              aws.String("ResponseTime"),
			Value:             aws.Float64(sqsData.ResponseTime),
			StorageResolution: aws.Int64(1),
		}
	}
	putMetricDataInput := &cloudwatch.PutMetricDataInput{
		MetricData: []*cloudwatch.MetricDatum{metricDataParam},
		Namespace:  aws.String("EC2"),
	}
	err = cw.SendCloudWatch(putMetricDataInput)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	sq.DeleteFictionbaseMessage(message)
	if err != nil {
		logger.Error(err.Error())
		return
	}
}
