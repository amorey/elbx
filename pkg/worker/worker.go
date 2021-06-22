package worker

import (
	"context"
	"fmt"
	"sync"
	
	"github.com/rs/zerolog/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
        "github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
        ExcludeFromELBsLabelKey = "node.kubernetes.io~1exclude-from-external-load-balancers"
        ExcludeFromELBsLabelVal = "elbx"
)

type Worker struct {
	k8sClient *kubernetes.Clientset
	ec2Client *ec2.Client
}

func (w *Worker) processEventBridgeEvent(ctx context.Context, instanceId string) error {
	log.Debug().Msg(fmt.Sprintf("Processing event (%s)", instanceId))

	// get instance details
	params := &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name: aws.String("instance-id"),
				Values: []string{instanceId},
			},
		},
	}
	
	response, err := w.ec2Client.DescribeInstances(ctx, params)
	if err != nil {
		return err
	} else if len(response.Reservations) == 0 || len(response.Reservations[0].Instances) == 0 {
		log.Error().Msg(fmt.Sprintf("No EC2 instance found: `%s`", instanceId))
		return nil
	}
	
	// get node name
	nodeName := response.Reservations[0].Instances[0].PrivateDnsName

	// add elbx label
	labelPatch := fmt.Sprintf(`[{"op":"add","path":"/metadata/labels/%s","value":"%s" }]`, ExcludeFromELBsLabelKey, ExcludeFromELBsLabelVal)
        _, err = w.k8sClient.CoreV1().Nodes().Patch(ctx, *nodeName, k8stypes.JSONPatchType, []byte(labelPatch), metav1.PatchOptions{})
        if err != nil {
		return err
        }

	log.Info().Msg(fmt.Sprintf("Added elbx label to node: %s", *nodeName))
	
	return nil
}

func (w *Worker) WatchForEventBridgeEvents(ctx context.Context, commsChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

Loop:
        for {
                select {
                case <-ctx.Done():
                        break Loop
		case instanceId := <-commsChan:
			err := w.processEventBridgeEvent(ctx, instanceId)
			if err != nil {
				log.Error().Msg(err.Error())
			}
                }
        }

        log.Debug().Msg("WatchForEventBridgeEvents exiting")
}

func New() (*Worker, error) {
	// init kubernetes clientset (using in-cluster config)
	k8scfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	k8sClient, err := kubernetes.NewForConfig(k8scfg)
	if err != nil {
		return nil, err
	}

	// init ec2 client
	awscfg, err := config.LoadDefaultConfig(context.Background())
        if err != nil {
                return nil, err
        }

	ec2Client := ec2.NewFromConfig(awscfg)

	// init worker instance
	w := Worker{
		k8sClient: k8sClient,
		ec2Client: ec2Client,
	}
	
	return &w, nil
}
