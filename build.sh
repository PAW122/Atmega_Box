#!/bin/bash

# Sprawdzamy system operacyjny
OS=$(uname -s)
GO_PROJECT_DIR=$(pwd)

# Funkcja budowania projektu
build_project() {
    local target_os=$1
    local target_arch=$2
    local output_name=$3

    echo "Budowanie projektu dla: $target_os/$target_arch..."

    # Ustawiamy zmienne środowiskowe dla systemu docelowego
    GOOS=$target_os GOARCH=$target_arch go build -o $output_name main.go

    if [ $? -eq 0 ]; then
        echo "Build zakończony sukcesem: $output_name"
    else
        echo "Błąd podczas budowania: $output_name"
    fi
}

# Budowanie na Linux (amd64)
build_project "linux" "amd64" "atmega-linux-amd64"

# Budowanie na Windows (amd64)
build_project "windows" "amd64" "atmega-windows-amd64.exe"
