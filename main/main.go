package main

import (
	"context"
	"flag"
	"fmt"
	gohttp "net/http"
	"time"
)

type SampleDoer struct {
	httpClient *gohttp.Client
}
func(doer *SampleDoer) Do(httpReq *gohttp.Request) (*gohttp.Response, error){
	return doer.Do(httpReq)
}

func main() {
	userName := "admin"
	password := "admin"
	//servicePtr := flag.String("s", "ziti-influx", "Name of Service")
	dbPtr := flag.String("db", "test", "Name of db")
	flag.Parse()

	//create a new "Doer" - in this case it is a simple struct which implements "Do"
	sampleDoer := &SampleDoer{} //

	// Create a new client using an InfluxDB server base URL and an authentication token
	// For authentication token supply a string in the form: "username:password" as a token. Set empty value for an unauthenticated server
	client:= influxdb2.NewClientWithDoer(sampleDoer, fmt.Sprintf("%s:%s",userName, password), influxdb2.DefaultOptions().SetBatchSize(20))

	// Get the blocking write client
	// Supply a string in the form database/retention-policy as a bucket. Skip retention policy for the default one, use just a database name (without the slash character)
	// Org name is not used
	bucket := *dbPtr + "/autogen"
	writeAPI := client.WriteAPIBlocking("", bucket)
	// create point using full params constructor
	p := influxdb2.NewPoint("stat",
		map[string]string{"unit": "temperature"},
		map[string]interface{}{"avg": 24.5, "max": 45},
		time.Now())
	// Write data
	err := writeAPI.WritePoint(context.Background(), p)
	if err != nil {
		fmt.Printf("Write error: %s\n", err.Error())
	}

	// Get query client. Org name is not used
	queryAPI := client.QueryAPI("")
	// Supply string in a form database/retention-policy as a bucket. Skip retention policy for the default one, use just a database name (without the slash character)
	result, err := queryAPI.Query(context.Background(), `from(bucket:"`+ bucket +`")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)
	if err == nil {
		for result.Next() {
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			fmt.Printf("row: %s\n", result.Record().String())
		}
		if result.Err() != nil {
			fmt.Printf("Query error: %s\n", result.Err().Error())
		}
	} else {
		fmt.Printf("Query error: %s\n", err.Error())
	}
	// Close client
	client.Close()
}