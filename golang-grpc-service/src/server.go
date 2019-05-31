//main.go
package main
import (
	"context"
	"log"
	"net"
	"sync"
	"errors"
	"os"
	"strings"
	// Import the generated protobuf code
	pb "github.com/kzozulya1/webpage-word-freq-counter-protobuf/protobuf"
	"google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)
//Interface for manage word freq records
type IRepository interface {
	Create(*pb.PageWordFrequency) (*pb.PageWordFrequency, error)
	Update(*pb.PageWordFrequency) (*pb.PageWordFrequency, error)
	GetAll(*pb.GetRequestFilter) ([]*pb.PageWordFrequency) 
	Remove(string) (*pb.PageWordFrequency, error)
}

// Repository simulates the use of a datastore
// of some kind.
type Repository struct {
	mu sync.RWMutex
	pageWordFrequencyRecords []*pb.PageWordFrequency
}

//Remove word freq record
func (repo *Repository) Remove(pageUrl string)  (*pb.PageWordFrequency,error) {
	pageWordFrequency := new(pb.PageWordFrequency)

	for i := len(repo.pageWordFrequencyRecords) - 1; i >= 0; i-- {
		if strings.Contains(repo.pageWordFrequencyRecords[i].GetPageUrl(),pageUrl){
			
			//Few boiler plate code:)
			pageWordFrequency = repo.pageWordFrequencyRecords[i]
			// pageWordFrequency.PageUrl = repo.pageWordFrequencyRecords[i].GetPageUrl()
			// pageWordFrequency.PageTitle = repo.pageWordFrequencyRecords[i].GetPageTitle()

			repo.mu.Lock()
			repo.pageWordFrequencyRecords = append(repo.pageWordFrequencyRecords[:i], repo.pageWordFrequencyRecords[i+1:]...)
			repo.mu.Unlock()
		}
	}
	return pageWordFrequency, nil
}

// Create new word freq record
func (repo *Repository) Create(pageWordFreq *pb.PageWordFrequency) (*pb.PageWordFrequency, error) {
	repo.mu.Lock()
	repo.pageWordFrequencyRecords = append(repo.pageWordFrequencyRecords, pageWordFreq)
	repo.mu.Unlock()
	return pageWordFreq, nil
}

//Update freq records
func (repo *Repository) Update(pageWordFreq *pb.PageWordFrequency) (*pb.PageWordFrequency, error) {
	for i := len(repo.pageWordFrequencyRecords) - 1; i >= 0; i-- {
		if repo.pageWordFrequencyRecords[i].GetPageUrl() == pageWordFreq.GetPageUrl() {
			repo.mu.Lock()
			repo.pageWordFrequencyRecords[i] = pageWordFreq
			repo.mu.Unlock()
			return pageWordFreq, nil
		}
	}

	return nil, errors.New("Can't find word freq record " + pageWordFreq.GetPageUrl())
}

// Find word freq records.
//Filter is available
func (repo *Repository) GetAll(req *pb.GetRequestFilter) []*pb.PageWordFrequency {
	records := repo.pageWordFrequencyRecords
	
	//Filter data by page url
	if (req.GetPageUrl() != ""){
		for i := len(records) - 1; i >= 0; i-- {
			//Remove item if it doesn't meet filter condition
            if ! strings.Contains(records[i].GetPageUrl(),req.GetPageUrl()){
				records = append(records[:i], records[i+1:]...)
			}
		}
	}
	//Filter records by word
	if (req.GetWord() != ""){
		for i := len(records) - 1; i >= 0; i-- {
			words := records[i].GetWords()
			if  len(words) > 0 { 
				for j := len(words) - 1; j >= 0; j-- {
					if !strings.Contains(words[j].GetValue(),req.GetWord()){
						words = append(words[:j], words[j+1:]...)
					}
				}
				records[i].Words = words
			}
		}
	}
	return records
}

// Service  implement all of the methods to satisfy the service
// we defined in our protobuf definition. 
type service struct {
	repo IRepository
}

// Update records, or create if it doesn't exist
func (s *service) UpdateOrCreatePageWordFrequency(ctx context.Context, req *pb.PageWordFrequency) (*pb.Response, error){
	
	//Try to update
	updated, err := s.repo.Update(req)
	if err == nil{
		return &pb.Response{Updated: true, PageWordFreq: updated}, nil  
	}else{
		//Create new one
		created, err := s.repo.Create(req)
		if err != nil {
			return nil, err 
		}
		return &pb.Response{Created: true, PageWordFreq: created}, nil  
	}
}


//Get all word freq records, appy filter pb.GetRequestFilter.PageUrl / pb.GetRequestFilter.Word
func (s *service) GetPageWordFrequency(ctx context.Context, req *pb.GetRequestFilter) (*pb.Response, error){
    allRecords := s.repo.GetAll(req)
	return &pb.Response{PageWordFreqs: allRecords}, nil
}

//Remove record by pb.GetRequestFilter.PageUrl
func (s *service) RemovePageWordFrequency(ctx context.Context, req *pb.GetRequestFilter) (*pb.Response, error){
	pageWordFreq, err := s.repo.Remove(req.GetPageUrl())
	response := &pb.Response{Removed: true, PageWordFreq: pageWordFreq} 
	 if err != nil{
	 	response.Removed = false
	 }
	return response, err
}

//Main routine
func main() {
	//Create empty repository
	repo := &Repository{}

	// Set-up gRPC server.
	//Port is set in end in docker-compose file
	port :=  os.Getenv("SERVICE_PORT")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err) 
	}
	s := grpc.NewServer()

	// Register our service with the gRPC server
	pb.RegisterWordFrequencyServiceServer(s, &service{repo})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	log.Println("Running on port:", port) 
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}