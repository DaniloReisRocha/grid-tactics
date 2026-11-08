import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import vm from "node:vm";
import ts from "typescript";
import path from "node:path";
import { createRequire } from "node:module";

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

test("opposite camera reverses depth and projection on rectangular maps without mutating world coordinates", () => {
  const { viewPosition, projectPosition, viewDepth } = compile("components/ProceduralMap/projection.ts");
  const normal = { width: 9, height: 5, reversed: false };
  const reverse = { ...normal, reversed: true };
  const tiles = Array.from({ length: 45 }, (_, i) => Object.freeze({ x: i % 9, y: Math.floor(i / 9) }));
  const projectedCorner = projectPosition({ x: 8, y: 4 }, normal);
  for (const tile of tiles) {
    const turned = viewPosition(tile, reverse);
    const restored = viewPosition(turned, reverse);
    assert.equal(restored.x, tile.x);
    assert.equal(restored.y, tile.y);
    const before = projectPosition(tile, normal);
    const after = projectPosition(tile, reverse);
    assert.equal(before.x + after.x, projectedCorner.x);
    assert.equal(before.y + after.y, projectedCorner.y);
    assert.equal(projectPosition(turned, normal).x, after.x);
    assert.equal(projectPosition(turned, normal).y, after.y);
  }
  assert.ok(viewDepth(tiles[0], normal) < viewDepth(tiles[44], normal));
  assert.ok(viewDepth(tiles[0], reverse) > viewDepth(tiles[44], reverse));
});

test("movement prefers cheap terrain and respects stamina, occupied corners and map bounds", () => {
  const { movementOptions } = compile("internal/game/pathfinding.ts");
  const board = [];
  for (let y = 0; y < 3; y++) for (let x = 0; x < 5; x++) {
    const height = y === 1 && x > 0 && x < 4 ? 2 : 0;
    board.push({ x, y, terrain: "grass", height, elevation: height });
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

test("trees and rocks are passable; buildings, water and bridge orientation still block", () => {
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
  for (const object of ["tree", "forest", "rock", "rockDirt"]) {
    board[1].object = object;
    assert.equal(movementOptions(board, unit, [unit]).costs.get("2,0"), 2, object);
    const inside = { ...unit, x:1 };
    assert.equal(movementOptions(board, inside, [inside]).costs.get("2,0"), 1, `leaving ${object}`);
    board[1].terrain = "water";
    assert.equal(movementOptions(board, unit, [unit]).paths.has("2,0"), false, `water with ${object}`);
    board[1].terrain = "grass";
  }
  for (const object of ["house", "ruin"]) {
    board[1].object = object;
    assert.equal(movementOptions(board, unit, [unit]).paths.has("2,0"), false, object);
  }
  delete board[1].object; board[1].height = 2;
  assert.equal(movementOptions(board, unit, [unit]).costs.get("2,0"), 4);
});

test("stamina charges uphill differences only, caps the bonus at 2 and previews cumulative costs", () => {
  const { movementCost } = compile("components/ProceduralMap/characterMovement.ts");
  for (const [from, to, expected] of [[0,0,1],[2,2,1],[3,3,1],[0,1,2],[1,2,2],[2,3,2],[0,2,3],[0,3,3],[0,8,3],[3,2,1],[3,0,1]]) {
    assert.equal(movementCost({height:from},{height:to}),expected,`height ${from} -> ${to}`);
  }
  const { movementOptions } = compile("internal/game/pathfinding.ts");
  const board = [0,1,1,3,3,1,2].map((height,x)=>({x,y:0,height,elevation:height,terrain:"grass"}));
  const unit = { x:0,y:0,currentStamina:10,currentLife:3,unit:{id:"a"} };
  const options = movementOptions(board,unit,[unit]);
  assert.equal(options.costs.get("1,0"),2);
  assert.equal(options.costs.get("2,0"),3);
  assert.equal(options.costs.get("3,0"),6);
  assert.equal(options.costs.get("4,0"),7);
  assert.equal(options.costs.get("5,0"),8);
  assert.equal(options.costs.get("6,0"),10);
  assert.equal(movementOptions(board,{...unit,currentStamina:5},[unit]).paths.has("3,0"),false);
  const elevated = { ...unit, x:3, currentStamina:2 };
  assert.equal(movementOptions(board,elevated,[elevated]).costs.get("5,0"),2);
});

test("attack availability and targets ignore stamina and staminaPenalty, but enforce one attack per unit", () => {
  const { canUnitAttack, isAttackTarget } = compile("internal/game/combat.ts");
  const attacker = {
    x:0,y:0,ownerID:"a",currentLife:3,currentStamina:0,hasAttacked:false,
    unit:{id:"a1",stamina:4,weapons:[{range:2,staminaPenalty:999}]},
  };
  const target = { ...attacker,x:1,ownerID:"b",unit:{...attacker.unit,id:"b1"} };
  assert.equal(canUnitAttack(attacker),true);
  assert.equal(isAttackTarget(attacker,target),true);
  assert.equal(isAttackTarget(attacker,undefined),false);
  assert.equal(isAttackTarget(attacker,{...target,x:3}),false);
  assert.equal(isAttackTarget(attacker,{...target,ownerID:"a"}),false);
  assert.equal(isAttackTarget(attacker,{...target,currentLife:0}),false);
  assert.equal(canUnitAttack({...attacker,hasAttacked:true}),false);
  assert.equal(isAttackTarget({...attacker,hasAttacked:true},target),false);
  assert.equal(canUnitAttack({...attacker,currentLife:0}),false);
  assert.equal(canUnitAttack({...attacker,unit:{...attacker.unit,weapons:[]}}),false);
  const other = {...attacker,unit:{...attacker.unit,id:"a2"}};
  assert.equal(canUnitAttack(other),true);
});

test("empty slots do not send attacks, consume allowance or exit attack mode", () => {
  const sent = [];
  const require = createRequire(import.meta.url);
  const socket = {
    subscribe() {}, onStatus() {},
    sendAction(type, payload) { sent.push({type,payload}); return true; },
  };
  const { gameStore } = compile("store/GameSocket.ts", { require: (specifier) => {
    if (specifier === "zustand") return require("zustand");
    if (specifier === "@/api/webSocket/GameSocket") return {GameSocket:socket,MessageType:{attackPiece:"attackPiece"}};
    if (specifier === "@/internal/game/combat") return compile("internal/game/combat.ts");
    throw new Error(`Unexpected import ${specifier}`);
  }});
  const attacker = { x:0,y:0,ownerID:"a",currentLife:3,currentStamina:0,hasAttacked:false,
    unit:{id:"a1",weapons:[{range:2}]} };
  const enemy = { ...attacker,x:1,ownerID:"b",unit:{...attacker.unit,id:"b1"} };
  const corpse = { ...enemy,x:2,currentLife:0,unit:{...enemy.unit,id:"b2"} };
  gameStore.setState({ accountID:"a",currentTurn:"a",status:"playing",unitID:"a1",attackMode:true,
    pending:false,players:{a:{units:[attacker]},b:{units:[enemy,corpse]}} });
  for (const [x,y] of [[0,1],[2,1],[2,0],[0,1]]) {
    gameStore.getState().attackPiece("a1",x,y);
    assert.equal(sent.length,0);
    assert.equal(gameStore.getState().pending,false);
    assert.equal(gameStore.getState().attackMode,true);
    assert.equal(attacker.hasAttacked,false);
  }
  gameStore.getState().attackPiece("a1",1,0);
  assert.equal(sent.length,1);
  assert.equal(sent[0].type,"attackPiece");
  assert.equal(gameStore.getState().pending,true);
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

test("only overlapping structures in front of the selected unit fade, then restore opacity", () => {
  const { occludesUnit, approachOpacity, OCCLUDED_OPACITY } = compile("components/ProceduralMap/occlusion.ts");
  const unit = { minX: 0, minY: 0, maxX: 32, maxY: 48 };
  const overlap = { minX: 10, minY: 15, maxX: 70, maxY: 90 };
  assert.equal(occludesUnit(unit, 15, overlap, 20), true);
  assert.equal(occludesUnit(unit, 15, overlap, 10), false);
  assert.equal(occludesUnit(unit, 15, overlap, 15), false);
  // Even intersecting foreground cubes remain solid at or below the unit's level.
  for (const [height, terrain, expected] of [[0,1,true],[1,1,false],[2,1,false],[2,2,false],[2,3,true],[2,4,true],[1.5,1,false],[1.5,2,true]]) {
    assert.equal(occludesUnit(unit, 15, overlap, 20, { unit: height, terrain }), expected, `unit level ${height}, ground layer ${terrain}`);
  }
  // Taller ground still needs to overlap and be in front, in either camera view.
  assert.equal(occludesUnit(unit, 15, overlap, 10, { unit: 0, terrain: 3 }), false);
  assert.equal(occludesUnit(unit, 15, { ...overlap, minX: 40 }, 20, { unit: 0, terrain: 3 }), false);
  assert.equal(occludesUnit(unit, 15, { ...overlap, minX: 40 }, 20), false);
  assert.equal(occludesUnit(unit, 15, { ...overlap, minY: 50 }, 20), false);
  // Moving beyond the obstacle stops the fade, without relying on a server position.
  assert.equal(occludesUnit({ minX: 80, minY: 0, maxX: 112, maxY: 48 }, 15, overlap, 20), false);
  let alpha = 1;
  for (let frame = 0; frame < 60; frame++) alpha = approachOpacity(alpha, OCCLUDED_OPACITY, 16);
  assert.equal(alpha, 0.2);
  for (let frame = 0; frame < 60; frame++) alpha = approachOpacity(alpha, 1, 16);
  assert.equal(alpha, 1);
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
