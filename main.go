package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/camunda-cloud/zeebe/clients/go/pkg/entities"
	"github.com/camunda-cloud/zeebe/clients/go/pkg/worker"
	"github.com/camunda-cloud/zeebe/clients/go/pkg/zbc"
)

const (
	gatewayAddress = "0.0.0.0:26500"

	// When this is a large number (e.g. 5000), the process experiences errors.
	dataCount = 5000
)

func main() {
	client := zeebeClient()

	client.NewJobWorker().JobType("first_data_loader").Handler(firstDataLoader).Open()
	client.NewJobWorker().JobType("second_data_processor").Handler(secondParallelMultiInstance).Open()
	client.NewJobWorker().JobType("third_data_printer").Handler(thirdDataPrinter).Open()

	resp, err := client.NewCreateInstanceCommand().BPMNProcessId("my_process").LatestVersion().Send(context.Background())
	must(err)

	log.Printf("started process: %d", resp.ProcessInstanceKey)

	// Sleep forever.
	select {}
}

type dataWrapper struct {
	Data []data `json:"inputCollection"`
}

type data struct {
	Name      string    `json:"name"`
	Index     int       `json:"index"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func firstDataLoader(client worker.JobClient, job entities.Job) {
	log.Println("first job handler called")

	data := generateData()

	cmd, err := client.NewCompleteJobCommand().JobKey(job.GetKey()).VariablesFromObject(data)
	must(err)

	_, err = cmd.Send(context.Background())
	must(err)
}

func generateData() dataWrapper {
	dw := dataWrapper{}

	for i := 0; i < dataCount; i++ {
		data := data{
			Name:      "name_" + strconv.Itoa(i),
			Index:     i,
			ExpiresAt: time.Now().Add(time.Second * time.Duration(i)),
		}

		dw.Data = append(dw.Data, data)
	}

	return dw
}

func secondParallelMultiInstance(client worker.JobClient, job entities.Job) {
	log.Println("second job handler called")

	_, err := client.NewCompleteJobCommand().JobKey(job.GetKey()).Send(context.Background())
	must(err)
}

func thirdDataPrinter(client worker.JobClient, job entities.Job) {
	log.Println("third job handler called")

	log.Printf("%+v", job.GetVariables())

	_, err := client.NewCompleteJobCommand().JobKey(job.GetKey()).Send(context.Background())
	must(err)
}

func zeebeClient() zbc.Client {
	client, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:         gatewayAddress,
		UsePlaintextConnection: true,
	})
	must(err)

	return client
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
