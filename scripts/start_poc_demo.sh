#!/bin/bash

# PIN Intent Network POC Demo Frontend Startup Script
# Launches Streamlit dashboard for monitoring the 4-node automation system

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
STREAMLIT_PORT=8080
STREAMLIT_HOST="localhost"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Print header
print_header() {
    echo "============================================================"
    echo "      PIN Intent Network - POC Demo Frontend"
    echo "      Real-time Dashboard for 4-Node Automation System"
    echo "============================================================"
    echo ""
}

# Check if uv is installed
check_uv() {
    if ! command -v uv &> /dev/null; then
        log_error "uv package manager not found!"
        log_info "Please install uv: curl -LsSf https://astral.sh/uv/install.sh | sh"
        exit 1
    fi
    log_success "uv package manager found"
}

# Check if PIN nodes are running
check_pin_nodes() {
    log_info "Checking PIN node availability..."
    
    local nodes_status=()
    local ports=(8100 8101 8102 8103)
    local node_names=("Intent Publisher" "Service Agent 1" "Service Agent 2" "Block Builder")
    
    for i in "${!ports[@]}"; do
        port="${ports[$i]}"
        name="${node_names[$i]}"
        
        if curl -s --connect-timeout 3 "http://localhost:${port}/health" >/dev/null 2>&1; then
            log_success "Node $((i+1)) (${name}) - Port ${port}: Online"
            nodes_status+=("online")
        else
            log_warning "Node $((i+1)) (${name}) - Port ${port}: Offline"
            nodes_status+=("offline")
        fi
    done
    
    # Count online nodes
    online_count=0
    for status in "${nodes_status[@]}"; do
        if [[ "$status" == "online" ]]; then
            ((online_count++))
        fi
    done
    
    echo ""
    log_info "PIN Nodes Status: ${online_count}/4 nodes online"
    
    if [[ $online_count -eq 0 ]]; then
        log_error "No PIN nodes are running!"
        log_info "Please start the automation system first:"
        log_info "  ./scripts/automation/start_automation_test.sh"
        echo ""
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    elif [[ $online_count -lt 4 ]]; then
        log_warning "Not all nodes are online. Dashboard will show partial data."
    fi
}

# Setup Python environment
setup_environment() {
    log_info "Setting up Python environment..."
    
    cd "$PROJECT_ROOT"
    
    # Check if pyproject.toml exists
    if [[ ! -f "pyproject.toml" ]]; then
        log_error "pyproject.toml not found in project root!"
        exit 1
    fi
    
    # Install dependencies using uv
    log_info "Installing Python dependencies with uv..."
    uv sync
    
    log_success "Python environment setup completed"
}

# Check if port is available
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warning "Port $port is already in use!"
        
        # Try to find alternative port
        for ((test_port=8080; test_port<=8090; test_port++)); do
            if ! lsof -Pi :$test_port -sTCP:LISTEN -t >/dev/null 2>&1; then
                STREAMLIT_PORT=$test_port
                log_info "Using alternative port: $test_port"
                return 0
            fi
        done
        
        log_error "No available ports found in range 8080-8090"
        exit 1
    fi
}

# Launch Streamlit application
launch_dashboard() {
    log_info "Launching POC Demo Dashboard..."
    
    cd "$PROJECT_ROOT"
    
    # Set environment variables
    export STREAMLIT_SERVER_PORT="$STREAMLIT_PORT"
    export STREAMLIT_SERVER_ADDRESS="$STREAMLIT_HOST"
    export STREAMLIT_SERVER_HEADLESS="true"
    export STREAMLIT_BROWSER_GATHER_USAGE_STATS="false"
    
    log_info "Dashboard configuration:"
    log_info "  URL: http://${STREAMLIT_HOST}:${STREAMLIT_PORT}"
    log_info "  Auto-refresh: Every 5 seconds"
    log_info "  API timeout: 3 seconds"
    echo ""
    
    # Launch Streamlit
    log_success "Starting Streamlit dashboard..."
    echo ""
    echo "============================================================"
    echo "  🌐 PIN Intent Network - POC Demo Dashboard"
    echo "  📊 Access dashboard at: http://${STREAMLIT_HOST}:${STREAMLIT_PORT}"
    echo "  🔄 Auto-refresh enabled (5 second interval)"
    echo "  ⏹️  Press Ctrl+C to stop"
    echo "============================================================"
    echo ""
    
    # Run with uv
    uv run streamlit run poc_demo/main.py \
        --server.port="$STREAMLIT_PORT" \
        --server.address="$STREAMLIT_HOST" \
        --server.headless=true \
        --browser.gatherUsageStats=false \
        --theme.primaryColor="#007bff" \
        --theme.backgroundColor="#ffffff" \
        --theme.secondaryBackgroundColor="#f8f9fa"
}

# Cleanup function
cleanup() {
    log_info "Shutting down POC Demo Dashboard..."
    # Kill any background processes if needed
    exit 0
}

# Set trap for cleanup
trap cleanup SIGINT SIGTERM

# Parse command line arguments
SKIP_NODE_CHECK=false
HELP=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-node-check)
            SKIP_NODE_CHECK=true
            shift
            ;;
        --port)
            STREAMLIT_PORT="$2"
            shift 2
            ;;
        --host)
            STREAMLIT_HOST="$2"
            shift 2
            ;;
        --help|-h)
            HELP=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            HELP=true
            break
            ;;
    esac
done

# Show help
if [[ "$HELP" == true ]]; then
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Launch PIN Intent Network POC Demo Dashboard"
    echo ""
    echo "Options:"
    echo "  --skip-node-check     Skip PIN nodes availability check"
    echo "  --port PORT          Set Streamlit port (default: 8080)"
    echo "  --host HOST          Set Streamlit host (default: localhost)"
    echo "  --help, -h           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                              # Launch with default settings"
    echo "  $0 --port 8081                 # Launch on port 8081"
    echo "  $0 --skip-node-check           # Launch without checking nodes"
    echo "  $0 --host 0.0.0.0 --port 8080  # Launch accessible from network"
    echo ""
    exit 0
fi

# Main execution
main() {
    print_header
    
    # Pre-flight checks
    check_uv
    check_port "$STREAMLIT_PORT"
    
    if [[ "$SKIP_NODE_CHECK" != true ]]; then
        check_pin_nodes
    else
        log_warning "Skipping PIN nodes check as requested"
    fi
    
    echo ""
    
    # Setup and launch
    setup_environment
    launch_dashboard
}

# Execute main function
main "$@"