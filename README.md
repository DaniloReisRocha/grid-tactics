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
6. Para atacar, selecione uma unidade sua, clique em **Atacar** e escolha um inimigo em alcance, na mesma linha ou coluna do atacante. Cada unidade pode atacar uma vez por turno, mesmo com stamina zero. Ataques não gastam stamina; `staminaPenalty` reduz a estamina máxima de movimento. O máximo efetivo é a estamina base menos as penalidades das armas equipadas, limitado a zero, e é usado no spawn e na recuperação do turno. A pistola causa 1 de dano e tem alcance 2. Escudo absorve dano antes da vida. Tentativas inválidas não consomem o ataque, e a disponibilidade é renovada no início do próximo turno do dono da unidade.
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
- Em **Selecionar / mover**, selecione uma unidade e clique em um destino. O hover mostra a rota, com movimento limitado à estamina disponível após a penalidade da arma. Mover consome o custo da rota; o botão **Recuperar estamina** restaura até o máximo efetivo.
- Equipe uma das cinco armas e use **Testar ataque**. Os alvos válidos ficam destacados; ataques só atingem a mesma linha ou coluna do atacante. A escopeta quebra o escudo; o sniper não atinge adjacentes; katana e martelo exigem o mesmo nível; o martelo empurra quando o destino permite.
- Use **Editar terreno** para pintar slots vazios com grama, terra, água, casas, árvores ou rochas, nos níveis de 0 a 3.
- A lista permite selecionar qualquer token, inclusive derrotados, para restaurar sua vida e escudo. Também é possível remover tokens, limpar a lista e reiniciar o cenário.

O laboratório permite ataques repetidos, inclusive com estamina zero, e aplica as penalidades de estamina das armas. Com base 4, pistola e katana mantêm 4, escopeta permite 3, e sniper e martelo permitem 2. O gasto de movimento é preservado ao trocar de arma; recuperar estamina ou reiniciar o cenário permite repetir o teste. A penalidade de habilidades continua desativada. As simulações são isoladas das sessões multiplayer. O registro mostra o dano à vida, ao escudo e o resultado dos empurrões. O cenário é temporário e volta ao estado inicial ao recarregar a página.

## Efeitos das armas

`backend/models/effects.go` define `Effect`, com `id`, `name`, `description` e `negative`. Cada arma possui `effect: []Effect` no JSON, incluindo um array vazio para armas sem efeitos. Os IDs são independentes dos IDs das armas: `knockback`, `shieldBreak`, `noAdjacentTargets` e `sameLevelOnly`.

`negative: true` identifica uma restrição para quem equipa a arma. Os efeitos continuam ativos: “sem ataques adjacentes” restringe a distância mínima, e “mesmo nível” restringe a altura do alvo. Katana e martelo compartilham a mesma definição de restrição de altura. Empurrão e quebra de escudo são efeitos positivos para quem ataca.

O combate multiplayer e o laboratório compartilham a aplicação dos efeitos em `backend/internal/realtime/game/weapon_effects.go`. As prévias do frontend consultam os mesmos IDs de efeito; os efeitos negativos também aparecem no painel do laboratório. Adicionar um efeito existente a outra arma permite reutilizar o comportamento sem verificar o ID da arma.

## Recursos e catálogo de habilidades

As constantes em `backend/models/unit_rules.go` definem as unidades iniciais: 3 de vida, capacidade de escudo de 3, 4 de estamina, 1 slot de item e 2 slots de habilidades. O escudo ativo começa em zero; a capacidade de escudo não concede proteção automática. Reconectar preserva o escudo ativo da unidade.

`backend/models/abilities.go` contém `Ability`, `AbilityOptions()` e `AbilityByID()`, seguindo o catálogo de armas. Os IDs fixos são `shield`, `barricade`, `trampoline` e `grenade`. Cada habilidade possui `effect`, política de uso, alcance, descrição e os parâmetros específicos necessários. A barricada compartilha o efeito negativo de mesmo nível com as armas. O trampolim registra separadamente o avanço de dois terrenos e a subida de um terreno de um nível superior.

A granada conserva `uses: "unlimited"` da tabela e `maxUsesPerMatch: 1` da descrição, representados separadamente. Sua área tem lado 3, alcance 2 e dano 2; o efeito afeta aliadas e adversárias, quebrando o escudo de quem possui proteção ou causando dano a quem está sem escudo. A quebra de escudo da granada é um efeito diferente da quebra após dano de uma arma.

O catálogo de habilidades também é enviado em `GET /api/skills-lab`. As unidades iniciais possuem dois slots disponíveis e uma lista de habilidades equipadas vazia. No laboratório, o painel **Skills** permite escolher até duas habilidades diferentes para o token selecionado ou para novos tokens. Os slots aceitam **Nenhuma**, exibem os efeitos e limites da habilidade escolhida e são preservados nas respostas de ataque.

No laboratório, clique em **Usar** na habilidade equipada. Escudo aplica 3 pontos de proteção à própria unidade. Barricada exige um terreno adjacente, seco, livre e do mesmo nível; bloqueia a passagem e usa um asset de muralha, com no máximo três barricadas ativas em todo o cenário. É possível removê-las com o editor de terreno. Trampolim salta dois terrenos em uma direção do grid, no mesmo nível, ou sobe um terreno adjacente de um nível acima; o destino precisa estar livre, mas o terreno intermediário pode conter água ou obstáculos. O salto tem animação própria e não gasta estamina.

Granada permite escolher seu centro até dois terrenos adiante, na mesma linha ou coluna. O hover marca a área 3 × 3, recortada nas bordas do mapa. Todas as unidades vivas na área, incluindo a própria unidade e aliadas, recebem 2 de dano se estiverem sem escudo; quem tem escudo apenas perde a proteção, sem dano excedente à vida. O limite de uma granada vale para o cenário inteiro, independentemente da equipe ou unidade.

`POST /api/skills-lab/ability` aplica os efeitos pelo ID de `Effect`, valida o cenário e retorna cópias do mapa, tokens e contadores. Alvos inválidos não consomem uso. As habilidades de uso único registram o consumo por unidade; trocar de habilidade, atacar, restaurar recursos ou recuperar estamina não renova esse uso. **Reiniciar cenário** ou gerar um novo mapa renova todos os usos. A ativação está conectada ao laboratório; o multiplayer ainda não possui comando de uso de habilidades.

Casas, ruínas e barricadas bloqueiam a linha dos ataques diretos, tanto no multiplayer quanto no laboratório. Árvores e rochas não bloqueiam disparos. Uma unidade protegida por uma estrutura não aparece como alvo válido. Barricadas podem ser selecionadas diretamente em **Testar ataque** ou **Atacar**: um golpe as remove e libera o terreno, sem dano à unidade atrás delas. No multiplayer, destruir uma barricada consome o ataque do turno. Não é possível destruir uma barricada através de outra estrutura. A área da granada também destrói as barricadas que atingir; casas e ruínas permanecem no mapa.

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
