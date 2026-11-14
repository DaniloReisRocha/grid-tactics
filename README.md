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

O frontend lê a URL completa do WebSocket da variável `VITE_WS_URL`, configurada em `frontend/.env`.

## Jogar

1. Abra `/home` e gere uma conta de teste.
2. Informe o nome da sessão e entre.
3. Em outra aba, gere outra conta e entre na mesma sessão.
4. No seu turno, selecione uma unidade. O hover mostra a rota e o clique envia o destino ao servidor.
5. Cada passo custa `1 + min(2, max(0, altura do destino - altura de origem))` de stamina. No mesmo nível ou descendo, o custo é 1; subir um nível custa 2 e subir dois ou mais níveis custa 3. O adicional é calculado a cada passo, não pela altitude absoluta do terreno. Árvores, florestas e rochas são transitáveis, sem custo adicional. Unidades vivas, casas e ruínas bloqueiam o caminho. Água só pode ser atravessada por pontes, no sentido da ponte; a locomoção usa apenas as quatro direções do grid, sem passos diagonais.
6. Para atacar, selecione uma unidade sua, clique em **Atacar** e escolha um inimigo em alcance, na mesma linha ou coluna do atacante. Cada unidade pode atacar uma vez por turno, mesmo com stamina zero. Ataques não gastam stamina; a propriedade `staminaPenalty` está temporariamente ignorada. A pistola causa 4 de dano e tem alcance 2. Escudo absorve dano antes da vida. Tentativas inválidas não consomem o ataque, e a disponibilidade é renovada no início do próximo turno do dono da unidade.
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

## Laboratório de skills

Abra `/skills-lab` pelo atalho na página inicial. Não é necessário criar conta ou entrar em uma partida. O backend precisa estar atualizado e em execução; as chamadas HTTP usam o mesmo servidor configurado em `VITE_WS_URL`.

- Escolha a arena de testes ou gere um mapa procedural por seed.
- Use **+ Token** para adicionar unidades de duas equipes, escolhendo personagem, arma, vida e escudo.
- Em **Selecionar / mover**, selecione uma unidade e clique em um destino. O hover mostra a rota, com movimento sem limite de estamina.
- Equipe uma das cinco armas e use **Testar ataque**. Os alvos válidos ficam destacados; ataques só atingem a mesma linha ou coluna do atacante. A escopeta quebra o escudo; o sniper não atinge adjacentes; katana e martelo exigem o mesmo nível; o martelo empurra quando o destino permite.
- Use **Editar terreno** para pintar slots vazios com grama, terra, água, casas, árvores ou rochas, nos níveis de 0 a 3.
- A lista permite selecionar qualquer token, inclusive derrotados, para restaurar sua vida e escudo. Também é possível remover tokens, limpar a lista e reiniciar o cenário.

O laboratório permite ataques repetidos e mantém as penalidades desativadas. As simulações são isoladas das sessões multiplayer. O registro mostra o dano à vida, ao escudo e o resultado dos empurrões. O cenário é temporário e volta ao estado inicial ao recarregar a página.

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
