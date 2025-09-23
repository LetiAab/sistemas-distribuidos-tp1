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
