#!/bin/bash

# Archivo: start_workers_ii.sh
# Script para iniciar workers con el plugin de inverted index

# Compilar el plugin si no existe
if [ ! -f "plugins/ii.so" ]; then
    echo "Compilando plugin inverted index..."
    go build -buildmode=plugin -o plugins/ii.so plugins/ii.go
fi

# Levantar 3 workers en terminales separadas
echo "Iniciando workers en terminales separadas..."

# Para gnome-terminal (Ubuntu/GNOME)
for i in {1..3}
do
    gnome-terminal --title="Worker $i" -- bash -c "go run cmd/worker/main.go plugins/ii.so; read -p 'Presiona Enter para cerrar...'"
done

echo "3 workers iniciados en terminales separadas"