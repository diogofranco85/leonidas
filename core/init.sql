-- Script de inicialização do banco de dados
-- Este arquivo é executado automaticamente quando o container PostgreSQL é criado

-- Criar extensões úteis
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Criar schema se não existir
CREATE SCHEMA IF NOT EXISTS leonidas;

-- Definir search_path para incluir o schema leonidas
ALTER DATABASE leonidas_core SET search_path TO leonidas, public;

-- Comentário sobre o banco
COMMENT ON DATABASE leonidas_core IS 'Banco de dados do Leonidas Core API';
