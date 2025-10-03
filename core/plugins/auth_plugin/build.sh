#!/bin/bash

# Script para compilar o AuthPlugin

echo "Compilando AuthPlugin..."

# Criar diretório de build se não existir
mkdir -p ../../plugins

# Executar go mod tidy
go mod tidy

# Compilar o plugin
go build -buildmode=plugin -o ../../plugins/auth_plugin.so *.go

if [ $? -eq 0 ]; then
    echo "✅ AuthPlugin compilado com sucesso!"
    echo "📁 Plugin salvo em: ../../plugins/auth_plugin.so"
    echo ""
    echo "🔧 Funcionalidades incluídas:"
    echo "   - Sistema de autenticação JWT"
    echo "   - RBAC (Role-Based Access Control)"
    echo "   - Gerenciamento de usuários"
    echo "   - Sistema de migrations automáticas"
    echo "   - Histórico de alterações"
    echo "   - Reset de senha"
    echo "   - Sessões de usuário"
    echo ""
    echo "🚀 O plugin será iniciado automaticamente quando o core subir"
else
    echo "❌ Erro ao compilar AuthPlugin"
    exit 1
fi
