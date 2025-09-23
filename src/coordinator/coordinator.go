package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sistemas-distribuidos-tp1/common/mapreduce"
	"sync"

	"google.golang.org/grpc"
)

type TaskInfo struct {
	jobType   mapreduce.JobType
	taskId    int
	completed bool
	file      string
}

type coordinatorServer struct {
	mapreduce.UnimplementedMasterServer
	mapTask               []TaskInfo
	mapFinished           bool
	reduceTask            []TaskInfo
	reduceFinished        bool
	mutex                 sync.Mutex
	counter_task_assigner int
}

// Función constructor para coordinatorServer
func NewCoordinatorServer(files []string, nReduce int) *coordinatorServer {
	// Crear tareas MAP (una por archivo)
	mapTasks := make([]TaskInfo, len(files))
	for i, file := range files {
		mapTasks[i] = TaskInfo{
			jobType:   mapreduce.JobType_MAP,
			taskId:    i,
			completed: false,
			file:      file,
		}
	}

	print(len(mapTasks))

	// Crear tareas REDUCE
	reduceTasks := make([]TaskInfo, nReduce)
	for i := 0; i < nReduce; i++ {
		reduceTasks[i] = TaskInfo{
			jobType:   mapreduce.JobType_REDUCE,
			taskId:    i,
			completed: false,
		}
	}

	return &coordinatorServer{
		mapTask:    mapTasks,
		reduceTask: reduceTasks,
	}
}

// Asigna un map mientras haya alguno sin completar.
// Asigna un reduce mientras haya alguno sin completar, si todos los maps están completados.
// Si no hay tareas disponibles, devuelve nil.
func GetNextTask(c *coordinatorServer) *TaskInfo {
	// if queda algun map not done
	// asignar
	// else if queda algun reduce not done
	// asignar
	// devolver nil

	if !c.mapFinished {
		n := len(c.mapTask)

		for i := 0; i < n; i++ {
			idx := c.counter_task_assigner % n
			task := &c.mapTask[idx]

			if !task.completed {
				return task
			}

			c.counter_task_assigner++ // avanzar siempre
		}
	}

	if !c.reduceFinished {
		n := len(c.reduceTask)

		for i := 0; i < n; i++ {
			idx := c.counter_task_assigner % n
			task := &c.reduceTask[idx]

			if !task.completed {
				return task
			}

			c.counter_task_assigner++ // avanzar siempre
		}
	}

	return nil
}

// RequestJob implementa la RPC para asignar trabajo
func (c *coordinatorServer) RequestJob(ctx context.Context, req *mapreduce.Request) (*mapreduce.JobReply, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var task *TaskInfo = GetNextTask(c)

	if task == nil {
		return &mapreduce.JobReply{Type: mapreduce.JobType_NONE}, nil
	}

	fmt.Printf("Asignando tarea: %v %d %s\n", task.jobType, task.taskId, task.file)

	return &mapreduce.JobReply{
		Type:    task.jobType,
		TaskId:  int32(task.taskId),
		Files:   []string{task.file},
		NReduce: int32(len(c.reduceTask)),
	}, nil
}

// Implementa la RPC para recibir notificación de tarea completada
func (c *coordinatorServer) ReportFinished(ctx context.Context, req *mapreduce.FinishedRequest) (*mapreduce.Ack, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	taskId := int(req.TaskId)

	switch req.Type {
	case mapreduce.JobType_MAP:
		if taskId < len(c.mapTask) {
			c.mapTask[taskId].completed = true
			fmt.Printf("MAP task %d marcada como completada\n", taskId)
		}

		// Verificar si todas las MAP están terminadas
		finishedMapCount := 0
		for _, task := range c.mapTask {
			if task.completed {
				finishedMapCount++
			}
		}
		if finishedMapCount == len(c.mapTask) {
			c.mapFinished = true
			fmt.Printf("Todas las tareas MAP completadas - habilitando REDUCE\n")
		}

	case mapreduce.JobType_REDUCE:
		if taskId < len(c.reduceTask) {
			c.reduceTask[taskId].completed = true
			fmt.Printf("REDUCE task %d marcada como completada\n", taskId)
		}

		// Verificar si todas las REDUCE están terminadas
		finishedReduceCount := 0
		for _, task := range c.reduceTask {
			if task.completed {
				finishedReduceCount++
			}
		}
		if finishedReduceCount == len(c.reduceTask) {
			c.reduceFinished = true
			fmt.Printf("Todas las tareas REDUCE completadas\n")
		}
	}

	fmt.Printf("Tarea completada: %v %d\n", req.Type, req.TaskId)
	return &mapreduce.Ack{}, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Uso: go run coordinator.go file1 file2 ...\n")
		os.Exit(1)
	}
	files := os.Args[1:]
	fmt.Printf("Archivos de entrada: %v\n", files)

	//Conexion con tcp No me borren :(
	/*
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Error al escuchar el puerto: %v", err)
		}
	*/

	//Conexion con Unix Domain Socket
	socketPath := "/tmp/mr.sock"
	os.Remove(socketPath)

	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Error al crear el socket Unix:  %v", err)
	}

	var nReduce = 3

	var coordinator = NewCoordinatorServer(files, nReduce)

	s := grpc.NewServer()
	mapreduce.RegisterMasterServer(s, coordinator)

	fmt.Printf("Coordinador escuchando en %s\n", socketPath)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error al iniciar el servidor gRPC: %v", err)
	}
}
