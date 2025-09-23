#!/bin/bash

# Archivo: start_workers_wc.sh
# Script para iniciar workers con el plugin de word count

# Compilar el plugin si no existe
if [ ! -f "plugins/wc.so" ]; then
    echo "Compilando plugin word count..."
    go build -buildmode=plugin -o plugins/wc.so plugins/wc.go
fi

# Levantar 3 workers en terminales separadas
echo "Iniciando workers en terminales separadas..."

# Para gnome-terminal (Ubuntu/GNOME)
for i in {1..3}
do
    gnome-terminal --title="Worker $i" -- bash -c "go run cmd/worker/main.go plugins/wc.so; read -p 'Presiona Enter para cerrar...'"
done

echo "3 workers iniciados en terminales separadas"