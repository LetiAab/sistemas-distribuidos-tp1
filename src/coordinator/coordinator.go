package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"sistemas-distribuidos-tp1/common/mapreduce"

	"google.golang.org/grpc"
)

type coordinatorServer struct {
	mapreduce.UnimplementedMasterServer
	mapDone    bool
	reduceDone bool
}

// RequestJob implementa la RPC para asignar trabajo
func (c *coordinatorServer) RequestJob(ctx context.Context, req *mapreduce.Request) (*mapreduce.JobReply, error) {
	if !c.mapDone {
		c.mapDone = true
		fmt.Printf("Asignando tarea MAP al worker %d\n", req.WorkerId)
		return &mapreduce.JobReply{
			Type:    mapreduce.JobType_MAP,
			TaskId:  0,
			Files:   []string{"test_data/pg-test.txt"},
			NReduce: 1,
			NMap:    1,
		}, nil
	} else if !c.reduceDone {
		c.reduceDone = true
		fmt.Printf("Asignando tarea REDUCE al worker %d\n", req.WorkerId)
		return &mapreduce.JobReply{
			Type:    mapreduce.JobType_REDUCE,
			TaskId:  0,
			Files:   []string{"mr-0-0"},
			NReduce: 1,
			NMap:    1,
		}, nil
	}
	// No hay más tareas
	fmt.Printf("No hay más tareas para el worker %d\n", req.WorkerId)
	return &mapreduce.JobReply{Type: mapreduce.JobType_NONE}, nil
}

// ReportFinished implementa la RPC para recibir notificación de tarea completada
func (c *coordinatorServer) ReportFinished(ctx context.Context, req *mapreduce.FinishedRequest) (*mapreduce.Ack, error) {
	fmt.Printf("Tarea completada: %v %d\n", req.Type, req.TaskId)
	return &mapreduce.Ack{}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Error al escuchar el puerto: %v", err)
	}

	s := grpc.NewServer()
	mapreduce.RegisterMasterServer(s, &coordinatorServer{})

	fmt.Println("Coordinador escuchando en el puerto :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error al iniciar el servidor gRPC: %v", err)
	}
}
