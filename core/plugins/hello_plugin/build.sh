#!/bin/bash

# Script para compilar o HelloPlugin

echo "Compilando HelloPlugin..."

# Criar diretório de build se não existir
mkdir -p ../../plugins

# Compilar o plugin
go build -buildmode=plugin -o ../../plugins/hello_plugin.so main.go

if [ $? -eq 0 ]; then
    echo "✅ HelloPlugin compilado com sucesso!"
    echo "📁 Plugin salvo em: ../../plugins/hello_plugin.so"
else
    echo "❌ Erro ao compilar HelloPlugin"
    exit 1
fi
