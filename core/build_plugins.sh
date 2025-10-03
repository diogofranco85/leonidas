#!/bin/bash

echo "Compilando plugins..."

# Compilar auth-plugin
echo "Compilando auth-plugin..."
cd /Users/diogofranco/Documents/Desenvolvimento/leonidas/core/plugins/auth_plugin
go mod tidy
go build -buildmode=plugin -o ../../plugins/auth_plugin.so main.go auth.go database.go entities.go migrations.go
if [ $? -eq 0 ]; then
    echo "✅ auth-plugin compilado com sucesso!"
else
    echo "❌ Erro ao compilar auth-plugin"
    exit 1
fi

# Compilar hello-plugin
echo "Compilando hello-plugin..."
cd /Users/diogofranco/Documents/Desenvolvimento/leonidas/core/plugins/hello_plugin
go mod tidy
go build -buildmode=plugin -o ../../plugins/hello_plugin.so main.go
if [ $? -eq 0 ]; then
    echo "✅ hello-plugin compilado com sucesso!"
else
    echo "❌ Erro ao compilar hello-plugin"
    exit 1
fi

# Compilar core
echo "Compilando core..."
cd /Users/diogofranco/Documents/Desenvolvimento/leonidas/core
go build -o leonidas-core main.go
if [ $? -eq 0 ]; then
    echo "✅ Core compilado com sucesso!"
else
    echo "❌ Erro ao compilar core"
    exit 1
fi

echo "🎉 Todos os componentes compilados com sucesso!"
