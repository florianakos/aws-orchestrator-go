package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/florianakos/awssh"
	"github.com/florianakos/awshelper"
	//"github.com/florianakos/ec2/credentials"
	"github.com/manifoldco/promptui"
	//awshelper "github.com/florianakos/ec2/helper"
)

var menuItems = []string{
	"\U0001F30D   CloudWatch",
	"\U0001F30D   All",
	"\U0001F5FA    Per Region",
	"\U0001F5FA    Create instances",
	"\U0001F440   Login",
	" \U000025B2   Start",
	" \U000025BC   Stop",
	" \U00002620   Terminate",
	" ⁜   Exit program",
}

var regions = []string{
	"eu-central-1 / Frankfurt",
	"eu-west-1 / Dublin",
	//"eu-west-2 / United Kingdom",
	"eu-west-3 / Paris",
	//"us-east-1 / USA-Virginia",
	//"us-east-2 / USA-Ohio",
	"us-west-1 / California",
	//"us-west-2 / USA-Oregon",
	"ca-central-1 / Toronto",
	"ap-northeast-1 / Tokyo",
	//"ap-northeast-2 / South Korea",
	//"ap-southeast-1 / Singapore",
	//"ap-southeast-2 / Australia",
	//"ap-south-1 / India",
	//"sa-east-1 / Brazil",
	"< return >",
}

type Tags struct {
	Key   string
	Value string
}

// ASYNC STUFF
// Step 1 - MUST BE GLOBAL VARIABLE AVAILABLE TO ALL GOROUTINES
var wg sync.WaitGroup

// Duration is something
type Duration int64

func createAndTagInst(region string, keyPair string, nameTag string, sgID string, amiID string) {
	defer wg.Done() // Step 3
	//time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

	newInstanceID, err := awshelper.CreateInstance(region, keyPair, sgID, amiID)
	//time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = awshelper.TagInstance(region, newInstanceID, nameTag)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Instance launched (ID: %s) and tagged (Name=\"%s\").\n\n", newInstanceID, nameTag)
}

// END OF ASYNC STUFF

func waitAndSee(text string, waitTime int) {
	fmt.Println(text)
	for i := 0; i < waitTime; i++ {
		fmt.Print("*")
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println()
}

func selectKeyPair(region string) (string, error) {
	keyPairs, err := awshelper.GetKeyPairs(region)
	if err != nil {
		return "", err
	}
	if len(keyPairs) != 0 {
		keyPairs = append(keyPairs, "< return >")
		selectedKey, err := promptUserMultiOption("Please select the keypair you want to use to log-in: ", keyPairs)
		if err != nil {
			return "Could not select key", err
		}
		return selectedKey, nil
	} else {
		return "", errors.New("No keypair found")
	}

}

// Function to abstract away the basic functionality of showing a prompt with optons to selectfrom
func promptUserMultiOption(mainText string, options []string) (string, error) {
	promptDetails := promptui.Select{
		Label: mainText,
		Items: options,
		Size:  len(options),
	}
	_, promptResult, err := promptDetails.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "error", err
	}
	return promptResult, nil
}

func convert(num string, err error) (int64, error) {
	val, _ := strconv.ParseInt(num, 10, 64)
	return val, err
}

func promptUserNumber() (int64, error) {
	validate := func(input string) error {
		_, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return errors.New("Invalid number")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    "Enter number of instances to create: ",
		Validate: validate,
		Default:  "1",
	}
	return convert(prompt.Run())
}

func promptUserString() (string, error) {
	validate := func(input string) error {
		if len(input) < 5 {
			return errors.New("The Tag must be at least 5 characters long")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    "Tag (Name=\"\"): ",
		Validate: validate,
		Default:  "FloMeru",
	}
	return prompt.Run()
}

// basic error checking functionality abstracted away into this function to compress code
func checkErr(err error) {
	if err != nil {
		fmt.Printf("Subprompt failed %v\n", err)
		os.Exit(1)
	}
}

// print a welcome message upon starting of the code
func printWelcome() {
	fmt.Printf("************************************************************\n")
	fmt.Printf("*                                                          *\n")
	fmt.Printf("*    Welome to the Go-rchestrator™ written by Flo&Meru     *\n")
	fmt.Printf("*    -------------------------------------------------     *\n")
	fmt.Printf("*                                                          *\n")
	fmt.Printf("*    This is a simple Command-Line Interface in Golang     *\n")
	fmt.Printf("*    and you can interact with it to manage your EC2 &     *\n")
	fmt.Printf("*    other kind of resource(s) in the Amazon AWS Cloud     *\n")
	fmt.Printf("*                                                          *\n")
	fmt.Printf("************************************************************\n")
}

// function to retrieve all instances tagged with Name="FloMeru"
func listInstancesGlobally() {
	baseSession, err := session.NewSessionWithOptions(session.Options{
		Profile: "personal-aws",
		Config: aws.Config{
			Region: aws.String("us-west-2"),
		},
	})
	checkErr(err)
	fmt.Printf("\n_____________________________________________________________________________________________________________\n")
	fmt.Println("\n    IP_address\t\tStatus\t\tRegion\t\tInstanceID\t\tTags\t\t InstanceType")

	counter := 0
	for _, val := range regions {

		currentRegion := strings.Split(val, " / ")[0]
		if currentRegion != "< return >" {

			iterSession := baseSession.Copy(&aws.Config{Region: aws.String(currentRegion)})
			svc := ec2.New(iterSession)
			input := &ec2.DescribeInstancesInput{
				Filters: []*ec2.Filter{
					{
						Name: aws.String("tag:Name"),
						Values: []*string{
							aws.String("FloMeru"),
						},
					},
				},
			}
			res, _ := svc.DescribeInstances(input)
			result := res.Reservations
			if len(result) != 0 {

				for i := 0; i < len(result); i++ {
					instance := *result[i].Instances[0]

					if *instance.State.Name == "running" {
						if len(*instance.PublicIpAddress) < 12 {
							counter++
							fmt.Printf("✔   %v\t\t%v\t\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.PublicIpAddress, *instance.State.Name, currentRegion, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
						} else {
							counter++
							fmt.Printf("✔   %v\t%v\t\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.PublicIpAddress, *instance.State.Name, currentRegion, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
						}
					} else if *instance.State.Name == "shutting-down" || *instance.State.Name == "stopping" {
						fmt.Printf("✗   NO_IP_ADDR\t\t%v\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.State.Name, currentRegion, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
					} else if *instance.State.Name != "terminated" || *instance.State.Name == "pending" || *instance.State.Name == "rebooting" {
						fmt.Printf("✗   NO_IP_ADDR\t\t%v\t\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.State.Name, currentRegion, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
					}
				}

			}
		}
	}
	fmt.Println("______________________________________________________________________________________________________________")
	fmt.Printf("\n>>> You have [ %d ] instances in total!\n\n", counter)
}

func listInstancesInRegion(region string) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	checkErr(err)
	svc := ec2.New(sess)

	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String("FloMeru"),
				},
			},
		},
	}

	res, err := svc.DescribeInstances(input)
	result := res.Reservations

	if len(result) == 0 {
		fmt.Printf("\n>>> You have %d instances in %s!\n", 0, region)
	} else {
		fmt.Println("____________________________________________________________________________________________________________")
		fmt.Println("\n    IP_address\t\tStatus\t\tRegion\t\tInstanceID\t\tTags\t\t InstanceType")

		counter := 0
		zone := strings.Split(region, " / ")[0]

		for i := 0; i < len(result); i++ {
			instance := *result[i].Instances[0]

			if *instance.State.Name == "running" {
				if len(*instance.PublicIpAddress) < 12 {
					counter++
					fmt.Printf("✔   %v\t\t%v\t\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.PublicIpAddress, *instance.State.Name, zone, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
				} else {
					counter++
					fmt.Printf("✔   %v\t%v\t\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.PublicIpAddress, *instance.State.Name, zone, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
				}
			} else if *instance.State.Name == "shutting-down" || *instance.State.Name == "stopping" {
				fmt.Printf("✗   NO_IP_ADDR\t\t%v\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.State.Name, zone, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
			} else if *instance.State.Name != "terminated" || *instance.State.Name == "pending" || *instance.State.Name == "rebooting" {
				fmt.Printf("✗   NO_IP_ADDR\t\t%v\t\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.State.Name, zone, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
			}
		}
		fmt.Println("______________________________________________________________________________________________________________")
		fmt.Printf("\n>>> You have [ %d ] usable instances in %s!\n\n", counter, region)

	} // enf-if
}

func selectIPorID(region string, which string) string {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	checkErr(err)
	svc := ec2.New(sess)

	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String("FloMeru"),
				},
			},
		},
	}

	res, err := svc.DescribeInstances(input)
	result := res.Reservations

	items := make([]string, 0)
	if len(result) == 0 {
		return ""
	} else {
		for i := 0; i < len(result); i++ {
			//instanceArray := result[i].Instances
			instance := *result[i].Instances[0]

			if which == "IP" && *instance.State.Name == "running" {
				items = append(items, *instance.PublicIpAddress)
			} else if which == "ID" && *instance.State.Name != "terminated" && *instance.State.Name != "shutting-down" && *instance.State.Name != "stopping" {
				if *instance.State.Name == "running" {
					items = append(items, *instance.PublicIpAddress+" | "+*instance.InstanceId)
				} else {
					items = append(items, "* NO-IP-ADDR *"+" | "+*instance.InstanceId)
				}
			} // end-if
		} // end-for
	} // enf-if

	if len(items) == 0 {
		return ""
	} else {
		items = append(items, "< return >")
		selectedInstance, _ := promptUserMultiOption("Please select an instance!", items)
		return selectedInstance
	}

}

func main() {

	printWelcome()
	rand.Seed(42)

	// basic infinite loop to execute the prompt over and over until Exit is called...
	for {
		whatToDo, err := promptUserMultiOption("Please select the action below", menuItems)
		checkErr(err)

		switch whatToDo {

		case "🌍   All":
			listInstancesGlobally()
			fmt.Println()
			break

		case "🗺    Per Region":

			subAction, err := promptUserMultiOption("Please select a region first", regions)
			checkErr(err)
			if subAction == "< return >" {
				continue
			}
			listInstancesInRegion(strings.Split(subAction, " / ")[0])
			break

		case "🗺    Create instances":

			selectedRegion, err := promptUserMultiOption("Please select region where you want to create an instance", regions)
			checkErr(err)
			if selectedRegion == "< return >" {
				continue
			}

			region := strings.Split(selectedRegion, " / ")[0]

			keyPair, err := selectKeyPair(region)
			if err != nil {
				fmt.Println("Please generate a new keypair for this region before proceeding!")
				continue
			}
			if keyPair == "< return >" {
				continue
			}

			count, err := promptUserNumber()
			checkErr(err)

			nameTag, err := promptUserString()
			checkErr(err)

			securityGroupID := awshelper.GetSecurityGroupID(region)

			AmazonImageID := awshelper.GetAmazonImageID(region)

			fmt.Printf("\nSetting up %d new instance(s) with params:\n\tregion: %v\n\tTag: Name=\"%v\"\n\tKeyPair: %v\n\tSG-ID: %v\n\tAMI-ID: %v \n\n", count, region, nameTag, keyPair, securityGroupID, AmazonImageID)

			for i := 0; i < int(count); i++ {
				wg.Add(1)                                                                     // Step 2
				go createAndTagInst(region, keyPair, nameTag, securityGroupID, AmazonImageID) // *
			}
			wg.Wait() // Step 4
			break

		case "👀   Login":
			selectedRegion, err := promptUserMultiOption("Please select region for Login", regions)
			checkErr(err)
			if selectedRegion == "< return >" {
				continue
			}

			region := strings.Split(selectedRegion, " / ")[0]

			keyPair, err := selectKeyPair(region)
			if err != nil {
				fmt.Println("Please generate a new keypair for this region before proceeding!")
				continue
			}

			if keyPair == "< return >" {
				continue
			}

			ipAddress := selectIPorID(region, "IP")
			if ipAddress == "" {
				fmt.Printf("\n✗ You don't have any IP addresses active in this region!\n\n")
				continue
			}

			if ipAddress == "< return >" {
				continue
			}

			waitAndSee(fmt.Sprintf("***********************************************\nOpening SSH to IP: [%s] with keypair: [%s].", ipAddress, keyPair), 47)
			awssh.OpenSession(ipAddress, keyPair+".pem")
			fmt.Println()
			break

		case " ☠   Terminate":
			selectedRegion, err := promptUserMultiOption("Please select a region", regions)
			checkErr(err)
			if selectedRegion == "< return >" {
				continue
			}

			region := strings.Split(selectedRegion, " / ")[0]
			temp := selectIPorID(region, "ID")
			if temp == "" {
				fmt.Println("No instance found")
				continue
			}

			if temp == "< return >" {
				continue
			}

			instanceID := strings.Split(temp, " | ")[1]

			//fmt.Println(instanceID)
			err = awshelper.TerminateInstanceByID(region, instanceID)
			if err != nil {
				fmt.Println(err)
			}
			break

		case " ▼   Stop":
			selectedRegion, err := promptUserMultiOption("Please select a region", regions)
			checkErr(err)
			if selectedRegion == "< return >" {
				continue
			}
			region := strings.Split(selectedRegion, " / ")[0]

			temp := selectIPorID(region, "ID")
			if temp == "" {
				fmt.Println("No instance found")
				continue
			}

			if temp == "< return >" {
				continue
			}

			instanceID := strings.Split(temp, " | ")[1]

			err = awshelper.StopInstance(region, instanceID)
			if err != nil {
				fmt.Println(err)
			}
			break

		case " ▲   Start":

			selectedRegion, err := promptUserMultiOption("Please select region", regions)
			checkErr(err)
			if selectedRegion == "< return >" {
				continue
			}

			region := strings.Split(selectedRegion, " / ")[0]
			temp := selectIPorID(region, "ID")
			if temp == "" {
				fmt.Println("No instance found")
				continue
			}

			if temp == "< return >" {
				continue
			}

			instanceID := strings.Split(temp, " | ")[1]

			err = awshelper.StartInstance(region, instanceID)
			if err != nil {
				fmt.Println(err)
			}
			break

		case " ⁜   Exit program":

			fmt.Println("It's sad to see you go... \U0001F622\n")
			os.Exit(0)
			break

		case "🌍   CloudWatch":

			selectedRegion, err := promptUserMultiOption("Please select a region", regions)
			checkErr(err)
			if selectedRegion == "< return >" {
				continue
			}

			region := strings.Split(selectedRegion, " / ")[0]
			temp := selectIPorID(region, "ID")
			if temp == "" {
				fmt.Println("No instance found")
				continue
			}

			if temp == "< return >" {
				continue
			}

			instanceID := strings.Split(temp, " | ")[1]

			data := awshelper.GetCloudWatchMetrics(region, instanceID)
			if data == nil || len(data[0].Values) == 0 {
				fmt.Println("\nNo data found in CloudWatch!\n")
			} else {
				awshelper.RenderGraphs(data)
				awshelper.PlotGraph(region, instanceID, data)
			}
			break
		} // end-switch
	} // end-for
} // end-main
