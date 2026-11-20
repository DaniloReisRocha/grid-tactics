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

Configure as duas URLs separadamente em `frontend/.env` (use `frontend/.env.example` como referência):

```dotenv
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/api/game-connect
```

`VITE_API_URL` define a origem HTTP/HTTPS das chamadas da API. `VITE_WS_URL` define a URL completa WS/WSS da conexão do jogo. Elas podem apontar para servidores diferentes. Reinicie o Vite após alterar as variáveis; em produção, gere um novo build. Para túneis ngrok, use `https://` na API e `wss://` no socket, e inclua a origem do frontend em `ALLOWED_ORIGINS` do backend.

`ALLOWED_ORIGINS` em `backend/envs` aceita origens exatas e padrões como `https://*.ngrok-free.app`. Esse padrão já está configurado para permitir os endereços variáveis do ngrok, tanto no CORS HTTP quanto no WebSocket, sem cadastrar cada túnel novo. Reinicie o backend após alterar essa configuração. Se `ALLOWED_ORIGINS` também estiver definida no ambiente do processo, atualize esse valor: ele tem prioridade sobre o arquivo `envs`.

## Jogar

1. Abra `/home`, monte as três unidades e gere uma conta de teste. Cada unidade recebe uma arma e até dois itens/habilidades, respeitando as penalidades da arma.
2. Informe o nome da sessão e entre.
3. Em outra aba, gere outra conta e entre na mesma sessão.
4. Cada round concede ao menos um turno a cada unidade viva. O servidor sorteia a prioridade inicial das unidades e recalcula os próximos turnos para alternar sempre entre jogadores, inclusive após eliminações e na troca de round. Quando os times têm quantidades diferentes de unidades vivas, unidades do time menor recebem turnos adicionais para manter a alternância. Apenas a unidade da vez pode mover, atacar e usar habilidades; ela é selecionada automaticamente. O hover mostra a rota e o clique envia o destino ao servidor.
5. Cada passo custa `1 + min(2, max(0, altura do destino - altura de origem))` de stamina. No mesmo nível ou descendo, o custo é 1; subir um nível custa 2 e subir dois ou mais níveis custa 3. O adicional é calculado a cada passo, não pela altitude absoluta do terreno. Árvores, florestas e rochas são transitáveis, sem custo adicional. Unidades vivas, casas e ruínas bloqueiam o caminho. Água só pode ser atravessada por pontes, no sentido da ponte; a locomoção usa apenas as quatro direções do grid, sem passos diagonais.
6. Para atacar, use a unidade da vez, clique em **Atacar** e escolha um inimigo em alcance. As oito casas vizinhas, incluindo diagonais, contam como distância 1; ataques mais distantes continuam na mesma linha ou coluna. Uma estrutura em qualquer uma das duas casas laterais de um canto bloqueia o golpe diagonal. Cada unidade pode atacar uma vez em seu turno, mesmo com stamina zero. Ataques não gastam stamina; `staminaPenalty` reduz a estamina máxima de movimento. O máximo efetivo é a estamina base menos as penalidades das armas equipadas, limitado a zero, e é usado no spawn e na recuperação do turno. A pistola causa 1 de dano e tem alcance 2. Escudo absorve dano antes da vida. Tentativas inválidas não consomem o ataque, e a disponibilidade é renovada no início do próximo turno daquela unidade.
7. **Finalizar turno da unidade** passa a vez para a próxima unidade da sequência. Somente a unidade que entra em turno recupera sua stamina e seu ataque. Quando todas as unidades vivas tiverem jogado, começa outro round, iniciando pelo adversário de quem encerrou o último turno. Usos únicos de habilidades não são renovados. Turnos já executados são preservados; os próximos são recalculados sem unidades eliminadas, mantendo a prioridade relativa das sobreviventes sempre que possível. Um jogador nunca recebe dois turnos consecutivos. Reconectar preserva o round e a unidade ativa.

Eliminar todas as unidades adversárias vence a partida. **Sair da partida** após o início conta como desistência. Uma queda de conexão pausa o jogo; entrar com a mesma conta restaura a partida. Abrir a mesma conta em outra conexão substitui a conexão anterior.

Cada turno de unidade dura no máximo **45 segundos**, incluindo movimento, ataque e habilidades. O servidor inicia o prazo ao entrar em turno e passa a vez automaticamente quando ele termina, mesmo sem ações ou com a aba em segundo plano. A interface mostra o tempo restante, destacando os últimos 10 segundos. Ações e reconexões durante uma partida em andamento não reiniciam o relógio. Uma desconexão pausa a partida e preserva o tempo restante; reconectar retoma esse saldo. Encerrar manualmente ou finalizar a partida cancela o timer anterior. O prazo e a hora do servidor são enviados pelo WebSocket para sincronizar a contagem dos dois jogadores.

No modo de ataque, os terrenos que a arma pode alvejar aparecem em amarelo, respeitando alcance, estruturas, cantos e restrições dos efeitos. Terrenos ocupados por inimigos que podem ser atacados aparecem em vermelho. Barricadas destrutíveis permanecem amarelas e clicáveis; casas vazias no alcance não consomem o ataque. A mesma prévia está disponível no laboratório.

Blocos de relevo entre atacante e alvo também bloqueiam ataques quando interceptam a linha de tiro. O cálculo considera as alturas das duas unidades e toda a largura dos blocos intermediários; terrenos no mesmo nível de um platô não bloqueiam por si só. Nos golpes diagonais adjacentes, as duas casas laterais do canto também são verificadas. A validação do servidor, o laboratório e a área destacada seguem a mesma regra.

Durante uma partida com seis unidades, o servidor cria uma orb ao atingir quatro unidades vivas e outra ao atingir duas. Eliminações múltiplas também contam ao cruzar esses marcos. Cada posição é escolhida pela seed da sessão entre terrenos secos, livres e alcançáveis, sem unidades ou caveiras. As duas orbs nunca compartilham uma posição, mesmo após coleta. Cada marco ocorre uma vez. Atravessar uma orb recupera os usos das habilidades e as passivas da unidade; movimento normal e empurrões coletam em qualquer passo, e saltos coletam apenas no destino. Vida, estamina, ataque utilizado e recarga da arma são preservados. Spawn, coleta e restauração são confirmados pelo servidor e enviados aos dois clientes; reconectar preserva o estado das orbs.

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

Abra `/skills-lab` pelo atalho na página inicial. Não é necessário criar conta ou entrar em uma partida. O backend precisa estar atualizado e em execução; as chamadas HTTP usam o servidor configurado em `VITE_API_URL`.

- Escolha a arena de testes ou gere um mapa procedural por seed.
- Use **+ Token** para adicionar unidades de duas equipes, escolhendo personagem, arma, vida e escudo.
- Use **+ Orb** para posicionar orbs manualmente em terrenos acessíveis e livres. Os frames 2–4 de `Explosion_5` ficam em loop; quando um personagem chega ao slot, inclusive durante uma rota com vários passos, a orb renova os usos das habilidades daquela unidade e reaplica suas passivas, restaurando o Escudo quando equipado. Vida, estamina e recarga da arma são preservadas. Os frames 5–9 são reproduzidos uma vez e a orb desaparece; cada orb só pode ser coletada uma vez. A ferramenta **Remover** também remove orbs, e reiniciar ou gerar outro cenário limpa todas elas.
- Em **Selecionar / mover**, selecione uma unidade e clique em um destino. O hover mostra a rota, com movimento limitado à estamina disponível após a penalidade da arma. Mover consome o custo da rota; o botão **Recuperar estamina** restaura até o máximo efetivo.
- Equipe uma das cinco armas e use **Testar ataque**. Os alvos válidos ficam destacados; ataques alcançam as oito casas vizinhas; além delas, exigem a mesma linha ou coluna. Casas, ruínas e barricadas bloqueiam os cantos diagonais. A escopeta quebra o escudo; o sniper não atinge adjacentes; katana e martelo exigem o mesmo nível; o martelo empurra quando o destino permite.
- Use **Editar terreno** para pintar slots vazios com grama, terra, água, casas, árvores ou rochas, nos níveis de 0 a 3.
- A lista permite selecionar qualquer token, inclusive derrotados, para restaurar sua vida e escudo. Também é possível remover tokens, limpar a lista e reiniciar o cenário.

O laboratório permite ataques repetidos, inclusive com estamina zero, e aplica as penalidades de estamina das armas. Com base 5, pistola e katana mantêm 5, escopeta e sniper permitem 4, e martelo permite 3. O gasto de movimento é preservado ao trocar de arma; recuperar estamina ou reiniciar o cenário permite repetir o teste. A penalidade de habilidades continua desativada. As simulações são isoladas das sessões multiplayer. O registro mostra o dano à vida, ao escudo e o resultado dos empurrões. O cenário é temporário e volta ao estado inicial ao recarregar a página.

## Efeitos das armas

`backend/models/effects.go` define `Effect`, com `id`, `name`, `description` e `negative`. Cada arma possui `effect: []Effect` no JSON, incluindo um array vazio para armas sem efeitos. Os IDs são independentes dos IDs das armas: `knockback`, `shieldBreak`, `noAdjacentTargets` e `sameLevelOnly`.

`negative: true` identifica uma restrição para quem equipa a arma. Os efeitos continuam ativos: “sem ataques adjacentes” restringe a distância mínima, e “mesmo nível” restringe a altura do alvo. Katana e martelo compartilham a mesma definição de restrição de altura. Empurrão e quebra de escudo são efeitos positivos para quem ataca.

O empurrão desloca o alvo vivo até duas casas para trás, validando cada passo e parando antes de obstáculos, outras unidades, bordas, água sem ponte ou subidas. Se apenas a primeira casa estiver livre, o alvo avança uma casa. O empurrão acompanha também golpes diagonais, respeitando os bloqueios em cada canto. A animação percorre as casas intermediárias, tanto no combate quanto no laboratório.

O rifle sniper possui o efeito negativo reutilizável `noHigherTargets`: só atinge alvos no mesmo nível ou abaixo do terreno do atacante. Essa restrição também remove terrenos mais altos da área de ataque destacada. Possui ainda o efeito `reloadTurn`: no primeiro turno da própria unidade ela recarrega e não pode atacar; no segundo pode atacar, no terceiro recarrega, e assim por diante. O ciclo avança mesmo sem disparar e considera somente os turnos daquela unidade. Movimento e habilidades continuam disponíveis durante a recarga. O contador `turnCount` é preservado em reconexões. No laboratório, **Próximo turno do token** avança esse ciclo e recupera a estamina; restaurar recursos ou trocar de arma não reinicia o contador. Nos turnos disponíveis, o laboratório mantém seus ataques repetidos para testes.

O combate multiplayer e o laboratório compartilham a aplicação dos efeitos em `backend/internal/realtime/game/weapon_effects.go`. As prévias do frontend consultam os mesmos IDs de efeito; os efeitos negativos também aparecem no painel do laboratório. Adicionar um efeito existente a outra arma permite reutilizar o comportamento sem verificar o ID da arma.

## Recursos e catálogo de habilidades

As constantes em `backend/models/unit_rules.go` definem as unidades iniciais: 3 de vida, capacidade de escudo de 3, 5 de estamina, 1 slot de item e 2 slots de habilidades. Escudo é uma habilidade passiva (`passive: true`): ao equipá-la, a unidade inicia a partida com 3 pontos de proteção. Sem a passiva, começa sem proteção. O escudo não se regenera entre turnos, rounds ou reconexões; não existe botão para ativá-lo, e o servidor rejeita tentativas de ativação manual.

`backend/models/abilities.go` contém `Ability`, `AbilityOptions()` e `AbilityByID()`, seguindo o catálogo de armas. Os IDs fixos são `shield`, `barricade`, `trampoline`, `grenade` e `medkit`. Cada habilidade possui `effect`, limite de usos, alcance, descrição e os parâmetros específicos necessários. A barricada compartilha o efeito negativo de mesmo nível com as armas. O trampolim registra separadamente o avanço de dois terrenos e a subida de um terreno de um nível superior.

`maxUsesPerMatch` é o único limite de utilizações, contado individualmente por unidade equipada: Barricada permite 3 usos; Granada, Trampolim e Kit Médico permitem 1. O valor zero significa ilimitado; habilidades passivas não podem ser ativadas. A área da granada tem lado 3, alcance 2 e dano 2; o efeito afeta aliadas e adversárias, quebrando o escudo de quem possui proteção ou causando dano a quem está sem escudo. A quebra de escudo da granada é um efeito diferente da quebra após dano de uma arma.

O **Kit Médico** (`medkit`) usa o efeito reutilizável `healAlly` e o parâmetro `heal: 2`. Cura a própria unidade ou uma aliada adjacente nas quatro direções, até a vida máxima. Não revive unidades, não cura inimigos e não consome uso em alvos inválidos ou com vida cheia. É equipado nos mesmos slots de itens/habilidades da equipe, funciona no combate e no laboratório, não gasta estamina ou ataque e pode ter seu uso renovado por orb.

O catálogo de habilidades também é enviado em `GET /api/skills-lab`. As unidades iniciais possuem dois slots disponíveis e uma lista de habilidades equipadas vazia. No laboratório, o painel **Skills** permite escolher até duas habilidades diferentes para o token selecionado ou para novos tokens. Os slots aceitam **Nenhuma**, exibem os efeitos e limites da habilidade escolhida e são preservados nas respostas de ataque.

No laboratório, Escudo aplica sua proteção automaticamente ao equipar a passiva ou adicionar um token com ela. Remover e reequipar a passiva não restaura proteção danificada; os controles de restauração do modo criativo continuam disponíveis. Nas habilidades ativas, clique em **Usar**. Cada unidade com Barricada pode criar até três barricadas na partida, sem compartilhar o limite com outras unidades. Destruir uma barricada não devolve o uso. Barricada exige um terreno adjacente, seco, livre e do mesmo nível; bloqueia a passagem e usa um asset de muralha. É possível removê-la com o editor de terreno. Trampolim salta dois terrenos em uma direção do grid, no mesmo nível, ou sobe um terreno adjacente de um nível acima; o destino precisa estar livre, mas o terreno intermediário pode conter água ou obstáculos. O salto tem animação própria e não gasta estamina.

Granada permite escolher seu centro até dois terrenos adiante, na mesma linha ou coluna. O hover marca a área 3 × 3, recortada nas bordas do mapa. Todas as unidades vivas na área, incluindo a própria unidade e aliadas, recebem 2 de dano se estiverem sem escudo; quem tem escudo apenas perde a proteção, sem dano excedente à vida. Cada unidade equipada pode usar uma granada por partida, mesmo que outra unidade aliada ou adversária já tenha usado a sua. Os contadores individuais persistem entre turnos, rounds e reconexões. O agregado `matchAbilityUses` é informativo e não bloqueia habilidades.

Após a confirmação da granada, os dez sprites importados de `D:\Pictures\PNG\Explosion_3` para `frontend/src/assets/Explosion_3` são reproduzidos uma vez, em ordem numérica, no centro da detonação. A animação respeita a elevação e a rotação do mapa. No combate, o evento `grenadeEvent` identifica a explosão para os dois clientes, inclusive em golpes que encerram a partida; atualizações seguintes não a repetem e reconexões não reproduzem eventos históricos. O laboratório usa a mesma animação após uma ativação válida.

`POST /api/skills-lab/ability` aplica os efeitos pelo ID de `Effect`, valida o cenário e retorna cópias do mapa, tokens e contadores. Alvos inválidos não consomem uso. As habilidades de uso único registram o consumo por unidade; trocar de habilidade, atacar, restaurar recursos ou recuperar estamina não renova esse uso. **Reiniciar cenário** ou gerar um novo mapa renova todos os usos. A ativação também está disponível no multiplayer, pelo comando `useAbility`, com a mesma resolução de efeitos e validação de turno e propriedade da unidade.

Casas, ruínas e barricadas bloqueiam a linha dos ataques diretos, tanto no multiplayer quanto no laboratório. Árvores e rochas não bloqueiam disparos. Uma unidade protegida por uma estrutura não aparece como alvo válido. Barricadas podem ser selecionadas diretamente em **Testar ataque** ou **Atacar**: um golpe as remove e libera o terreno, sem dano à unidade atrás delas. No multiplayer, destruir uma barricada consome o ataque do turno. Não é possível destruir uma barricada através de outra estrutura. A área da granada também destrói as barricadas que atingir; casas e ruínas permanecem no mapa.

## Montagem da equipe

A home consulta `GET /api/team-options` para obter armas, habilidades e recursos do servidor. Cada cartão permite selecionar uma arma e habilidades diferentes para a unidade, mostrando dano, alcance, efeitos e estamina efetiva. A penalidade de habilidades da arma reduz os slots disponíveis: escopeta, sniper e martelo permitem uma habilidade; pistola e katana permitem duas. Trocar de arma mantém apenas as habilidades que cabem no novo limite.

O rascunho é salvo no `sessionStorage` da aba. Ao entrar, a formação também é salva por sessão e conta e enviada no campo `team` de `joinGame`. O servidor aceita somente IDs do catálogo, exige exatamente três unidades e valida duplicatas e limites; vida, escudo, estamina, dano e efeitos são sempre definidos pelo servidor. Links diretos sem seleção conservam a equipe padrão de três pistolas. Reentrar na mesma partida preserva a equipe confirmada, seus recursos e os usos consumidos, mesmo que outro equipamento seja enviado na reconexão.

As habilidades equipadas aparecem no painel da unidade durante o combate. Escudo aparece como **Passiva**, sem botão de ativação. As habilidades ativas permitem selecionar o terreno alvo, com prévia da área da granada. Só a unidade ativa pode usá-las, estando viva e com a habilidade equipada. O combate e o laboratório compartilham a resolução dos efeitos em `backend/internal/realtime/game/ability_effects.go`. Usar uma habilidade não consome o ataque nem estamina, e finalizar o turno não renova usos únicos. A granada mantém o limite de uma utilização por unidade na partida e pode eliminar unidades aliadas, inclusive encerrando a partida sem vencedor se eliminar ambos os times. Se ela eliminar apenas a unidade ativa, o turno avança automaticamente.

## Histórico local das partidas

O servidor cria automaticamente um arquivo JSON por saguão em `historico/`, relativo ao diretório de execução. Ao iniciar o backend dentro de `backend`, os arquivos ficam em `backend/historico`. É possível definir outro caminho com `HISTORY_DIR` no arquivo de ambiente do backend. Os históricos gerados são ignorados pelo Git.

Cada arquivo recebe data/hora UTC e um UUID no nome. Reutilizar um código de sala depois de remover a sessão cria outro arquivo, sem sobrescrever a partida anterior. O histórico começa na criação do saguão e é atualizado após cada alteração confirmada, inclusive enquanto a partida estiver incompleta.

O formato possui `schemaVersion`, identificação da partida e do saguão, seed, capacidade, duração do turno, datas, status, vencedor, `initialState`, `latestState`, `lobbyEvents` e `rounds`. Cada rodada contém `turns`, com número global do turno, posição na ordem da rodada (`roundSlot`, começando em zero), jogador, unidade, horários, motivo do encerramento, estado inicial e `events`.

Cada evento registra uma sequência global crescente, revisão, horário UTC, tipo de ação, jogador, unidade, habilidade e destino quando aplicáveis, além de `state`: o estado completo após a ação. Os estados incluem jogadores e equipes com suas definições de armas/habilidades/efeitos, mapa e elevações, posições e caminhos, vida, escudo, estamina, usos de habilidades, recarga, ordem sorteada dos turnos, orbs, explosões e relógio. Conexões WebSocket e credenciais do servidor não são serializadas.

Movimentação, ataques, habilidades, encerramentos manuais e por tempo, desconexões, reconexões, desistências e fechamento da sessão ficam registrados. Comandos rejeitados não geram eventos. A ação que encerra um turno pertence ao turno que estava ativo antes dela; seu `state` já pode apresentar a próxima unidade. O primeiro turno recebe como `initialState` o resultado da entrada do último jogador, registrado em `lobbyEvents`.

Para uma futura ferramenta de replay, comece em `initialState` e reproduza os eventos de `lobbyEvents` e `rounds[].turns[].events` pela ordem de `sequence`, aplicando cada `state`. Os horários permitem preservar os intervalos, e os caminhos permitem reproduzir os deslocamentos. O resultado independe de sortear novamente a ordem dos turnos ou executar uma versão futura das regras. `latestState` permite consultar diretamente o último estado salvo; a interface de replay ainda não está implementada.

A gravação usa arquivo temporário, sincronização em disco e substituição do arquivo anterior. Falhas de gravação são registradas no log e os eventos permanecem em memória para nova tentativa na próxima alteração. No encerramento normal do servidor, as sessões são fechadas e gravadas novamente. Se o diretório de histórico não puder ser preparado para escrita, o servidor informa o erro e não inicia.

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
