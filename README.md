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