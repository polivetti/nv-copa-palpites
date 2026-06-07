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

## Estrutura

```txt
cmd/server      entrada da aplicacao
internal/app    handlers HTTP e renderizacao
internal/copa   grupos e selecoes
internal/db     SQLite, migracao e consultas
web/templates   HTML renderizado no servidor
web/static      CSS
scripts         backup do banco
```
