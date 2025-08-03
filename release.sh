#!/bin/bash

# Go Module Release Script
# This script helps release Go modules with proper semantic versioning tags
# Usage: ./release.sh <module> <version>
# Example: ./release.sh log v1.0.0

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to show usage
show_usage() {
    echo "Usage: $0 <module> <version>"
    echo ""
    echo "Examples:"
    echo "  $0 log v1.0.0              # Release log module version 1.0.0"
    echo "  $0 log v1.1.0              # Release log module version 1.1.0"
    echo "  $0 . v2.0.0                # Release root module version 2.0.0"
    echo ""
    echo "Available modules:"
    find . -name "go.mod" -not -path "./vendor/*" | sed 's|./||' | sed 's|/go.mod||' | sort
    echo ""
    echo "Version format: vX.Y.Z (semantic versioning)"
    echo "  - v1.0.0: Major version (breaking changes)"
    echo "  - v1.1.0: Minor version (new features, backward compatible)"
    echo "  - v1.0.1: Patch version (bug fixes, backward compatible)"
}

# Function to validate version format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        print_error "Invalid version format: $version"
        print_error "Version must be in format vX.Y.Z (e.g., v1.0.0)"
        exit 1
    fi
}

# Function to check if module exists
check_module() {
    local module=$1
    local module_path="."
    
    if [ "$module" != "." ]; then
        module_path="./$module"
    fi
    
    if [ ! -f "$module_path/go.mod" ]; then
        print_error "Module '$module' not found. No go.mod file in $module_path"
        echo ""
        echo "Available modules:"
        find . -name "go.mod" -not -path "./vendor/*" | sed 's|./||' | sed 's|/go.mod||' | sort
        exit 1
    fi
}

# Function to check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository"
        exit 1
    fi
}

# Function to check if working directory is clean
check_working_directory() {
    if [ -n "$(git status --porcelain)" ]; then
        print_error "Working directory is not clean. Please commit or stash changes first."
        git status --short
        exit 1
    fi
}

# Function to check if tag already exists
check_tag_exists() {
    local tag=$1
    if git tag -l | grep -q "^$tag$"; then
        print_error "Tag '$tag' already exists"
        print_info "Existing tags:"
        git tag -l | grep -E "^v[0-9]+\.[0-9]+\.[0-9]+$" | sort -V | tail -10
        exit 1
    fi
}

# Function to run tests for a module
run_tests() {
    local module_path=$1
    local module_name=$2
    
    print_info "Running tests for module '$module_name'..."
    
    cd "$module_path"
    
    # Check if module has tests
    if find . -name "*_test.go" -not -path "./vendor/*" | grep -q .; then
        if go test ./...; then
            print_success "All tests passed"
        else
            print_error "Tests failed for module '$module_name'"
            exit 1
        fi
    else
        print_warning "No tests found for module '$module_name'"
    fi
    
    cd - > /dev/null
}

# Function to build the module
build_module() {
    local module_path=$1
    local module_name=$2
    
    print_info "Building module '$module_name'..."
    
    cd "$module_path"
    
    if go build ./...; then
        print_success "Module built successfully"
    else
        print_error "Build failed for module '$module_name'"
        exit 1
    fi
    
    cd - > /dev/null
}

# Function to create and push tag
create_and_push_tag() {
    local tag=$1
    local module=$2
    
    print_info "Creating tag '$tag'..."
    
    # Create annotated tag with module information
    local tag_message="Release $module $tag"
    if [ "$module" = "." ]; then
        tag_message="Release $tag"
    fi
    
    git tag -a "$tag" -m "$tag_message"
    
    print_info "Pushing tag to remote..."
    git push origin "$tag"
    
    print_success "Tag '$tag' created and pushed successfully"
}

# Function to show module information
show_module_info() {
    local module_path=$1
    local module_name=$2
    
    print_info "Module Information:"
    echo "  Name: $module_name"
    echo "  Path: $module_path"
    
    cd "$module_path"
    
    if [ -f "go.mod" ]; then
        local go_mod_name=$(head -1 go.mod | cut -d' ' -f2)
        echo "  Module Name: $go_mod_name"
        
        local go_version=$(grep "^go " go.mod | cut -d' ' -f2)
        echo "  Go Version: $go_version"
        
        echo "  Dependencies:"
        if grep -q "require" go.mod; then
            grep -A 100 "require" go.mod | grep -E "^\s+[^\s]+" | head -5 | sed 's/^/    /'
            local dep_count=$(grep -A 100 "require" go.mod | grep -E "^\s+[^\s]+" | wc -l)
            if [ "$dep_count" -gt 5 ]; then
                echo "    ... and $((dep_count - 5)) more"
            fi
        else
            echo "    None"
        fi
    fi
    
    cd - > /dev/null
}

# Function to generate release notes
generate_release_notes() {
    local tag=$1
    local module=$2
    local previous_tag=$(git tag -l | grep -E "^${module}/v[0-9]+\.[0-9]+\.[0-9]+$|^v[0-9]+\.[0-9]+\.[0-9]+$" | sort -V | tail -2 | head -1)
    
    if [ -z "$previous_tag" ]; then
        previous_tag=$(git rev-list --max-parents=0 HEAD)
    fi
    
    print_info "Generating release notes since $previous_tag..."
    
    echo ""
    echo "## Release Notes for $tag"
    echo ""
    echo "### Changes since $previous_tag:"
    git log --oneline "$previous_tag..HEAD" -- "${module}" | head -20
    echo ""
    echo "### Installation:"
    if [ "$module" = "." ]; then
        echo '```bash'
        echo "go get github.com/nguyenth/goutils@$tag"
        echo '```'
    else
        echo '```bash'
        echo "go get github.com/nguyenth/goutils/$module@$tag"
        echo '```'
    fi
    echo ""
}

# Main script
main() {
    # Check arguments
    if [ $# -ne 2 ]; then
        print_error "Invalid number of arguments"
        echo ""
        show_usage
        exit 1
    fi
    
    local module=$1
    local version=$2
    
    # Show help if requested
    if [ "$module" = "-h" ] || [ "$module" = "--help" ]; then
        show_usage
        exit 0
    fi
    
    # Validate inputs
    validate_version "$version"
    check_git_repo
    check_module "$module"
    check_working_directory
    
    # Set module path
    local module_path="."
    if [ "$module" != "." ]; then
        module_path="./$module"
    fi
    
    # Create tag name (for submodules, prefix with module name)
    local tag="$version"
    if [ "$module" != "." ]; then
        tag="$module/$version"
    fi
    
    check_tag_exists "$tag"
    
    print_info "Starting release process for module '$module' version '$version'"
    echo ""
    
    # Show module information
    show_module_info "$module_path" "$module"
    echo ""
    
    # Run tests and build
    run_tests "$module_path" "$module"
    build_module "$module_path" "$module"
    
    echo ""
    print_info "Pre-release checks passed. Ready to create tag '$tag'"
    echo ""
    
    # Confirm release
    read -p "Do you want to proceed with the release? (y/N): " -n 1 -r
    echo ""
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Release cancelled"
        exit 0
    fi
    
    # Create and push tag
    create_and_push_tag "$tag" "$module"
    
    # Generate release notes
    generate_release_notes "$tag" "$module"
    
    echo ""
    print_success "Module '$module' version '$version' released successfully!"
    print_info "Tag: $tag"
    
    if [ "$module" = "." ]; then
        print_info "Import: go get github.com/nguyenth/goutils@$tag"
    else
        print_info "Import: go get github.com/nguyenth/goutils/$module@$tag"
    fi
    
    echo ""
    print_info "Next steps:"
    echo "  1. Create a GitHub release from the tag"
    echo "  2. Update documentation if needed"
    echo "  3. Announce the release"
}

# Run main function with all arguments
main "$@"