#!/bin/bash

# PIN Intent Network POC Demo Frontend Startup Script
# Launches Streamlit web application for real-time monitoring

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
STREAMLIT_PORT=${STREAMLIT_PORT:-8080}
STREAMLIT_HOST=${STREAMLIT_HOST:-localhost}

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  PIN Intent Network - POC Demo Frontend${NC}"
echo -e "${BLUE}================================================${NC}"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the correct directory
if [[ ! -f "$PROJECT_DIR/pyproject.toml" ]]; then
    print_error "pyproject.toml not found. Please run this script from the PIN project root."
    exit 1
fi

# Change to project directory
cd "$PROJECT_DIR"

# Check if uv is installed
if ! command -v uv &> /dev/null; then
    print_error "uv is not installed. Please install it first:"
    echo "  curl -LsSf https://astral.sh/uv/install.sh | sh"
    exit 1
fi

print_info "Using project directory: $PROJECT_DIR"

# Install dependencies if needed
if [[ ! -d ".venv" ]] || [[ ! -f ".venv/pyvenv.cfg" ]]; then
    print_info "Setting up Python virtual environment with uv..."
    uv venv
    print_success "Virtual environment created"
fi

print_info "Installing/updating dependencies..."
uv pip install -e .
print_success "Dependencies installed"

# Check if PIN automation nodes are running
print_info "Checking PIN automation system status..."
check_node() {
    local port=$1
    local name=$2
    
    if curl -s --connect-timeout 2 "http://localhost:$port/health" > /dev/null 2>&1; then
        print_success "Node $name (port $port) is running"
        return 0
    else
        print_warning "Node $name (port $port) is not responding"
        return 1
    fi
}

nodes_running=0
check_node 8100 "Intent Publisher" && ((nodes_running++))
check_node 8101 "Service Agent 1" && ((nodes_running++))
check_node 8102 "Service Agent 2" && ((nodes_running++))
check_node 8103 "Block Builder" && ((nodes_running++))

if [[ $nodes_running -eq 0 ]]; then
    print_error "No PIN automation nodes are running!"
    echo ""
    echo "Please start the automation system first:"
    echo "  ./scripts/automation/start_automation_test.sh"
    echo ""
    echo "Or start individual nodes:"
    echo "  ./scripts/automation/start_node.sh 1  # Intent Publisher"
    echo "  ./scripts/automation/start_node.sh 2  # Service Agent 1"  
    echo "  ./scripts/automation/start_node.sh 3  # Service Agent 2"
    echo "  ./scripts/automation/start_node.sh 4  # Block Builder"
    echo ""
    print_warning "Starting dashboard anyway - it will show connection errors until nodes are available"
elif [[ $nodes_running -lt 4 ]]; then
    print_warning "$nodes_running out of 4 nodes are running. Some dashboard panels may show errors."
else
    print_success "All 4 PIN automation nodes are running!"
fi

# Check if port is already in use
if lsof -Pi :$STREAMLIT_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
    print_error "Port $STREAMLIT_PORT is already in use!"
    print_info "Please stop the existing service or use a different port:"
    echo "  STREAMLIT_PORT=8081 $0"
    exit 1
fi

echo ""
print_info "Starting Streamlit dashboard..."
print_info "Dashboard will be available at: http://$STREAMLIT_HOST:$STREAMLIT_PORT"
print_info "Press Ctrl+C to stop"
echo ""

# Function to handle cleanup on exit
cleanup() {
    print_info "Shutting down dashboard..."
    exit 0
}

# Set up signal handlers for graceful shutdown
trap cleanup SIGINT SIGTERM

# Launch Streamlit application
export PYTHONPATH="$PROJECT_DIR:$PYTHONPATH"

# Change to the streamlit_ui directory for correct module loading
cd "$PROJECT_DIR/streamlit_ui"

# Start Streamlit with configuration
uv run streamlit run main.py \
    --server.port=$STREAMLIT_PORT \
    --server.address=$STREAMLIT_HOST \
    --server.headless=false \
    --browser.gatherUsageStats=false \
    --theme.primaryColor="#FF6B6B" \
    --theme.backgroundColor="#FFFFFF" \
    --theme.secondaryBackgroundColor="#F0F2F6" \
    --theme.textColor="#262730"