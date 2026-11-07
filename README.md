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
5. Cada passo custa `1 + elevação` de stamina. Unidades vivas bloqueiam o caminho; diagonais não atravessam cantos ocupados.
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

`/example` continua sendo a demonstração local do mapa isométrico procedural, com seus terrenos, personagens e minimapa. O multiplayer usa o tabuleiro tático 10 × 10 do servidor; os dois modos ainda não compartilham o mapa.

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
