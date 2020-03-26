//snippet-comment:[These are tags for the AWS doc team's sample catalog. Do not remove.]
//snippet-sourceauthor:[Doug-AWS]
//snippet-sourcedescription:[Retrieves a list of published AWS CloudWatch metrics.]
//snippet-keyword:[AWS CloudWatch]
//snippet-keyword:[ListMetrics function]
//snippet-keyword:[Go]
//snippet-service:[cloudwatch]
//snippet-sourcetype:[full-example]
//snippet-sourcedate:[2018-03-16]
/*
   Copyright 2010-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
   This file is licensed under the Apache License, Version 2.0 (the "License").
   You may not use this file except in compliance with the License. A copy of
   the License is located at
    http://aws.amazon.com/apache2.0/
   This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
   CONDITIONS OF ANY KIND, either express or implied. See the License for the
   specific language governing permissions and limitations under the License.
*/

package main

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/cloudwatch"

    "fmt"
    "strconv"
    "time"
    "log"
    _"os"
)

var counter int64 = 0

func buildMetricDataQuery(name, instance string) *cloudwatch.MetricDataQuery {
	counter++
	instance = "i-08b39fb1d301c6cf9"
	return &cloudwatch.MetricDataQuery{
		Id: aws.String("id" + strconv.FormatInt(counter, 10)),
		MetricStat: &cloudwatch.MetricStat{
			Period:     aws.Int64(60),
			Stat: aws.String("Average"),
			Metric: &cloudwatch.Metric{
				MetricName: aws.String(name),
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  aws.String("InstanceId"),
						Value: aws.String(instance),
					},
				},
				Namespace:  aws.String("AWS/EC2"),
			},
		},
	}
}

func main() {
   
    // Initialize a session that the SDK uses to load
    // credentials from the shared credentials file ~/.aws/credentials
    // and configuration from the shared configuration file ~/.aws/config.
    sess, err := session.NewSession(&aws.Config{ Region: aws.String("eu-central-1")}, )

    // Create CloudWatch client
    cw := cloudwatch.New(sess)

// The new way: use GetMetricData
	dataInput := &cloudwatch.GetMetricDataInput{
		StartTime:  aws.Time(time.Now().Add(-60* time.Minute)),
		EndTime:    aws.Time(time.Now()),
		// Limit of 100 of these
		MetricDataQueries: []*cloudwatch.MetricDataQuery{
      	// you can include up to 100 combinations of metric name and instance
			buildMetricDataQuery("CPUUtilization", "i-08b39fb1d301c6cf9"),
			buildMetricDataQuery("NetworkPacketsIn", "i-08b39fb1d301c6cf9"),
			buildMetricDataQuery("NetworkPacketsOut", "i-08b39fb1d301c6cf9"),
		},
	}
	dataOutput, err := cw.GetMetricData(dataInput)
	if err != nil {
		log.Fatal("error GetMetricStatistics: ", err)
	}
	fmt.Printf("GetMetricData: one API call can fetch up to 100 metrics (we have 8 x N metrics, where N is the number of EC2 instances): %+v\n", dataOutput)


/*
	// Our current approach: use GetMetricStatistics
	input := &cloudwatch.GetMetricStatisticsInput{
		Dimensions: []*cloudwatch.Dimension{
			&cloudwatch.Dimension{
				Name:  aws.String("InstanceId"),
				Value: aws.String("i-08b39fb1d301c6cf9"),
			},
		},
		MetricName: aws.String("CPUUtilization"),
		Namespace:  aws.String("AWS/EC2"),
		Period:     aws.Int64(60),
		StartTime:  aws.Time(time.Now().Add(-60 * time.Minute)),
		EndTime:    aws.Time(time.Now()),
		Statistics: []*string{aws.String("Average")},
	}
	output, err := cw.GetMetricStatistics(input)
	if err != nil {
		log.Fatal("error GetMetricStatistics: ", err)
	}
	fmt.Printf("GetMetricStatistics: one API call gets us one metric: %+v\n", output)*/
}