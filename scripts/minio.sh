#!/bin/bash

# MinIO Management Script for 1337b04rd Project
# This script helps manage MinIO server for the application

MINIO_DATA_DIR="$HOME/minio-data"
MINIO_CONSOLE_PORT="9001"
MINIO_SERVER_PORT="9000"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Function to check if MinIO is running
check_minio_status() {
    if curl -s http://localhost:$MINIO_SERVER_PORT/minio/health/live > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Function to start MinIO
start_minio() {
    print_header "Starting MinIO Server"
    
    if check_minio_status; then
        print_warning "MinIO is already running!"
        return 0
    fi
    
    # Create data directory if it doesn't exist
    if [ ! -d "$MINIO_DATA_DIR" ]; then
        print_status "Creating MinIO data directory: $MINIO_DATA_DIR"
        mkdir -p "$MINIO_DATA_DIR"
    fi
    
    print_status "Starting MinIO server..."
    print_status "Data directory: $MINIO_DATA_DIR"
    print_status "Server URL: http://localhost:$MINIO_SERVER_PORT"
    print_status "Console URL: http://localhost:$MINIO_CONSOLE_PORT"
    print_status "Default credentials: minioadmin / minioadmin"
    
    # Start MinIO in background
    nohup minio server "$MINIO_DATA_DIR" --console-address ":$MINIO_CONSOLE_PORT" > /tmp/minio.log 2>&1 &
    
    # Wait for MinIO to start
    print_status "Waiting for MinIO to start..."
    for i in {1..10}; do
        if check_minio_status; then
            print_status "MinIO started successfully!"
            print_status "You can access the console at: http://localhost:$MINIO_CONSOLE_PORT"
            return 0
        fi
        sleep 2
    done
    
    print_error "Failed to start MinIO. Check /tmp/minio.log for details."
    return 1
}

# Function to stop MinIO
stop_minio() {
    print_header "Stopping MinIO Server"
    
    if ! check_minio_status; then
        print_warning "MinIO is not running!"
        return 0
    fi
    
    # Find and kill MinIO process
    MINIO_PID=$(pgrep -f "minio server")
    if [ -n "$MINIO_PID" ]; then
        print_status "Stopping MinIO process (PID: $MINIO_PID)"
        kill "$MINIO_PID"
        sleep 2
        
        if check_minio_status; then
            print_warning "MinIO is still running, force killing..."
            kill -9 "$MINIO_PID"
        fi
        
        print_status "MinIO stopped successfully!"
    else
        print_warning "No MinIO process found"
    fi
}

# Function to restart MinIO
restart_minio() {
    print_header "Restarting MinIO Server"
    stop_minio
    sleep 2
    start_minio
}

# Function to show MinIO status
show_status() {
    print_header "MinIO Status"
    
    if check_minio_status; then
        print_status "MinIO is running ✓"
        print_status "Server URL: http://localhost:$MINIO_SERVER_PORT"
        print_status "Console URL: http://localhost:$MINIO_CONSOLE_PORT"
        print_status "Data directory: $MINIO_DATA_DIR"
        
        # Show process info
        MINIO_PID=$(pgrep -f "minio server")
        if [ -n "$MINIO_PID" ]; then
            print_status "Process ID: $MINIO_PID"
        fi
    else
        print_error "MinIO is not running ✗"
    fi
}

# Function to show logs
show_logs() {
    print_header "MinIO Logs"
    if [ -f "/tmp/minio.log" ]; then
        tail -n 50 /tmp/minio.log
    else
        print_warning "No log file found at /tmp/minio.log"
    fi
}

# Function to show help
show_help() {
    print_header "MinIO Management Script"
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  start     Start MinIO server"
    echo "  stop      Stop MinIO server"
    echo "  restart   Restart MinIO server"
    echo "  status    Show MinIO status"
    echo "  logs      Show MinIO logs"
    echo "  help      Show this help message"
    echo ""
    echo "Default credentials:"
    echo "  Username: minioadmin"
    echo "  Password: minioadmin"
    echo ""
    echo "URLs:"
    echo "  Server:   http://localhost:$MINIO_SERVER_PORT"
    echo "  Console:  http://localhost:$MINIO_CONSOLE_PORT"
}

# Main script logic
case "${1:-help}" in
    start)
        start_minio
        ;;
    stop)
        stop_minio
        ;;
    restart)
        restart_minio
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
