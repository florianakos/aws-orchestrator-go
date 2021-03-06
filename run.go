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
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/florianakos/awsutils"
	"github.com/manifoldco/promptui"
)

var menuItems = []string{
	"🌍   EC2 - list instances by region",
	"✅   EC2 - create instance(s)",
	"💻   EC2 - ssh login",
	"🛠    EC2 - instance maintenance",
	"📊   CW - get instance metrics",
	"❌   Exit program",
}

var ec2Actions = []string{
	"▶️  start instance",
	"⏸  stop instance",
	"⏹  terminate instance",
	"< return >",
}

var regions = []string{
	"eu-central-1 / Frankfurt-DE",
	"eu-west-1 / Dublin-IR",
	"eu-west-3 / Paris-FR",
	"us-west-1 / California-USA",
	"ca-central-1 / Toronto-CA",
	"ap-northeast-1 / Tokyo-JP",
	"< return >",
}

// Tags struct for holding Tagging information
type Tags struct {
	Key   string
	Value string
}

// wg helps with parallel starting of EC2 instances
var wg sync.WaitGroup

func createAndTagInst(region string, keyPair string, nameTag string, sgID string, amiID string) {
	defer wg.Done() // Step 3

	newInstanceID, err := awsutils.CreateInstance(region, keyPair, sgID, amiID)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = awsutils.TagInstance(region, newInstanceID, nameTag)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Instance launched (ID: %s) and tagged (Name=\"%s\").\n\n", newInstanceID, nameTag)
}

func waitAndSee(text string, waitTime int) {
	fmt.Println(text)
	for i := 0; i < waitTime; i++ {
		fmt.Print("*")
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println()
}

func selectKeyPair(region string) (string, error) {
	keyPairs, err := awsutils.GetKeyPairs(region)
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
	}
	return "", errors.New("no keypair found")
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

func promptForNumber() (int64, error) {
	validate := func(input string) error {
		_, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return errors.New("Invalid number")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    "Number of instances you need: ",
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
		Default:  "FLRNKS",
	}
	return prompt.Run()
}

func getInstancesByState(region string, states ...string) []*ec2.Reservation {
	// create session using AWS profile for my personal account
	svc, err := awsutils.GetEC2ServiceClient(region)
	checkErr(err)

	// construct filter from passed arguments
	filter := []*string{}
	for _, v := range states {
		filter = append(filter, aws.String(v))
	}

	// Describe EC2 instances matching the state filter
	res, _ := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{{
			Name:   aws.String("instance-state-name"),
			Values: filter,
		}},
	})
	return res.Reservations
}

func promptForUsername() (string, error) {
	validate := func(input string) error {
		if len(input) < 1 {
			return errors.New("username cannot be empty")
		}
		return nil
	}
	prompt := promptui.Prompt{
		Label:    "Login Username (Username=\"\"): ",
		Validate: validate,
		Default:  "ec2-user",
	}
	return prompt.Run()
}

func listEC2InstanceByRegion(region string) {
	// get list of instances in region wiwht given states
	instances := getInstancesByState(region, "running", "pending", "stopping", "stopped")

	// parse results and print accordingly
	if len(instances) == 0 {
		fmt.Printf("\n>>> You have %d instances in %s!\n", 0, region)
	} else {
		fmt.Println("_____________________________________________________________________________________________________________")
		fmt.Println("\n    IP_address\t\tStatus\t\tRegion\t\tInstanceID\t\tTags\t\t InstanceType")

		counter := 0
		zone := strings.Split(region, " / ")[0]
		for i := 0; i < len(instances); i++ {
			instance := *instances[i].Instances[0]
			if *instance.State.Name == "running" {
				counter++
				if len(*instance.PublicIpAddress) < 12 {
					fmt.Printf("✔   %v\t\t%v\t\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.PublicIpAddress, *instance.State.Name, zone, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
				} else {
					fmt.Printf("✔   %v\t%v\t\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.PublicIpAddress, *instance.State.Name, zone, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
				}
			} else if *instance.State.Name == "shutting-down" || *instance.State.Name == "stopping" {
				fmt.Printf("✗   NO_IP_ADDR\t\t%v\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.State.Name, zone, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
			} else if *instance.State.Name != "terminated" || *instance.State.Name == "pending" || *instance.State.Name == "rebooting" {
				fmt.Printf("✗   NO_IP_ADDR\t\t%v\t\t%v\t%v\t%v=\"%v\"\t %v\n", *instance.State.Name, zone, *instance.InstanceId, *instance.Tags[0].Key, *instance.Tags[0].Value, *instance.InstanceType)
			}
		}
		fmt.Println("_____________________________________________________________________________________________________________")
		fmt.Printf("\n>>> You have [ %d ] running instances in %s!\n\n", counter, region)

	}
}

func selectEC2InstanceWithStates(region string, kind string, states ...string) string {
	// get list of instances in region wiwht given states
	instances := getInstancesByState(region, states...)

	// parse results obtained from DescribeEC2 operation
	items := make([]string, 0)
	if len(instances) != 0 {
		for i := 0; i < len(instances); i++ {
			instance := *instances[i].Instances[0]

			if kind == "IP" && *instance.State.Name == "running" {
				items = append(items, *instance.PublicIpAddress)
			} else if kind == "ID" && *instance.State.Name != "terminated" && *instance.State.Name != "shutting-down" {
				if *instance.State.Name == "running" {
					items = append(items, fmt.Sprintf("%s | %s", *instance.PublicIpAddress, *instance.InstanceId))
				} else {
					items = append(items, fmt.Sprintf("* NO-IP-ADDR * | %s", *instance.InstanceId))
				}
			}
		}
		if len(items) != 0 {
			items = append(items, "< return >")
			selectedInstance, _ := promptUserMultiOption("Please select an instance!", items)
			return selectedInstance
		}
	}
	return ""
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
	fmt.Printf("*    Welome to the AWS-GO-Orchestrator™ by: [flrnks] ©     *\n")
	fmt.Printf("*    -------------------------------------------------     *\n")
	fmt.Printf("*                                                          *\n")
	fmt.Printf("*    This is a simple Command-Line Interface in Golang     *\n")
	fmt.Printf("*    and you can interact with it to manage your EC2 &     *\n")
	fmt.Printf("*    other kind of resource(s) in the Amazon AWS Cloud     *\n")
	fmt.Printf("*                                                          *\n")
	fmt.Printf("************************************************************\n")
}

func main() {

	printWelcome()

	rand.Seed(42)

	// basic infinite loop to execute the prompt over and over until Exit is called...
	for {
		action, err := promptUserMultiOption("Select an action", menuItems)
		checkErr(err)
		switch action {

		case "🌍   EC2 - list instances by region":

			region, err := promptUserMultiOption("Please select a region first", regions)
			checkErr(err)
			if region == "< return >" {
				continue
			}
			listEC2InstanceByRegion(strings.Split(region, " / ")[0])
			break

		case "✅   EC2 - create instance(s)":

			selectedRegion, err := promptUserMultiOption("Please select region where you want to create an instance", regions)
			checkErr(err)
			if selectedRegion == "< return >" {
				continue
			}
			region := strings.Split(selectedRegion, " / ")[0]

			keyPair, err := selectKeyPair(region)
			fmt.Println(keyPair)
			if err != nil {
				fmt.Println("Please generate a new keypair for this region before proceeding!")
				continue
			}
			if keyPair == "< return >" {
				continue
			}

			count, err := promptForNumber()
			checkErr(err)

			nameTag, err := promptUserString()
			checkErr(err)

			securityGroupID := awsutils.GetSecurityGroupID(region)
			AmazonImageID := awsutils.GetAmazonImageID(region)

			fmt.Printf("\nSetting up %d new instance(s) with params:\n\tregion: %v\n\tTag: Name=\"%v\"\n\tKeyPair: %v\n\tSG-ID: %v\n\tAMI-ID: %v \n\n",
				count, region, nameTag, keyPair, securityGroupID, AmazonImageID)

			for i := 0; i < int(count); i++ {
				wg.Add(1)
				go createAndTagInst(region, keyPair, nameTag, securityGroupID, AmazonImageID)
			}
			wg.Wait()
			break

		case "💻   EC2 - ssh login":

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

			ipAddress := selectEC2InstanceWithStates(region, "IP", "running")
			if ipAddress == "" {
				fmt.Printf("\n✗ You don't have any IP addresses active in this region!\n\n")
				continue
			} else if ipAddress == "< return >" {
				continue
			}

			username, err := promptForUsername()
			checkErr(err)

			waitAndSee(fmt.Sprintf("***********************************************\nOpening SSH to IP: [%s] with keypair: [%s].", ipAddress, keyPair), 47)
			awsutils.OpenSession(username, ipAddress, keyPair+".pem")
			fmt.Println()
			break

		case "🛠    EC2 - instance maintenance":

			selectedRegion, err := promptUserMultiOption("Please select region", regions)
			checkErr(err)
			if selectedRegion == "< return >" {
				continue
			}
			region := strings.Split(selectedRegion, " / ")[0]
			temp := selectEC2InstanceWithStates(region, "ID", "running", "stopped", "stopping", "pending")
			if temp == "" {
				fmt.Printf("\nNo instance found in '%s' region!\n\n", region)
				continue
			} else if temp == "< return >" {
				continue
			}
			instanceID := strings.Split(temp, " | ")[1]
			selectedAction, err := promptUserMultiOption("Please select action on selected instance:", ec2Actions)
			checkErr(err)
			if selectedAction == "< return >" {
				continue
			}
			switch selectedAction {
			case "▶️  start instance":
				err = awsutils.StartInstance(region, instanceID)
				break
			case "⏸  stop instance":
				err = awsutils.StopInstance(region, instanceID)
				break
			case "⏹  terminate instance":
				err = awsutils.TerminateInstanceByID(region, instanceID)
				break
			}
			if err != nil {
				fmt.Println(err)
			}
			break

		case "❌   Exit program":

			fmt.Println("\nIt's sad to see you go... 😢")
			os.Exit(0)
			break

		case "📊   CW - get instance metrics":

			selectedRegion, err := promptUserMultiOption("Please select a region", regions)
			checkErr(err)
			if selectedRegion == "< return >" {
				continue
			}
			region := strings.Split(selectedRegion, " / ")[0]

			instance := selectEC2InstanceWithStates(region, "ID", "running", "stopped")
			if instance == "" {
				fmt.Println("No instance found")
				continue
			} else if instance == "< return >" {
				continue
			}

			instanceID := strings.Split(instance, " | ")[1]
			data := awsutils.GetCloudWatchMetrics(region, instanceID)
			if data == nil || len(data[0].Values) == 0 {
				fmt.Printf("\nNo data found in CloudWatch!\n")
			} else {
				awsutils.RenderGraphs(data)
				awsutils.PlotGraph(region, instanceID, data)
			}
			break
		}
	}
}
