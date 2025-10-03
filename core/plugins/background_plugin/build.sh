#!/bin/bash

# Script para compilar o BackgroundPlugin

echo "Compilando BackgroundPlugin..."

# Criar diretório de build se não existir
mkdir -p ../../plugins

# Compilar o plugin
go build -buildmode=plugin -o ../../plugins/background_plugin.so main.go

if [ $? -eq 0 ]; then
    echo "✅ BackgroundPlugin compilado com sucesso!"
    echo "📁 Plugin salvo em: ../../plugins/background_plugin.so"
else
    echo "❌ Erro ao compilar BackgroundPlugin"
    exit 1
fi
