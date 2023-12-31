/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mako

import (
	"context"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"cloud.google.com/go/compute/metadata"
	kubeclient "github.com/Yangfisher1/knative-common-pkg/client/injection/kube/client"

	"github.com/Yangfisher1/knative-common-pkg/changeset"
	"github.com/Yangfisher1/knative-common-pkg/controller"
	"github.com/Yangfisher1/knative-common-pkg/injection"
	"github.com/Yangfisher1/knative-common-pkg/test/mako/alerter"
	"github.com/Yangfisher1/knative-common-pkg/test/mako/config"
	"github.com/google/mako/go/quickstore"
	qpb "github.com/google/mako/proto/quickstore/quickstore_go_proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	// sidecarAddress is the address of the Mako sidecar to which we locally
	// write results, and it authenticates and publishes them to Mako after
	// assorted preprocessing.
	sidecarAddress = "localhost:9813"

	// org is the organization name that is used by Github client
	org = "knative"

	// slackUserName is the slack user name that is used by Slack client
	slackUserName = "Knative Testgrid Robot"

	// These token settings are for alerter.
	// If we want to enable the alerter for a benchmark, we need to mount the
	// token to the pod, with the same name and path.
	// See https://github.com/knative/serving/blob/main/test/performance/benchmarks/dataplane-probe/continuous/dataplane-probe.yaml
	tokenFolder     = "/var/secret"
	githubToken     = "github-token"
	slackReadToken  = "slack-read-token"
	slackWriteToken = "slack-write-token"
)

// Client is a wrapper that wraps all Mako related operations
type Client struct {
	Quickstore   *quickstore.Quickstore
	Context      context.Context
	ShutDownFunc func(context.Context)

	benchmarkKey  string
	benchmarkName string
	alerter       *alerter.Alerter
}

// StoreAndHandleResult stores the benchmarking data and handles the result.
func (c *Client) StoreAndHandleResult() error {
	out, err := c.Quickstore.Store()
	return c.alerter.HandleBenchmarkResult(c.benchmarkKey, c.benchmarkName, out, err)
}

var tagEscaper = strings.NewReplacer("+", "-", "\t", "_", " ", "_")

// EscapeTag replaces characters that Mako doesn't accept with ones it does.
func EscapeTag(tag string) string {
	return tagEscaper.Replace(tag)
}

// SetupHelper sets up the mako client for the provided benchmarkKey.
// It will add a few common tags and allow each benchmark to add custom tags as well.
// It returns the mako client handle to store metrics, a method to close the connection
// to mako server once done and error in case of failures.
func SetupHelper(ctx context.Context, benchmarkKey *string, benchmarkName *string, extraTags ...string) (*Client, error) {
	tags := append(config.MustGetTags(), extraTags...)
	// Get the commit of the benchmarks
	commitID := changeset.Get()
	if commitID == changeset.Unknown {
		log.Println("Cannot find commit ID")
	}

	// Setup a deployment informer, so that we can use the lister to track
	// desired and available pod counts.
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	ctx, informers := injection.Default.SetupInformers(ctx, cfg)
	if err := controller.StartInformers(ctx.Done(), informers...); err != nil {
		return nil, err
	}

	// Get the Kubernetes version from the API server.
	kc := kubeclient.Get(ctx)
	version, err := kc.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}

	// Determine the number of Kubernetes nodes through the kubernetes client.
	nodes, err := kc.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	tags = append(tags, "nodes="+strconv.Itoa(len(nodes.Items)))

	// Decorate GCP metadata as tags (when we're running on GCP).
	if projectID, err := metadata.ProjectID(); err != nil {
		log.Print("GCP project ID is not available: ", err)
	} else {
		tags = append(tags, "project-id="+EscapeTag(projectID))
	}
	if zone, err := metadata.Zone(); err != nil {
		log.Print("GCP zone is not available: ", err)
	} else {
		tags = append(tags, "zone="+EscapeTag(zone))
	}
	if machineType, err := metadata.Get("instance/machine-type"); err != nil {
		log.Print("GCP machine type is not available: ", err)
	} else if parts := strings.Split(machineType, "/"); len(parts) != 4 {
		tags = append(tags, "instanceType="+EscapeTag(parts[3]))
	}
	tags = append(tags,
		"commit="+commitID,
		"kubernetes="+EscapeTag(version.String()),
		"goversion="+EscapeTag(runtime.Version()),
	)
	log.Printf("The tags for this run are: %+v", tags)
	// Create a new Quickstore that connects to the microservice
	qs, qclose, err := quickstore.NewAtAddress(ctx, &qpb.QuickstoreInput{
		BenchmarkKey: benchmarkKey,
		Tags:         tags,
	}, sidecarAddress)
	if err != nil {
		return nil, err
	}

	// Create a new Alerter that alerts for performance regressions
	alerter := &alerter.Alerter{}
	alerter.SetupGitHub(
		org,
		config.GetRepository(),
		tokenPath(githubToken),
	)
	alerter.SetupSlack(
		slackUserName,
		tokenPath(slackReadToken),
		tokenPath(slackWriteToken),
		config.GetSlackChannels(*benchmarkName),
	)

	client := &Client{
		Quickstore:    qs,
		Context:       ctx,
		ShutDownFunc:  qclose,
		alerter:       alerter,
		benchmarkKey:  *benchmarkKey,
		benchmarkName: *benchmarkName,
	}

	return client, nil
}

func Setup(ctx context.Context, extraTags ...string) (*Client, error) {
	bench := config.MustGetBenchmark()
	return SetupHelper(ctx, bench.BenchmarkKey, bench.BenchmarkName, extraTags...)
}

func tokenPath(token string) string {
	return filepath.Join(tokenFolder, token)
}
