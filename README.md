# Grid Tactics

Protótipo de jogo tático para dois jogadores, com React/TypeScript no frontend e Go no backend.

## Executar

O backend usa Go 1.23 e exige um PostgreSQL acessível. Configure `backend/envs` usando `backend/envs.example` como referência. `ALLOWED_ORIGINS` deve incluir a origem exata do frontend (por exemplo, `http://localhost:5173`).

No primeiro terminal, a partir da raiz:

```powershell
cd backend
go run ./cmd/server ./envs
```

No segundo terminal:

```powershell
cd frontend
npm install
npm run dev
```

O frontend usa `ws://<hostname>:8080/api/game-connect` em HTTP e `wss://` em HTTPS. Para outro endereço, configure `VITE_WS_URL` no ambiente do Vite.

## Jogar

1. Abra `/home` e gere uma conta de teste.
2. Informe o nome da sessão e entre.
3. Em outra aba, gere outra conta e entre na mesma sessão.
4. No seu turno, selecione uma unidade. O hover mostra a rota e o clique envia o destino ao servidor.
5. Cada passo custa `1 + elevação` de stamina. Unidades vivas, árvores, rochas e estruturas bloqueiam o caminho. Água só pode ser atravessada por pontes, no sentido da ponte; diagonais não cortam obstáculos e desníveis maiores que um cubo não são transitáveis.
6. Para atacar, selecione uma unidade sua, clique em **Atacar** e escolha um inimigo em alcance. A pistola causa 4 de dano, tem alcance 2 e custa 2 de stamina. Escudo absorve dano antes da vida.
7. Finalizar o turno passa a vez e recupera a stamina das unidades vivas do próximo jogador.

Eliminar todas as unidades adversárias vence a partida. **Sair da partida** após o início conta como desistência. Uma queda de conexão pausa o jogo; entrar com a mesma conta restaura a partida. Abrir a mesma conta em outra conexão substitui a conexão anterior.

Contas por UUID são identificadores de teste, não autenticação. O WebSocket de jogo ainda não usa o login JWT disponível nas rotas HTTP. Unidades iniciais são fornecidas pelo serviço de protótipo, e sessões ficam em memória: reiniciar o servidor perde as partidas. Depois de uma partida encerrada, use outro nome de sessão para começar uma nova.

## Organização

- `frontend/src/api/webSocket/GameSocket.ts`: conexão única, reconexão e protocolo.
- `frontend/src/store/GameSocket.ts`: estado confirmado pelo servidor e envio de ações.
- `frontend/src/internal/game/pathfinding.ts`: previsão de caminhos e custos para a interface.
- `frontend/src/containers/home/` e `containers/game/`: entrada, tabuleiro, seleção, combate e turnos.
- `backend/service/game_socket.go`: interpretação das mensagens e encaminhamento das ações.
- `backend/internal/realtime/ws.go`: transporte e fila de envio.
- `backend/internal/realtime/game/`: sessão, geração do tabuleiro, movimento, combate e turnos em arquivos separados.

O multiplayer usa o mapa isométrico procedural de 15 × 15 tiles, gerado pelo servidor. Os tiles completos são enviados aos dois clientes; seleção, hover de caminhos, pontes, relevo e estruturas seguem as regras do servidor. As unidades aparecem no cenário e no minimapa. Cada jogador recebe um modelo diferente entre os três blocos do spritesheet AxulArt (modelo branco, menina e menino), com atribuição estável por sessão. A caminhada anima somente a rota confirmada pelo servidor.

O spawn reserva espaço para os dois times de uma vez. Todas as unidades começam a pelo menos 5 blocos umas das outras, contando as diagonais. Cada posição fica em solo livre, com pelo menos duas saídas transitáveis para uma área conectada do mapa. Se o cenário inicial não comportar essas regras, o servidor tenta variações determinísticas da seed antes de aceitar o primeiro jogador; nunca posiciona unidades em locais inválidos para completar o time.

O componente compartilhado `ProceduralMap` recebe as dimensões, os tiles e as unidades da partida. A geração do cenário e a escolha dos spawns pertencem ao backend.

## Verificações

No frontend:

```powershell
npm run test:game
npm run build
```

No backend:

```powershell
go test ./...
go test -race ./internal/realtime/... ./service/...
```

Os testes de integração usam um servidor HTTP de teste com dois clientes WebSocket, sem exigir PostgreSQL. A checagem de concorrência com `-race` exige suporte a CGO.
