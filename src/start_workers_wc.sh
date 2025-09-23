#!/bin/bash

# Archivo: start_workers.sh

# Compilar el plugin si no existe
if [ ! -f "plugins/wc.so" ]; then
    echo "Compilando plugin..."
    go build -buildmode=plugin -o plugins/wc.so
fi

# Levantar 3 workers en terminales separadas
echo "Iniciando workers en terminales separadas..."

# Para gnome-terminal (Ubuntu/GNOME)
for i in {1..3}
do
    gnome-terminal --title="Worker $i" -- bash -c "go run worker/worker.go plugins/wc.so; read -p 'Presiona Enter para cerrar...'"
done

echo "3 workers iniciados en terminales separadas"