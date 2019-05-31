// shippy-cli-consignment/main.go
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"context"
	pb "github.com/kzozulya1/webpage-word-freq-counter-protobuf/protobuf"
	"google.golang.org/grpc"
)

//Sample data for adding in gRPC service
const (
	sampleData = "SampleData.json"
)

//Parse json file and return corresponding []*PageWordFrequency slice
func parseFile(file string) ([]*pb.PageWordFrequency, error) {
	
	//Allocate slice for 2 elements
	var pageWordFreqs = make([]*pb.PageWordFrequency, 2)
	
	//var pageWordFreq *pb.PageWordFrequency
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, &pageWordFreqs)
	return pageWordFreqs, err
}

//Main routine
func main() {
	// Set up a connection to the server.
	//Is taken from docker-composer.yml file
	address := os.Getenv("GRPC_SERVICE_ADDRESS")
	conn, err := grpc.Dial(address,  grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	//Create new client instance
	client := pb.NewWordFrequencyServiceClient(conn)

	//Read sample data from  SampleData.json file, unmarshal json into []*pb.PageWordFrequency slice
	pageWordFreqs, err := parseFile(sampleData)
	if err != nil {
	 	log.Fatalf("Could not parse file: %v", err)
	}
	
	
	//1. Create / update records
	for i, pageWordFreq := range (pageWordFreqs){
		r, err := client.UpdateOrCreatePageWordFrequency(context.Background(), pageWordFreq)
		if err != nil {
			log.Fatalf("Could not update or create %s, error %v", pageWordFreqs[i].GetPageTitle() ,err)
		}else{
			log.Printf("Successfully processed: %v", r)
		}
	}
	

	//2. Search added records
	
	//Search filter - empty fetchs all records
	searchFilter := "tu"
	searchResult, err := client.GetPageWordFrequency(context.Background(),&pb.GetRequestFilter{PageUrl: searchFilter})
	if err != nil{
		log.Printf("Word freq records fetch error: %v", err)
	}else{
		log.Println("")
		log.Println("~~~~ gRPC Service data ~~~~")
		for i, v := range searchResult.PageWordFreqs {
			log.Println("ID",i," -> ",v)
		}
		log.Println("~~~~ data end ~~~~")
	}

	//3. Remove all records
	removeUrls := []string{"nature","wiki"}
	for _, url := range removeUrls {
		responseRemoved, err := client.RemovePageWordFrequency(context.Background(), &pb.GetRequestFilter{PageUrl: url})
		if err != nil{
			log.Printf("Remove %s error %v",url,err)
		}else{
			log.Printf("Removed %s (%s)", responseRemoved.GetPageWordFreq().GetPageTitle(), responseRemoved.GetPageWordFreq().GetPageUrl())
		}
	}
}