# NV Copa

Boilerplate leve para bolao da Copa com Go, HTMX e SQLite.

## Rodando local

```sh
go mod download
go run ./cmd/server
```

Abra `http://localhost:8080`.

## Docker

```sh
docker compose up --build
```

O banco fica em `./data/copa.db`. Backups diarios ficam em `./backups` e sao mantidos por 14 dias.

Na primeira visita, crie um usuario em `/login`. Usuarios usam apenas nome e senha.

Depois do cadastro/login, o app exige o palpite principal:

- campeao
- vice-campeao
- terceiro colocado

Depois disso, a home mostra o podio do usuario e a tela de grupos:

- 1º colocado de todos os 12 grupos
- 2º colocado de todos os 12 grupos
- exatamente 8 terceiros colocados classificados

Para gerar um backup manual:

```sh
docker compose exec app /app/scripts/backup.sh
```

Para agendar na VPS, use um cron no host chamando o mesmo comando:

```cron
0 3 * * * cd /caminho/nv-copa && docker compose exec -T app /app/scripts/backup.sh
```

## CI/CD no GitHub

O repositorio agora tem dois workflows em `.github/workflows`:

- `ci.yml`: roda `go test ./...`, `go build ./cmd/server` e `docker build`
- `cd.yml`: em push para `master`, publica a imagem no `GHCR` e faz deploy na VPS por `SSH`

### Secrets do GitHub

Configure estes secrets no repositorio:

- `DEPLOY_HOST`: IP ou dominio da VPS
- `DEPLOY_USER`: usuario SSH da VPS
- `DEPLOY_PORT`: porta SSH da VPS, normalmente `22`
- `DEPLOY_SSH_KEY`: chave privada usada pelo GitHub Actions
- `DEPLOY_PATH`: caminho do projeto na VPS, por exemplo `/opt/nv-copa`
- `GHCR_USERNAME`: usuario do GitHub com acesso ao pacote no GHCR
- `GHCR_TOKEN`: token com permissao de leitura do pacote no GHCR

### Preparacao da VPS

Na VPS, deixe este projeto clonado no caminho definido em `DEPLOY_PATH` e garanta que estes itens existam:

- Docker
- Docker Compose
- diretorios `data/` e `backups/`

O deploy usa:

- imagem publicada em `ghcr.io/<owner>/nv-copa:latest`
- arquivo [docker-compose.prod.yml](/home/olivetti/nvs/nv-copa/docker-compose.prod.yml:1)
- script [scripts/deploy.sh](/home/olivetti/nvs/nv-copa/scripts/deploy.sh:1)

### Primeiro deploy manual na VPS

Depois de clonar o projeto na VPS:

```sh
mkdir -p data backups
export IMAGE=ghcr.io/SEU_USUARIO_OU_ORG/nv-copa:latest
docker compose -f docker-compose.prod.yml up -d
```

Depois disso, os proximos deploys podem ser feitos pelo GitHub Actions.

## Estrutura

```txt
cmd/server      entrada da aplicacao
internal/app    handlers HTTP e renderizacao
internal/copa   grupos e selecoes
internal/db     SQLite, migracao e consultas
web/templates   HTML renderizado no servidor
web/static      CSS
scripts         backup do banco
.github         workflows de CI/CD
```
