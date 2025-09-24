if [ ! -f "plugins/wc.so" ]; then
    echo "Compilando plugin..."
    go build -buildmode=plugin -o plugins/wc_con_fallos.so plugins/wc_con_fallos.so
fi

echo "Iniciando workers en terminales separadas..."

for i in {1..3}
do
    gnome-terminal --title="Worker $i" -- bash -c "go run worker/worker.go plugins/wc_con_fallos.so; read -p 'Presiona Enter para cerrar...'"
done

echo "3 workers iniciados en terminales separadas"