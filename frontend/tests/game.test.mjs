import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import vm from "node:vm";
import ts from "typescript";
import path from "node:path";

function compile(file, globals = {}) {
  const source = readFileSync(new URL(`../src/${file}`, import.meta.url), "utf8")
    .replaceAll("import.meta.env.VITE_WS_URL", "undefined");
  const context = { exports: {}, require: (specifier) => {
    const target = specifier.startsWith("@/") ? specifier.slice(2)
      : path.posix.join(path.posix.dirname(file), specifier);
    return compile(target.endsWith(".ts") ? target : `${target}.ts`);
  }, ...globals };
  vm.createContext(context);
  vm.runInContext(ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } }).outputText, context);
  return context.exports;
}

test("movement prefers cheap terrain and respects stamina, occupied corners and map bounds", () => {
  const { movementOptions } = compile("internal/game/pathfinding.ts");
  const board = [];
  for (let y = 0; y < 3; y++) for (let x = 0; x < 5; x++) {
    board.push({ x, y, terrain: "grass", height: 0, elevation: y === 1 && x > 0 && x < 4 ? 2 : 0 });
  }
  const unit = { x: 0, y: 1, currentStamina: 4, currentLife: 3, unit: { id: "a" } };
  const options = movementOptions(board, unit, [unit]);
  assert.equal(options.costs.get("4,1"), 4);
  assert.equal(options.paths.get("4,1").length, 4);
  assert.equal(options.paths.has("5,1"), false);
  const tired = movementOptions(board, { ...unit, currentStamina: 0 }, [unit]);
  assert.equal(tired.paths.size, 1);
  const blocker = { x: 1, y: 1, currentLife: 3, unit: { id: "b" } };
  const blocked = movementOptions(board, unit, [unit, blocker]);
  assert.equal(blocked.paths.has("1,1"), false);
  assert.equal(blocked.costs.get("1,0"), 2); // Must walk around the occupied corner.
});

test("multiplayer terrain blocks forests, cliffs and water; bridges enforce orientation", () => {
  const { movementOptions } = compile("internal/game/pathfinding.ts");
  const unit = { x: 0, y: 0, currentStamina: 20, currentLife: 3, unit: { id: "a" } };
  const board = [0, 1, 2].map((x) => ({ x, y: 0, terrain: "grass", height: 0, elevation: 0 }));
  board[1].terrain = "water";
  assert.equal(movementOptions(board, unit, [unit]).paths.has("2,0"), false);
  board[1].object = "bridgeX";
  assert.equal(movementOptions(board, unit, [unit]).paths.has("2,0"), true);
  board[1].object = "bridgeY";
  assert.equal(movementOptions(board, unit, [unit]).paths.has("2,0"), false);
  board[1].terrain = "grass";
  board[1].object = "forest";
  assert.equal(movementOptions(board, unit, [unit]).paths.has("2,0"), false);
  delete board[1].object; board[1].height = 2;
  assert.equal(movementOptions(board, unit, [unit]).paths.has("2,0"), false);
});

test("each player has a distinct stable character model from the three sheet blocks", () => {
  const { playerAppearance } = compile("internal/game/appearance.ts");
  const seen = new Set();
  for (const seed of ["room-a", "room-b", "room-c", "grid-tactics"]) {
    const models = [0, 1, 2].map((index) => playerAppearance(seed, index).model);
    assert.equal(new Set(models).size, 3);
    for (const model of models) { assert.ok(model >= 0 && model <= 2); seen.add(model); }
    assert.equal(playerAppearance(seed, 0).model, models[0]);
  }
  assert.equal(seen.size, 3);
});

function transport() {
  const sockets = [];
  const timers = new Map();
  class Socket {
    static OPEN = 1;
    readyState = 0;
    sent = [];
    constructor(url) { this.url = url; sockets.push(this); }
    send(data) { this.sent.push(JSON.parse(data)); }
    close() { this.readyState = 3; this.onclose?.({ code: 1000 }); }
    open() { this.readyState = 1; this.onopen?.(); }
    drop(code = 1006) { this.readyState = 3; this.onclose?.({ code }); }
  }
  const { GameSocket } = compile("api/webSocket/GameSocket.ts", {
    WebSocket: Socket, location: { protocol: "https:", hostname: "example.test" },
    setTimeout: (fn) => { const id = Symbol(); timers.set(id, fn); return id; },
    clearTimeout: (id) => timers.delete(id),
  });
  return { client: GameSocket, sockets, timers };
}

test("single lazy WebSocket joins on open, reconnects, and cancels retries on exit", () => {
  const { client, sockets, timers } = transport();
  assert.equal(sockets.length, 0);
  assert.equal(client.sendAction("movePiece", {}), false);
  client.connect("room", "account");
  client.connect("room", "account");
  assert.equal(sockets.length, 1);
  assert.equal(sockets[0].url, "wss://example.test:8080/api/game-connect");
  assert.equal(sockets[0].sent.length, 0);
  sockets[0].open();
  assert.equal(sockets[0].sent[0].type, "joinGame");
  assert.equal(sockets[0].sent[0].payload.sessionID, "room");
  sockets[0].drop();
  assert.equal(timers.size, 1);
  const retry = [...timers.values()][0]; timers.clear(); retry();
  sockets[1].open();
  assert.equal(sockets[1].sent[0].type, "joinGame");
  sockets[1].drop();
  client.disconnect();
  assert.equal(timers.size, 0);
});

test("account replacement does not create a reconnect war; malformed messages report errors", () => {
  const { client, sockets, timers } = transport();
  const messages = [];
  client.subscribe((message) => messages.push(message));
  client.connect("room", "account"); sockets[0].open();
  sockets[0].onmessage({ data: "invalid JSON" });
  assert.equal(messages[0].type, "actionFailed");
  sockets[0].drop(4001);
  assert.equal(timers.size, 0);
  assert.match(messages[1].payload.message, /outra conexão/);
});
