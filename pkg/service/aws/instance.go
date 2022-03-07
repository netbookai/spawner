package aws

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
)

func WaitTillInstanceRunning(sess *Session, region string, instanceLabelMap map[string]string) error {
	return sess.getEC2Client().WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", constants.NodeNameLabel)),
				Values: aws.StringSlice([]string{instanceLabelMap[constants.NodeNameLabel]}),
			},
		},
	})
}

func WaitTillInstanceTerminated(sess *Session, region string, instanceLabelMap map[string]string) error {

	return sess.getEC2Client().WaitUntilInstanceTerminated(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", constants.NodeNameLabel)),
				Values: aws.StringSlice([]string{instanceLabelMap[constants.NodeNameLabel]}),
			},
		},
	})
}

// Writes data to filename in the current location
func writeFile(filename string, data interface{}) error {
	jsBytArr, err := json.Marshal(data)
	if err != nil {
		fmt.Println("error marshalling to json " + filename)
		return err
	}

	err = os.WriteFile(filename, jsBytArr, 0644)
	if err != nil {
		fmt.Println("error writing file")
		return err
	}

	return nil
}
