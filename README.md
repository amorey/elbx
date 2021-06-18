# ELBX (External Load Balancer Excluder)

The `elbx` app monitors an AWS SQS queue for ASG Termination events and EC2 Spot Instance Interruption notices and, in response, adds a "node.kubernetes.io/exclude-from-external-load-balancers: elbx" label to the targeted kubernetes nodes which causes the kubernetes control plane to remove said node from the cluster's external load balancers. The `elbx` app is designed to run in a pod inside a kubernetes cluster. The AWS and Kubernetes configuration files for the app are located in the [manifests/](manifests/) directory.

## Installation and Configuration

To use the `elbx` app, first you must create an AWS SQS queue and set up the event listeners for ASG Termination events and EC2 Spot Instance Interruption notices. Then you can deploy the app to your kubernetes cluster.

### AWS Infrastructure Setup

To create a CloudFormation stack that listens for ASG Termination events and EC2 Spot Instance Interruption notices and sends events to an SQS queue that is monitored by the `elbx` app you can use the [manifests/cf-stack.yaml](manifests/cf-stack.yaml) file located in this project. Make sure to choose a name for your stack and to replace the `ASG1`, `ASG2` parameter values with the names of your cluster's Auto Scaling Groups:

```sh
$ aws cloudformation create-stack \
  --stack-name REPLACEME \
  --template-body file://cf-stack.yaml \
  --parameters \
      ParameterKey=ASG1,ParameterValue=REPLACEME-GROUP-1 \
      ParameterKey=ASG2,ParameterValue=REPLACEME-GROUP-2
```

### Kubernetes Installation

The `elbx` app requires the following AWS permissions:

  * ec2:DescribeInstances
  * sqs:ReceiveMessage
  * sqs:DeleteMessage

You can use the [manifests/iam-policy.yaml](manifests/iam-policy.yaml) file located in this project to create an IAM policy for a user or a role to use. Then you can download and modify the [manifests/k8s-elbx.yaml](manifests/k8s-elbx.yaml) kubernetes yaml file to add the AWS credentials you would like the `elbx` app to use. Once you have configured your yaml file you can deploy it to your kubernetes cluster:

```sh
$ kubectl apply -f k8s-elbx.yaml
```

### Configuration Options

| ENV Variable          | Datatype | Description     | Default |
| --------------------- | -------- | --------------- | ------- |
| DEBUG                 | bool     | Debug mode      | false   |
| QUEUE_URL             | string   | SQS queue URL   | ""      |
| AWS_ACCESS_KEY_ID     | string   | AWS credentials | ""      |
| AWS_SECRET_ACCESS_KEY | string   | AWS credentials | ""      |
| AWS_DEFAULT_REGION    | string   | AWS config      | ""      |

## Development

To run the app in development you can use the `go run` command:

```sh
$ go run ./cmd/main \
  -debug \
  -queue-url="https://sqs.us-east-1.amazonaws.com/REPLACEME"
```

## Build Docker image

Build image:

```sh
$ docker build -t elbx:latest -f build/package/Dockerfile .
```

Start container (with debug enabled):
```sh
$ docker run --rm elbx:latest \
  -debug \
  -queue-url="https://sqs.us-east-1.amazonaws.com/REPLACEME"
```
