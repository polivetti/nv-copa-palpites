# Estado Atual do NV Copa

Este documento resume o que ja foi implementado no projeto ate agora.

## Visao Geral

O NV Copa e um app web leve para bolao da Copa, com:

- backend em Go
- renderizacao server-side com templates HTML
- interacoes pontuais com HTMX
- persistencia em SQLite
- execucao em um unico container Docker
- backup local do banco em volume montado

## Fluxo do Usuario

### 1. Autenticacao

Existem duas telas separadas:

- `/login`: entrar com nome e senha
- `/signup`: criar usuario com nome e senha

O login usa sessao com cookie `HttpOnly`.

### 2. Palpite principal obrigatorio

Depois do login, se o usuario ainda nao tiver salvo o palpite principal, ele e direcionado para a tela de podio.

Esse palpite exige:

- campeao
- vice-campeao
- terceiro colocado

Enquanto esse palpite nao existir, o usuario nao entra na area principal.

### 3. Palpite dos grupos

O palpite dos grupos fica em uma tela separada:

- `/groups`

Nessa tela o usuario escolhe:

- 1o colocado de todos os 12 grupos
- 2o colocado de todos os 12 grupos
- exatamente 8 terceiros colocados classificados

Regra importante:

- depois do primeiro salvamento, o palpite dos grupos fica bloqueado
- a partir disso, a tela vira somente leitura

### 4. Palpites das rodadas

A home `/` agora e a area de palpites das rodadas.

Ela mostra:

- menu lateral
- resumo do podio do usuario
- grupo selecionado
- classificacao atual do grupo
- rodada vigente
- jogos da rodada vigente para o grupo selecionado

Cada usuario pode palpitar o placar dos jogos da fase de grupos.

Regra de prazo:

- o palpite fecha na vespera do jogo
- no dia do jogo o campo ja aparece bloqueado

## Rodadas da Fase de Grupos

Ja foram cadastradas as 3 rodadas completas da fase de grupos.

Cada fixture salvo possui:

- fase: `groups`
- rodada
- grupo
- data do jogo
- selecao mandante
- selecao visitante
- resultado real, quando existir

## Rodada Vigente

A rodada vigente e calculada automaticamente.

Regra atual:

- enquanto existir pelo menos um jogo sem resultado real em uma rodada, essa rodada continua sendo a rodada vigente
- a proxima rodada so aparece quando a rodada anterior estiver totalmente encerrada com resultados preenchidos

## Classificacao dos Grupos

Na tela principal, ao selecionar um grupo, o sistema monta a tabela daquele grupo com base nos resultados reais cadastrados nos jogos.

Campos calculados:

- pontos
- jogos
- vitorias
- empates
- derrotas
- gols pro
- gols contra
- saldo de gols

Ordenacao atual:

1. pontos
2. saldo de gols
3. gols pro
4. ordem base das selecoes no grupo

## Rotas Atuais

- `GET /login`: tela de login
- `POST /login`: autentica usuario
- `GET /signup`: tela de cadastro
- `POST /signup`: cria usuario
- `POST /logout`: encerra sessao
- `GET /`: area principal de palpites das rodadas
- `POST /podium`: salva campeao, vice e terceiro
- `GET /groups`: tela do palpite dos grupos
- `POST /groups`: salva palpite dos grupos uma unica vez
- `POST /round-predictions`: salva palpites da rodada para os jogos exibidos

## Estrutura de Dados

Tabelas principais em uso:

- `users`
- `sessions`
- `podium_predictions`
- `group_predictions`
- `fixtures`
- `fixture_predictions`

Tabelas legadas ainda existentes no banco:

- `players`
- `matches`
- `predictions`

Essas tabelas vieram da primeira versao do prototipo e hoje nao sao o foco do fluxo principal.

## Interface Atual

### Home

A home foi reorganizada com:

- menu lateral fixo
- painel de classificacao do grupo
- painel da rodada vigente

### Tela de grupos

A tela `/groups` e separada da home e mostra:

- todos os 12 grupos
- radios de 1o, 2o e 3o lugar
- contador global de terceiros classificados
- estado bloqueado quando ja salvo

## Deploy e Banco

### Docker

O projeto roda via:

```sh
docker compose up --build
```

### Banco

O SQLite fica persistido em:

```txt
./data/copa.db
```

### Backup

O diretorio de backup fica montado em:

```txt
./backups
```

Backup manual:

```sh
docker compose exec app /app/scripts/backup.sh
```

## Estado Atual do Projeto

Ja implementado:

- autenticacao simples com nome e senha
- sessao por cookie
- palpite principal obrigatorio
- tela separada de palpite dos grupos
- bloqueio definitivo do palpite dos grupos apos salvar
- cadastro das 3 rodadas da fase de grupos
- tela de palpites da rodada vigente
- trava de prazo ate a vespera do jogo
- classificacao automatica do grupo a partir dos resultados reais
- menu lateral para navegacao

Ainda pendente:

- tela administrativa para cadastro do resultado real
- permissao de administrador
- avancar oficialmente de rodada a partir dos resultados
- usar os resultados reais para alimentar por completo a experiencia do bolao
