# Starting 3 chunk servers on different ports
echo "Starting chunkservers..."

go run ./cmd/chunkserver --port=9001 --data-dir=./chunkserver_data_1 --master=localhost:9000 & CHUNKSERVER1_PID=$!
go run ./cmd/chunkserver --port=9002 --data-dir=./chunkserver_data_2 --master=localhost:9000 & CHUNKSERVER2_PID=$!
go run ./cmd/chunkserver --port=9003 --data-dir=./chunkserver_data_3 --master=localhost:9000 & CHUNKSERVER3_PID=$!

echo "Chunkservers started:"
echo "  Chunkserver 1: PID $CHUNKSERVER1_PID (port 9001)"
echo "  Chunkserver 2: PID $CHUNKSERVER2_PID (port 9002)" 
echo "  Chunkserver 3: PID $CHUNKSERVER3_PID (port 9003)"

# Wait for user to stop
echo "Press Ctrl+C to stop all chunkservers"
wait

# Cleanup on exit
echo "Stopping chunkservers..."
kill $CHUNKSERVER1_PID $CHUNKSERVER2_PID $CHUNKSERVER3_PID