# TP1: MapReduce
Primer Trabajo Práctico de la materia Sistemas Distribuidos I

Consigna: 
https://docs.google.com/document/d/1ZiBafN9aLjimpEpSQG9Z6beNReBKkOSQo1wv9lgORjo/edit?tab=t.0#heading=h.6n82hblvw465

#### Fecha de entrega: 24 de Septiembre de 2025

### Integrantes
* Alen Davies Leccese - 107084
* Agustín Murseli - 107752
* Luca Lazcano - 107044
* Leticia Aab - 106053


### Como Ejecutar (plugins)

Plugin Word Count
Compilar: Desde src/plugins/, ejecutar go build -buildmode=plugin wc.go para generar wc.so
Probar: Ejecutar ./test_wc para verificar que cuenta palabras correctamente (requiere compilar primero go build test_wc.go)

### Descargar libros de Gutenberg

Para demostrar el funcionamiento del TP, se pueden utilizar libros del Proyecto Gutenberg, una librería online con más de 75 000 libros en formato digital, disponibles de forma gratuita.

Para automatizar el proceso de descargar libros de Gutenberg, una forma es usar github.com/michaelnmmeyer/gutenberg. Es necesario contar con `python3`.

- Clonar el repositorio e instalar:

```sh
git clone https://github.com/michaelnmmeyer/gutenberg.git
sudo make -C gutenberg install
```

- Ejecutar el script provisto para descargar libros en `./books`. Se puede modificar el script para cambiar el filtro de búsqueda, el directorio destino, el formato de los archivos descargados y la cantidad de libros.

```sh
chmod +x ./download_pg_books.sh
./download_pg_books.sh
```

### Demo wordcount

Pasos para demostrar el funcionamiento de la implementación de wordcount:

1. Descargar los libros según lo expuesto [anteriormente](#descargar-libros-de-gutenberg).
2. Ejecutar el coordinador:
   
```sh
go run coordinator/coordinator.go ../books/pg-*.txt
```

3. Ejecutar el script `./start_workers_wc.sh`:

```sh
chmod +x ./start_workers_wc.sh
./start_workers_wc.sh
```

4. Al finalizar, concatenar y ordenar los resultados con:

```sh
cat files/mr-out-* | sort > distributed.txt
```

5. Ejecutar la versión secuencial de wordcount mediante:

```sh
go run secuencia.go ../books/pg-*.txt
```

6. Al finalizar, ordenar los resultados:

```sh
sort mr-out-0 > sequential.txt
```

7. Por último, comparar los resultados con `diff`:

```sh
diff sequential.txt distributed.txt
```

8. El output debería ser vacío, demostrando que no existe diferencia en el resultado entre ambas implementaciones de wordcount.

### Tests de integración automatizados

#### Test de funcionamiento normal

Para automatizar la comparación entre las versiones secuencial y distribuida:

```sh
chmod +x ./success_test.sh
./success_test.sh
```

Este test verifica que ambas implementaciones producen resultados idénticos en condiciones normales.

#### Test de tolerancia a fallos

Para demostrar que el sistema maneja fallos de workers correctamente:

```sh
chmod +x ./fault_tolerance_test.sh
./fault_tolerance_test.sh
```

Este test simula fallos de workers durante la ejecución:
- Inicia 4 workers
- Mata 2 workers en diferentes momentos durante el procesamiento
- Verifica que los workers restantes completen todas las tareas
- Compara que el resultado final sea idéntico a la versión secuencial

El test **PASA** si el sistema se recupera de los fallos y produce resultados correctos, demostrando la tolerancia a fallos del coordinador con su sistema de timeouts y reasignación de tareas.
