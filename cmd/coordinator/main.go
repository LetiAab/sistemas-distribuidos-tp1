package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sistemas-distribuidos-tp1/internal/common/mapreduce"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type TaskInfo struct {
	jobType      mapreduce.JobType
	taskId       int
	completed    bool
	file         string
	assigned     bool
	assignedTime time.Time
}

type coordinatorServer struct {
	mapreduce.UnimplementedMasterServer
	mapTask        []TaskInfo
	mapFinished    bool
	reduceTask     []TaskInfo
	reduceFinished bool
	mutex          sync.Mutex
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

// Asigna un map mientras haya alguno sin completar y sin asignar.
// Asigna un reduce mientras haya alguno sin completar y sin asignar, si todos los maps están completados.
// Si no hay tareas disponibles, devuelve nil.
func GetNextTask(c *coordinatorServer) *TaskInfo {
	// Verificar tareas MAP
	if !c.mapFinished {
		for i := 0; i < len(c.mapTask); i++ {
			task := &c.mapTask[i]
			if !task.completed && !task.assigned {
				task.assigned = true
				task.assignedTime = time.Now()
				return task
			}
		}
	}

	// Verificar tareas REDUCE
	if c.mapFinished && !c.reduceFinished {
		for i := 0; i < len(c.reduceTask); i++ {
			task := &c.reduceTask[i]
			if !task.completed && !task.assigned {
				task.assigned = true
				task.assignedTime = time.Now()
				return task
			}
		}
	}

	// No hay tareas disponibles
	return nil
}

// RequestJob implementa la RPC para asignar trabajo
func (c *coordinatorServer) RequestJob(ctx context.Context, req *mapreduce.Request) (*mapreduce.JobReply, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var task *TaskInfo = GetNextTask(c)

	if task == nil {
		// No hay tareas disponibles actualmente
		return &mapreduce.JobReply{Type: mapreduce.JobType_NONE}, nil
	}

	log.Printf("Asignando tarea: %v %d\n", task.jobType, task.taskId)

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
			task := &c.mapTask[taskId]
			task.completed = true
			task.assigned = false
		}

		// Verificar si todas las tareas MAP están terminadas
		allMapsCompleted := true
		for _, task := range c.mapTask {
			if !task.completed {
				allMapsCompleted = false
				break
			}
		}
		if allMapsCompleted {
			c.mapFinished = true
			log.Println("Todas las tareas MAP completadas - habilitando REDUCE")
		}

	case mapreduce.JobType_REDUCE:
		if taskId < len(c.reduceTask) {
			task := &c.reduceTask[taskId]
			task.completed = true
			task.assigned = false
		}

		// Verificar si todas las tareas REDUCE están terminadas
		allReducesCompleted := true
		for _, task := range c.reduceTask {
			if !task.completed {
				allReducesCompleted = false
				break
			}
		}
		if allReducesCompleted {
			c.reduceFinished = true
			log.Println("Todas las tareas REDUCE completadas")
		}
	}

	log.Printf("Tarea completada: %v %d\n", req.Type, req.TaskId)
	return &mapreduce.Ack{}, nil
}

func (c *coordinatorServer) monitorTasks() {
	for {
		time.Sleep(1 * time.Second) // Supervisar cada segundo

		c.mutex.Lock()
		now := time.Now()

		// Verificar tareas MAP
		for i := range c.mapTask {
			task := &c.mapTask[i]
			if task.assigned && !task.completed && now.Sub(task.assignedTime) > 10*time.Second {
				log.Printf("Tarea MAP %d falló. Reasignando...\n", task.taskId)
				// Eliminar archivos intermedios generados por la tarea fallida
				for r := 0; r < len(c.reduceTask); r++ {
					filename := fmt.Sprintf("files/mr-%d-%d", task.taskId, r)
					if err := os.Remove(filename); err == nil {
						log.Printf("Archivo intermedio eliminado: %s\n", filename)
					}
				}
				task.assigned = false
			}
		}

		// Verificar tareas REDUCE
		for i := range c.reduceTask {
			task := &c.reduceTask[i]
			if task.assigned && !task.completed && now.Sub(task.assignedTime) > 10*time.Second {
				log.Printf("Tarea REDUCE %d falló. Reasignando...\n", task.taskId)
				// Eliminar archivo de salida generado por la tarea fallida
				filename := fmt.Sprintf("files/mr-out-%d", task.taskId)
				if err := os.Remove(filename); err == nil {
					log.Printf("Archivo de salida eliminado: %s\n", filename)
				}
				task.assigned = false
			}
		}

		c.mutex.Unlock()
	}
}

func (c *coordinatorServer) allTasksCompleted() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Verificar si todas las tareas MAP están completadas
	for _, task := range c.mapTask {
		if !task.completed {
			return false
		}
	}

	// Verificar si todas las tareas REDUCE están completadas
	for _, task := range c.reduceTask {
		if !task.completed {
			return false
		}
	}

	// Si todas las tareas están completadas
	return true
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Uso: go run coordinator.go file1 file2 ...\n")
		os.Exit(1)
	}
	files := os.Args[1:]
	log.Printf("Archivos de entrada: %v\n", files)

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

	// Iniciar la supervisión de tareas
	go coordinator.monitorTasks()

	s := grpc.NewServer()
	mapreduce.RegisterMasterServer(s, coordinator)

	// Iniciar el servidor gRPC en una goroutine
	go func() {
		log.Printf("Coordinador escuchando en %s\n", socketPath)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Error al iniciar el servidor gRPC: %v", err)
		}
	}()

	// Esperar a que todas las tareas estén completas
	for {
		if coordinator.allTasksCompleted() {
			log.Println("Todas las tareas MAP y REDUCE están completas. Finalizando coordinador.")
			break
		}
		time.Sleep(1 * time.Second)
	}
}
