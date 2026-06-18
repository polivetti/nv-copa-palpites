# Revert da liberacao temporaria de palpites

## Contexto

Em 2026-06-18, removemos temporariamente a trava que impedia salvar palpites de jogos depois da vespera da partida.

A intencao e permitir que um usuario consiga registrar os 4 palpites dos jogos do dia e, logo depois, voltar a regra original.

## O que foi alterado

Arquivos modificados:

- `internal/db/db.go`
- `internal/app/app.go`
- `web/templates/main.html`
- `web/templates/groups.html`

Comportamento temporario:

- o backend nao bloqueia mais o salvamento pelo `match_date`;
- os inputs nao ficam mais desabilitados apenas porque chegou o dia do jogo;
- os textos da tela nao falam mais que o palpite fecha na vespera.

Regras que continuam valendo:

- nao permite alterar palpite se o resultado oficial do jogo ja foi informado;
- se o usuario ja tinha criado um palpite, ele continua podendo editar somente em ate 12 horas apos a criacao.

## Validacao feita

O build Docker passou:

```sh
docker build -t nv-copa-verify .
```

Nao foi possivel rodar `go test ./...` diretamente neste ambiente porque o binario `go` nao esta instalado localmente, e a imagem final do app tambem nao inclui `go`.

## Como subir temporariamente

Gerar a imagem normalmente e fazer deploy pelo fluxo atual do projeto.

Antes de liberar para o usuario, confirmar que a imagem em producao contem esta alteracao.

Depois que o usuario salvar os palpites, seguir a secao de revert abaixo.

## Como reverter se a alteracao ainda nao foi commitada

Na maquina com o working tree contendo somente esta alteracao temporaria:

```sh
git checkout -- internal/db/db.go internal/app/app.go web/templates/main.html web/templates/groups.html
```

Depois validar:

```sh
docker build -t nv-copa-verify .
```

## Como reverter se a alteracao foi commitada

Se esta alteracao tiver sido commitada sozinha, use:

```sh
git revert <HASH_DO_COMMIT_DA_LIBERACAO_TEMPORARIA>
```

Depois validar:

```sh
docker build -t nv-copa-verify .
```

## Checklist para depois que o usuario salvar

1. Confirmar no banco que os 4 palpites do usuario foram criados em `fixture_predictions`.
2. Reverter a alteracao temporaria.
3. Gerar nova imagem.
4. Fazer deploy novamente.
5. Conferir na tela que jogos do dia voltaram a aparecer bloqueados pela regra da vespera.

Query util para conferir os palpites do dia:

```sql
SELECT
  u.id AS usuario_id,
  u.name,
  f.match_date,
  f.group_name,
  f.home_team,
  fp.home_score,
  fp.away_score,
  f.away_team,
  fp.created_at,
  fp.updated_at
FROM fixture_predictions fp
JOIN users u ON u.id = fp.user_id
JOIN fixtures f ON f.id = fp.fixture_id
WHERE u.id = USUARIO_ID
  AND f.match_date = '2026-06-18'
ORDER BY f.id;
```
