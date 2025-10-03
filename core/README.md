# Leonidas Core API

API REST desenvolvida em Go com Fiber, PostgreSQL (GORM) e Redis, com sistema de plugins extensível.

## 🚀 Funcionalidades

- **API REST** com Fiber framework
- **PostgreSQL** com GORM ORM
- **Redis** para cache e sessões
- **Health Checks** robustos
- **Sistema de Plugins** extensível
- **Docker** e Docker Compose
- **Configuração** via variáveis de ambiente

## 📋 Pré-requisitos

- Go 1.24+
- Docker e Docker Compose (opcional)
- PostgreSQL (se não usar Docker)
- Redis (se não usar Docker)

## 🛠️ Instalação

### 1. Clone o repositório
```bash
git clone <repository-url>
cd leonidas/core
```

### 2. Instale as dependências
```bash
go mod download
```

### 3. Configure as variáveis de ambiente
```bash
cp env.example .env
# Edite o arquivo .env com suas configurações
```

### 4. Execute com Docker (Recomendado)
```bash
# Iniciar todos os serviços
docker-compose up -d

# Ver logs
docker-compose logs -f app
```

### 5. Execute localmente
```bash
# Com PostgreSQL e Redis rodando localmente
go run main.go

# Ou compile e execute
go build -o leonidas-core main.go
./leonidas-core
```

## 🔧 Configuração

### Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|---------|
| `PORT` | Porta do servidor | `3000` |
| `BASE_PATH` | Caminho base da API | `/core` |
| `API_DESCRIPTION` | Descrição da API | `API Rest do Leonidas Core` |
| `DB_HOST` | Host do PostgreSQL | `localhost` |
| `DB_PORT` | Porta do PostgreSQL | `5432` |
| `DB_NAME` | Nome do banco | `leonidas_core` |
| `DB_USER` | Usuário do banco | `postgres` |
| `DB_PASSWORD` | Senha do banco | `password` |
| `DB_SSLMODE` | Modo SSL do banco | `disable` |
| `REDIS_HOST` | Host do Redis | `localhost` |
| `REDIS_PORT` | Porta do Redis | `6379` |
| `REDIS_PASSWORD` | Senha do Redis | `` |
| `PLUGINS_PATH` | Caminho dos plugins | `./plugins` |

## 📡 Endpoints da API

### Health Checks
- `GET /core/health` - Health check básico
- `GET /core/health/detailed` - Health check detalhado
- `GET /core/health/ready` - Readiness check
- `GET /core/health/live` - Liveness check

### API Principal
- `GET /core/` - Informações da API
- `GET /core/plugins` - Lista de plugins

## 🐳 Docker

### Build da imagem
```bash
docker build -t leonidas-core .
```

### Executar container
```bash
docker run -p 3000:3000 \
  -e DB_HOST=host.docker.internal \
  -e REDIS_HOST=host.docker.internal \
  leonidas-core
```

### Docker Compose
```bash
# Iniciar todos os serviços
docker-compose up -d

# Iniciar com pgAdmin
docker-compose --profile tools up -d

# Parar serviços
docker-compose down

# Ver logs
docker-compose logs -f
```

## 🗄️ Banco de Dados

### Estrutura
- **Users**: Tabela de usuários
- **Plugins**: Tabela de plugins

### Migrações
As migrações são executadas automaticamente pelo GORM na inicialização.

### Acessar pgAdmin
- URL: http://localhost:8080
- Email: admin@leonidas.com
- Senha: admin

## 🔴 Redis

### Comandos úteis
```bash
# Conectar ao Redis
docker exec -it core_redis_1 redis-cli

# Testar conexão
redis-cli ping
```

## 🧪 Testes

### Testar health checks
```bash
# Health check básico
curl http://localhost:3000/core/health

# Health check detalhado
curl http://localhost:3000/core/health/detailed
```

### Testar com variáveis de ambiente
```bash
PORT=8080 BASE_PATH=/api ./leonidas-core
```

## 📁 Estrutura do Projeto

```
core/
├── cmd/                    # Comandos da aplicação
├── internal/
│   ├── entity/            # Entidades do domínio
│   │   ├── user.go
│   │   └── plugin.go
│   └── infrastructure/
│       ├── config/        # Configurações
│       ├── http/          # Servidor HTTP
│       ├── database.go    # Conexão PostgreSQL
│       └── redis.go       # Conexão Redis
├── main.go               # Ponto de entrada
├── Dockerfile           # Imagem Docker
├── docker-compose.yml   # Orquestração
├── init.sql            # Script de inicialização
└── env.example         # Exemplo de configuração
```

## 🔧 Desenvolvimento

### Adicionar nova entidade
1. Crie o arquivo em `internal/entity/`
2. Adicione a migração em `main.go`
3. Crie handlers em `internal/infrastructure/http/`

### Adicionar novo endpoint
1. Adicione a rota em `setupRoutes()`
2. Implemente o handler
3. Teste com curl ou Postman

## 🚀 Deploy

### Produção
```bash
# Build otimizado
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Executar com variáveis de produção
PORT=80 BASE_PATH=/api DB_HOST=prod-db ./main
```

### Kubernetes
Use os health checks `/core/health/ready` e `/core/health/live` para probes.

## 📝 Logs

Os logs são estruturados e incluem:
- Timestamp
- Level (info, warn, error)
- Mensagem
- Contexto adicional

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.
