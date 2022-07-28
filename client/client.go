package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	calculatorpb "example.com/ask/proto"
	"google.golang.org/grpc"
)

func main() {

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)

	//Unary API function - Greet
	Sum(c)

	//Server-Side Streaming function - GreetManyTimes
	PrimeNumbers(c)

	//Client-Side Straming function - LongGreet
	ComputeAverage(c)

	//bi-directional stream
	FindMaxNumber(c)

}

func Sum(c calculatorpb.CalculatorServiceClient) {

	fmt.Println("Strating to do a unary GRPC....")

	req := calculatorpb.SumRequest{
		Num1: 2,
		Num2: 3,
	}

	resp, err := c.Sum(context.Background(), &req)

	if err != nil {
		log.Fatalf("error while calling Sum grpc unary call: %v", err)
	}

	log.Printf("Response from Sum Unary Call : %v", resp.Result)

}

//PrimeNumbers(ctx context.Context, in *PrimeRequest, opts ...grpc.CallOption) (CalculatorService_PrimeNumbersClient, error)

func PrimeNumbers(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Staring ServerSide GRPC streaming ....")
	req := calculatorpb.PrimeRequest{
		Num: 7,
	}

	respStream, err := c.PrimeNumbers(context.Background(), &req)
	if err != nil {
		log.Fatalf("error while calling PrimeNumbers grpc server side stream call: %v", err)
	}

	for {
		msg, err := respStream.Recv()
		if err == io.EOF {
			//we have reached to the end of the file
			break
		}

		if err != nil {
			log.Fatalf("error while receving server stream : %v", err)
		}

		fmt.Println("Response From GreetManyTimes Server : ", msg.Result)
	}

}

func ComputeAverage(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting Client Side Streaming over GRPC ....")

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("error occured while performing client-side streaming : %v", err)
	}

	requests := []*calculatorpb.AvgRequest{
		&calculatorpb.AvgRequest{
			Num: 2,
		}, &calculatorpb.AvgRequest{
			Num: 3,
		}, &calculatorpb.AvgRequest{
			Num: 1,
		},
	}
	for _, req := range requests {
		fmt.Println("\nSending Request.... : ", req)
		stream.Send(req)
		time.Sleep(1 * time.Second)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from server : %v", err)
	}
	fmt.Println("\n****Response From Server : ", resp.GetResult())

}

func FindMaxNumber(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting Bi-directional stream by calling GreetEveryone over GRPC......")
	requests := []*calculatorpb.MaxRequest{
		&calculatorpb.MaxRequest{
			Num: 2,
		}, &calculatorpb.MaxRequest{
			Num: 3,
		}, &calculatorpb.MaxRequest{
			Num: 1,
		},
	}
	stream, err := c.FindMaxNumber(context.Background())
	if err != nil {
		log.Fatalf("error occured while performing client side streaming : %v", err)
	}

	//wait channel to block receiver
	waitchan := make(chan struct{})

	go func(requests []*calculatorpb.MaxRequest) {
		for _, req := range requests {

			fmt.Println("\nSending Request..... : ", req.Num)
			err := stream.Send(req)
			if err != nil {
				log.Fatalf("error while sending request to GreetEveryone service : %v", err)
			}
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}(requests)

	go func() {
		for {

			resp, err := stream.Recv()
			if err == io.EOF {
				close(waitchan)
				return
			}

			if err != nil {
				log.Fatalf("error receiving response from server : %v", err)
			}

			fmt.Printf("\nResponse From Server : %v", resp.GetResult())
		}
	}()

	//block until everything is finished
	<-waitchan

}
