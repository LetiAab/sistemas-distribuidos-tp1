package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"sistemas-distribuidos-tp1/common"
	"sistemas-distribuidos-tp1/common/mapreduce"
	"time"

	"google.golang.org/grpc"
)

// Punteros a funciones cargadas desde el plugin
var mapf func(string, string) []common.KeyValue
var reducef func(string, []string) string

func loadPlugin(filename string) {
	p, err := plugin.Open(filename)
	if err != nil {
		log.Fatalf("No se pudo cargar el plugin %v: %v", filename, err)
	}

	xmapf, err := p.Lookup("Map")
	if err != nil {
		log.Fatalf("No se encontró la función Map en %v: %v", filename, err)
	}
	mapf = *xmapf.(*func(string, string) []common.KeyValue)

	xreducef, err := p.Lookup("Reduce")
	if err != nil {
		log.Fatalf("No se encontró la función Reduce en %v: %v", filename, err)
	}
	reducef = *xreducef.(*func(string, []string) string)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Uso: worker plugin.so\n")
		os.Exit(1)
	}
	pluginFile := os.Args[1]
	loadPlugin(pluginFile)

	// Conexión gRPC (TCP por ahora)
	/*
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("No se pudo conectar con el coordinador: %v", err)
		}
		defer conn.Close()
	*/

	//Conexion con Unix Domain Sockets
	socketPath := "/tmp/mr.sock"
	conn, err := grpc.Dial("unix://"+socketPath,
		grpc.WithInsecure(),
		grpc.WithBlock())

	if err != nil {
		log.Fatalf("No se pudo conectar al coordinator: %v", err)
	}
	defer conn.Close() //cuando termina main, cierra la conexion

	client := mapreduce.NewMasterClient(conn)

	workerID := int32(os.Getpid())

	for {
		// Pedir tarea
		req := &mapreduce.Request{WorkerId: workerID}
		reply, err := client.RequestJob(context.Background(), req)
		if err != nil {
			log.Fatalf("Error al solicitar tarea: %v", err)
		}

		switch reply.Type {
		case mapreduce.JobType_MAP:
			fmt.Printf("Worker recibió tarea MAP %d\n", reply.TaskId)
			doMap(reply)
			_, err := client.ReportFinished(context.Background(), &mapreduce.FinishedRequest{
				Type:   mapreduce.JobType_MAP,
				TaskId: reply.TaskId,
			})
			if err != nil {
				log.Printf("Advertencia: no se pudo reportar finalización de tarea MAP %d: %v", reply.TaskId, err)
			}

		case mapreduce.JobType_REDUCE:
			fmt.Printf("Worker recibió tarea REDUCE %d\n", reply.TaskId)
			doReduce(reply)
			_, err := client.ReportFinished(context.Background(), &mapreduce.FinishedRequest{
				Type:   mapreduce.JobType_REDUCE,
				TaskId: reply.TaskId,
			})
			if err != nil {
				log.Printf("Advertencia: no se pudo reportar finalización de tarea REDUCE %d: %v", reply.TaskId, err)
			}

		case mapreduce.JobType_NONE:
			fmt.Println("No hay más tareas, el worker finaliza")
			return
		}

		time.Sleep(time.Second) // chequear si es necesario
	}
}

func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// Ejecuta la tarea MAP
func doMap(reply *mapreduce.JobReply) {
	filename := reply.Files[0]
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		print(filename)
		log.Fatalf("No se pudo leer el archivo %v: %v", filename, err)
	}

	kva := mapf(filename, string(content))

	encoders := make([]*json.Encoder, reply.NReduce)
	files := make([]*os.File, reply.NReduce)

	for r := 0; r < int(reply.NReduce); r++ {
		outfile := fmt.Sprintf("mr-%d-%d", reply.TaskId, r)
		files[r], err = os.Create(outfile)
		if err != nil {
			log.Fatalf("No se pudo crear el archivo %v: %v", outfile, err)
		}

		defer files[r].Close()
		encoders[r] = json.NewEncoder(files[r])
	}

	for _, kv := range kva {

		reduceTask := ihash(kv.Key) % int(reply.NReduce)

		if err := encoders[reduceTask].Encode(&kv); err != nil {
			log.Printf("Advertencia: no se pudo codificar kv %v: %v", kv, err)
		}
	}

	/*
		// Versión mínima: siempre reduce=0
		outfile := fmt.Sprintf("mr-%d-%d", reply.TaskId, 0)
		ofile, err := os.Create(outfile)
		if err != nil {
			log.Fatalf("No se pudo crear el archivo de salida %v: %v", outfile, err)
		}
		defer ofile.Close()

		enc := json.NewEncoder(ofile)
		for _, kv := range kva {
			if err := enc.Encode(&kv); err != nil {
				log.Printf("Advertencia: no se pudo codificar kv %v: %v", kv, err)
			}
		}*/
}

// Ejecuta la tarea REDUCE
func doReduce(reply *mapreduce.JobReply) {
	intermediate := []common.KeyValue{}

	// Buscar todos los mr-*-TaskId
	files, err := filepath.Glob(fmt.Sprintf("src/mr-*-%d", reply.TaskId))
	if err != nil {
		log.Fatalf("Error al buscar archivos intermedios: %v", err)
	}

	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			log.Printf("Advertencia: no se pudo abrir archivo intermedio %v: %v", f, err)
			continue
		}

		dec := json.NewDecoder(file)
		for {
			var kv common.KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			intermediate = append(intermediate, kv)
		}
		file.Close()
	}

	// Agrupar por clave
	kvsByKey := make(map[string][]string)
	for _, kv := range intermediate {
		kvsByKey[kv.Key] = append(kvsByKey[kv.Key], kv.Value)
	}

	// Escribir salida
	oname := fmt.Sprintf("mr-out-%d", reply.TaskId)
	ofile, err := os.Create(oname)
	if err != nil {
		log.Fatalf("No se pudo crear el archivo de salida %v: %v", oname, err)
	}
	defer ofile.Close()

	for k, vs := range kvsByKey {
		output := reducef(k, vs)
		fmt.Fprintf(ofile, "%v %v\n", k, output)
	}
}
