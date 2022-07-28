package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	calculatorpb "example.com/ask/proto"
	"google.golang.org/grpc"
)

type server struct {
	calculatorpb.UnimplementedCalculatorServiceServer
}

//Sum(context.Context, *SumRequest) (*SumResponse, error)

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (res *calculatorpb.SumResponse, err error) {
	fmt.Println("Sum Function was invoked to demonstrate unary streaming")

	num1 := req.GetNum1()
	num2 := req.GetNum2()

	result := num1 + num2

	res = &calculatorpb.SumResponse{
		Result: result,
	}

	return res, nil
}

//PrimeNumbers(*PrimeRequest, CalculatorService_PrimeNumbersServer) error

func isPrime(n int64) bool {
	for i := int64(2); i < n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func (*server) PrimeNumbers(req *calculatorpb.PrimeRequest, resp calculatorpb.CalculatorService_PrimeNumbersServer) (err error) {
	fmt.Println("PrimeNumbers function invoked for server side streaming")

	num := req.GetNum()

	for i := int64(2); i < num; i++ {
		if isPrime(i) {
			res := calculatorpb.PrimeResponse{
				Result: i,
			}

			time.Sleep(1000 * time.Millisecond)
			resp.Send(&res)
		}
	}
	return nil

}

//ComputeAverage(CalculatorService_ComputeAverageServer) error

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	fmt.Println("ComputeAverage Function is invoked to demonstrate client side streaming")
	var result, count float32 = 0, 0

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			//we have finished reading client stream
			return stream.SendAndClose(&calculatorpb.AvgResponse{
				Result: result / count,
			})
		}

		if err != nil {
			log.Fatalf("Error while reading client stream : %v", err)
		}

		temp := msg.GetNum()
		result += temp
		count += 1.0
	}
}

func (*server) FindMaxNumber(stream calculatorpb.CalculatorService_FindMaxNumberServer) error {
	result := int64(0)
	for {

		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalf("error while receiving data from GreetEveryone client : %v", err)
			return err
		}

		num := req.GetNum()

		if num > result {
			result = num
			sendErr := stream.Send(&calculatorpb.MaxResponse{
				Result: result,
			})

			if sendErr != nil {
				log.Fatalf("error while sending response to GreetEveryone Client : %v", err)
				return err
			}
		}

	}
}

func main() {
	fmt.Println("Welcome!!")
	listen, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to Listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err = s.Serve(listen); err != nil {
		log.Fatalf("failed to serve : %v", err)
	}

}
