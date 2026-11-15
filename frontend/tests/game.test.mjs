import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import vm from "node:vm";
import ts from "typescript";
import path from "node:path";
import { createRequire } from "node:module";

function compile(file, globals = {}) {
  const source = readFileSync(new URL(`../src/${file}`, import.meta.url), "utf8")
    .replaceAll("import.meta.env.VITE_WS_URL", '"ws://localhost:8080/api/game-connect"')
    .replaceAll("import.meta.env.VITE_API_URL", '"https://api.example.test"');
  const context = { exports: {}, require: (specifier) => {
    const target = specifier.startsWith("@/") ? specifier.slice(2)
      : path.posix.join(path.posix.dirname(file), specifier);
    return compile(target.endsWith(".ts") ? target : `${target}.ts`);
  }, ...globals };
  vm.createContext(context);
  vm.runInContext(ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText, context);
  return context.exports;
}

test("HTTP API uses its own URL independently of the WebSocket host", () => {
  const { backendURL } = compile("api/backendURL.ts", {
    URL, window: { location: { href: "https://frontend.example.test/" } },
  });
  assert.equal(backendURL("/api/team-options").href, "https://api.example.test/api/team-options");
  assert.equal(backendURL("/api/skills-lab").href, "https://api.example.test/api/skills-lab");
});

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

test("skills lab previews weapon restrictions and reuses movement with limited stamina", () => {
  const { labAttackProblem, labMovementUnit, starterTokens } = compile("internal/game/skillsLab.ts");
  const board = Array.from({ length: 225 }, (_, i) => ({ x: i % 15, y: Math.floor(i / 15), terrain: "grass", height: 0, elevation: 0 }));
  const weapons = [
    { id: "pistol", range: 2, effect: [] },
    { id: "shotgun", range: 2, effect: [{id:"shieldBreak",negative:false}] },
    { id: "sniper", range: 5, effect: [{id:"noAdjacentTargets",negative:true}] },
    { id: "katana", range: 1, effect: [{id:"sameLevelOnly",negative:true}] },
    { id: "warHammer", range: 1, effect: [{id:"knockback",negative:false},{id:"sameLevelOnly",negative:true}] },
  ];
  const attacker = { id: "a", team: 0, x: 4, y: 5, life: 6, shield: 0, maxLife: 6, maxShield: 0, weaponID: "pistol", path: [] };
  const target = { ...attacker, id: "b", team: 1, x: 5 };
  for (const weapon of weapons) {
    const user = { ...attacker, weaponID: weapon.id };
    assert.equal(labAttackProblem(user, target, board, weapons) === null, weapon.id !== "sniper");
    assert.ok(labAttackProblem(user, { ...target, y: 6 }, board, weapons));
    assert.ok(labAttackProblem(user, { ...target, y: 7 }, board, weapons));
  }
  assert.equal(labAttackProblem({ ...attacker, weaponID: "sniper" }, { ...target, x: 9 }, board, weapons), null);
  assert.ok(labAttackProblem({ ...attacker, weaponID: "sniper" }, { ...target, x: 10 }, board, weapons));
  assert.ok(labAttackProblem(attacker, undefined, board, weapons));
  assert.ok(labAttackProblem(attacker, { ...target, team: 0 }, board, weapons));
  assert.ok(labAttackProblem(attacker, { ...target, life: 0 }, board, weapons));
  board.find((tile) => tile.x === 5 && tile.y === 5).height = 1;
  for (const id of ["katana", "warHammer"]) assert.ok(labAttackProblem({ ...attacker, weaponID: id }, target, board, weapons));
  const unit = labMovementUnit(attacker, weapons);
  assert.equal(unit.currentStamina, 4);
  assert.equal(unit.unit.weapons[0].id, "pistol");
  const starters = starterTokens(board, weapons);
  assert.equal(starters.length, 3);
  assert.equal(new Set(starters.map((token) => `${token.x},${token.y}`)).size, 3);
  assert.equal(starters.filter((token) => token.team === 0).length, 1);
  for (const target of starters.slice(1)) assert.ok(target.x === starters[0].x || target.y === starters[0].y);
});

test("weapon stamina penalties limit lab paths and preserve expenditure across equipment changes", () => {
  const { labStamina, labMovementUnit, labAttackProblem } = compile("internal/game/skillsLab.ts");
  const { movementOptions } = compile("internal/game/pathfinding.ts");
  const weapons = [
    {id:"pistol",staminaPenalty:0,range:2,effect:[]},
    {id:"shotgun",staminaPenalty:1,range:2,effect:[]},
    {id:"sniper",staminaPenalty:2,range:5,effect:[]},
    {id:"katana",staminaPenalty:0,range:1,effect:[]},
    {id:"warHammer",staminaPenalty:2,range:1,effect:[]},
  ];
  const token = {id:"a",x:0,y:0,team:0,life:3,shield:0,maxLife:3,maxShield:3,weaponID:"pistol",staminaSpent:0,path:[]};
  for (const [index,cap] of [4,3,2,4,2].entries()) assert.equal(labStamina({...token,weaponID:weapons[index].id},weapons).max,cap);
  const spent = {...token,staminaSpent:3};
  assert.equal(labStamina(spent,weapons).current,1);
  assert.equal(labStamina({...spent,weaponID:"sniper"},weapons).current,0);
  assert.equal(labStamina({...spent,weaponID:"pistol"},weapons).current,1);
  const board = Array.from({length:6},(_,x)=>({x,y:0,terrain:"grass",height:0}));
  const user = {...token,weaponID:"shotgun",staminaSpent:1};
  const unit = labMovementUnit(user,weapons);
  const options = movementOptions(board,unit,[unit]);
  assert.equal(unit.currentStamina,2);
  assert.equal(options.paths.has("2,0"),true);
  assert.equal(options.paths.has("3,0"),false);
  board[1].height = 1;
  assert.equal(movementOptions(board,unit,[unit]).paths.has("2,0"),false);
  assert.equal(labAttackProblem({...token,staminaSpent:4},{...token,id:"b",x:1,team:1},board,weapons),null);
  assert.equal(labStamina({...token,weaponID:"sniper",staminaSpent:0},weapons).current,2);
});

test("passive shields apply on lab equipment once and cannot refill after damage", () => {
  const {applyPassiveAbilities,labAbilityProblem} = compile("internal/game/abilities.ts");
  const ability = {id:"shield",passive:true,shield:3,effect:[{id:"grantShield"}]};
  const token = {id:"a",life:3,shield:0,maxShield:0,abilityIDs:["shield"],staminaSpent:3};
  const applied = applyPassiveAbilities(token,[ability]);
  assert.equal(applied.shield,3);
  assert.equal(applied.maxShield,3);
  assert.equal(applied.staminaSpent,3);
  assert.equal(token.shield,0);
  assert.ok(labAbilityProblem(applied,ability,[],[applied],{}));
  const damaged = {...applied,shield:0};
  const unequipped = applyPassiveAbilities({...damaged,abilityIDs:[]},[ability]);
  assert.equal(applyPassiveAbilities({...unequipped,abilityIDs:["shield"]},[ability]).shield,0);
  assert.equal(applyPassiveAbilities({...token,life:0},[ability]).shield,0);
});

test("ability previews enforce usage, cardinal targets, height, and the grenade area", () => {
  const {labAbilityProblem,abilityArea} = compile("internal/game/abilities.ts");
  const board = Array.from({length:225},(_,i)=>({x:i%15,y:Math.floor(i/15),terrain:"grass",height:0}));
  const base = {uses:"single",range:0,maxUsesPerMatch:0,maxActivePerMatch:0,effect:[]};
  const shield = {...base,id:"shield",passive:true,shield:3,effect:[{id:"grantShield"}]};
  const barricade = {...base,id:"barricade",range:1,maxActivePerMatch:3,effect:[{id:"createBarricade"},{id:"sameLevelOnly"}]};
  const jump = {...base,id:"trampoline",range:2,movement:{forwardDistance:2,climbDistance:1,climbLevels:1},effect:[{id:"leap"}]};
  const grenade = {...base,id:"grenade",uses:"unlimited",range:2,areaSize:3,maxUsesPerMatch:1,effect:[{id:"areaDamageOrShieldBreak"}]};
  const actor = {id:"a",x:5,y:5,life:3,shield:0,staminaSpent:4,abilityIDs:["shield","barricade"]};
  assert.ok(labAbilityProblem(actor,shield,board,[actor],{},actor));
  assert.ok(labAbilityProblem({...actor,shield:3},shield,board,[actor],{},actor));
  assert.ok(labAbilityProblem({...actor,abilityUses:{shield:1}},shield,board,[actor],{},actor));
  assert.equal(labAbilityProblem(actor,barricade,board,[actor],{},{x:6,y:5}),null);
  assert.ok(labAbilityProblem(actor,barricade,board,[actor],{},{x:6,y:6}));
  board[5*15+6].height=1;
  assert.ok(labAbilityProblem(actor,barricade,board,[actor],{},{x:6,y:5}));
  const jumper = {...actor,abilityIDs:["trampoline"]};
  assert.equal(labAbilityProblem(jumper,jump,board,[jumper],{},{x:6,y:5}),null);
  board[5*15+6].height=0; board[5*15+6].terrain="water";
  assert.equal(labAbilityProblem(jumper,jump,board,[jumper],{},{x:7,y:5}),null);
  assert.ok(labAbilityProblem(jumper,jump,board,[jumper],{},{x:6,y:5}));
  const grenadier={...actor,abilityIDs:["grenade"]};
  assert.equal(labAbilityProblem(grenadier,grenade,board,[grenadier],{},{x:7,y:5}),null);
  assert.ok(labAbilityProblem(grenadier,grenade,board,[grenadier],{grenade:1},{x:7,y:5}));
  assert.ok(labAbilityProblem(grenadier,grenade,board,[grenadier,{...actor,id:"used",abilityUses:{grenade:1}}],{},{x:7,y:5}));
  assert.equal(abilityArea(grenade,{x:7,y:5},board).length,9);
  assert.equal(abilityArea(grenade,{x:0,y:0},board).length,4);
});

test("direct attack previews respect cover and allow targeting barricades without a unit", () => {
  const {isAttackTarget,isBarricadeTarget}=compile("internal/game/combat.ts");
  const {labAttackProblem}=compile("internal/game/skillsLab.ts");
  const weapon={id:"pistol",range:2,effect:[]};
  const attacker={x:0,y:0,ownerID:"a",currentLife:3,hasAttacked:false,unit:{weapons:[weapon]}};
  const target={...attacker,x:2,ownerID:"b"};
  const board=[0,1,2].map(x=>({x,y:0,height:0,terrain:"grass"}));
  for (const object of ["house","ruin","barricade"]) {
    board[1].object=object;
    assert.equal(isAttackTarget(attacker,target,board),false);
    assert.ok(labAttackProblem({x:0,y:0,life:3,team:0,weaponID:"pistol"},{x:2,y:0,life:3,team:1},board,[weapon]));
    assert.equal(isBarricadeTarget(attacker,board[1],board),object==="barricade");
  }
  assert.equal(labAttackProblem({x:0,y:0,life:3,team:0,weaponID:"pistol"},undefined,board,[weapon],{x:1,y:0}),null);
  board[1].object=undefined;
  assert.equal(isAttackTarget(attacker,target,board),true);
  for (const object of ["tree","forest","rock"]) { board[1].object=object;assert.equal(isAttackTarget(attacker,target,board),true); }
});

test("reusable negative effects control multiplayer and lab previews independently of weapon ID", () => {
  const { weaponEffectProblem } = compile("internal/game/effects.ts");
  const { isAttackTarget } = compile("internal/game/combat.ts");
  const { labAttackProblem } = compile("internal/game/skillsLab.ts");
  const weapon = { id:"pistol",range:3,effect:[{id:"noAdjacentTargets",negative:true},{id:"sameLevelOnly",negative:true}] };
  assert.ok(weaponEffectProblem(weapon,1,0,0));
  assert.ok(weaponEffectProblem(weapon,2,0,1));
  assert.equal(weaponEffectProblem(weapon,2,1,1),null);
  assert.equal(weaponEffectProblem({...weapon,id:"sniper",effect:[]},1,0,1),null);
  const attacker = { x:0,y:0,ownerID:"a",currentLife:6,hasAttacked:false,unit:{weapons:[weapon]} };
  const target = { ...attacker, x:1,ownerID:"b" };
  const board = [0,1,2].map((x)=>({x,y:0,height:0}));
  assert.equal(isAttackTarget(attacker,target,board),false);
  assert.equal(isAttackTarget(attacker,{...target,x:2},board),true);
  board[2].height = 1;
  assert.equal(isAttackTarget(attacker,{...target,x:2},board),false);
  const source = {x:0,y:0,team:0,life:6,weaponID:"pistol"};
  assert.ok(labAttackProblem(source,{x:1,y:0,team:1,life:6},board,[weapon]));
  assert.ok(labAttackProblem(source,{x:2,y:0,team:1,life:6},board,[weapon]));
});

test("movement prefers cheap terrain and respects stamina, occupied corners and map bounds", () => {
  const { movementOptions } = compile("internal/game/pathfinding.ts");
  const board = [];
  for (let y = 0; y < 3; y++) for (let x = 0; x < 5; x++) {
    const height = y === 1 && x > 0 && x < 4 && x % 2 === 1 ? 2 : 0;
    board.push({ x, y, terrain: "grass", height, elevation: height });
  }
  const unit = { x: 0, y: 1, currentStamina: 6, currentLife: 3, unit: { id: "a" } };
  const options = movementOptions(board, unit, [unit]);
  assert.equal(options.costs.get("4,1"), 6);
  assert.equal(options.paths.get("4,1").length, 6);
  assert.equal(options.paths.has("5,1"), false);
  const tired = movementOptions(board, { ...unit, currentStamina: 0 }, [unit]);
  assert.equal(tired.paths.size, 1);
  const blocker = { x: 1, y: 1, currentLife: 3, unit: { id: "b" } };
  const blocked = movementOptions(board, unit, [unit, blocker]);
  assert.equal(blocked.paths.has("1,1"), false);
  assert.equal(blocked.costs.get("1,0"), 2); // Must walk around the occupied corner.
});

test("movement previews use four directions and cannot take diagonal shortcuts", () => {
  const { movementOptions } = compile("internal/game/pathfinding.ts");
  const { canStep, getDestination } = compile("components/ProceduralMap/characterMovement.ts");
  const board = Array.from({ length: 9 }, (_, i) => ({ x:i%3,y:Math.floor(i/3),terrain:"grass",height:0 }));
  const unit = { x:0,y:0,currentStamina:4,currentLife:3,unit:{id:"a"} };
  const options = movementOptions(board, unit, [unit]);
  assert.equal(options.costs.get("2,2"), 4);
  const short = movementOptions(board, { ...unit, currentStamina:1 }, [unit]);
  assert.equal(short.paths.has("1,1"), false);
  const lookup = new Map(board.map((tile) => [`${tile.x},${tile.y}`,tile]));
  assert.equal(canStep(board[0], board[4]), false);
  assert.equal(getDestination(board[0],1,1,lookup), undefined);
  for (const path of options.paths.values()) {
    let previous = unit;
    for (const step of path) {
      assert.equal(Math.abs(step.x-previous.x)+Math.abs(step.y-previous.y),1);
      previous = step;
    }
  }
  const blockers = [{...unit,x:1,unit:{id:"b"}},{...unit,y:1,unit:{id:"c"}}];
  assert.equal(movementOptions(board, unit, [unit,...blockers]).paths.has("1,1"),false);
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

test("reload effect blocks targets on first and odd own turns while remaining reusable", () => {
  const {weaponReloading} = compile("internal/game/effects.ts");
  const {canUnitAttack,isAttackTarget,isBarricadeTarget} = compile("internal/game/combat.ts");
  const {labAttackProblem} = compile("internal/game/skillsLab.ts");
  const weapon = {id:"custom",range:5,effect:[{id:"reloadTurn"}]};
  const board = [0,1,2,3].map(x=>({x,y:0,terrain:"grass",height:0}));
  const target = {x:3,y:0,ownerID:"b",currentLife:3};
  for (const turnCount of [0,1,2,3,4,5,6]) {
    const reloading = turnCount<=0 || turnCount%2===1;
    const attacker = {x:0,y:0,ownerID:"a",currentLife:3,hasAttacked:false,currentStamina:0,turnCount,unit:{weapons:[weapon]}};
    assert.equal(weaponReloading(weapon,turnCount),reloading);
    assert.equal(canUnitAttack(attacker),!reloading);
    assert.equal(isAttackTarget(attacker,target,board),!reloading);
    assert.equal(isBarricadeTarget(attacker,{...board[2],object:"barricade"},board),!reloading);
    const token = {x:0,y:0,life:3,team:0,weaponID:weapon.id,turnCount};
    assert.equal(Boolean(labAttackProblem(token,{...target,life:3,team:1},board,[weapon])),reloading);
  }
  assert.equal(weaponReloading({...weapon,effect:[]},1),false);
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
  for (const [x,y] of [[2,0],[-2,0],[0,2],[0,-2]]) assert.equal(isAttackTarget(attacker,{...target,x,y}),true);
  for (const [x,y] of [[1,1],[-1,1],[1,-1],[-1,-1],[1,2]]) assert.equal(isAttackTarget(attacker,{...target,x,y}),false);
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
  gameStore.setState({ accountID:"a",currentTurn:"a",currentUnitID:"a1",status:"playing",unitID:"a1",attackMode:true,
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
  gameStore.setState({pending:false,board:[{x:0,y:0,height:0,terrain:"grass"},{x:0,y:1,height:0,terrain:"grass",object:"barricade"}]});
  gameStore.getState().attackPiece("a1",0,1);
  assert.equal(sent.length,2);
  assert.equal(sent[1].payload.to.y,1);
});

test("unit turns guard all actions and select the next active unit on server updates", () => {
  const sent = [];
  let receive;
  const require = createRequire(import.meta.url);
  const socket = { subscribe(fn) { receive = fn; }, onStatus() {}, sendAction(type, payload) { sent.push({type,payload}); return true; } };
  const types = { movePiece:"movePiece", attackPiece:"attackPiece", useAbility:"useAbility", endTurn:"endTurn" };
  const { gameStore } = compile("store/GameSocket.ts", { require: (specifier) => {
    if (specifier === "zustand") return require("zustand");
    if (specifier === "@/api/webSocket/GameSocket") return { GameSocket:socket, MessageType:types };
    if (specifier === "@/internal/game/combat") return compile("internal/game/combat.ts");
    throw new Error(`Unexpected import ${specifier}`);
  }});
  const first = { x:0,y:0,ownerID:"a",currentLife:3,hasAttacked:false,unit:{id:"a1",weapons:[{range:2}]} };
  const second = { ...first, y:1,unit:{...first.unit,id:"a2"} };
  const enemy = { ...first, x:1,ownerID:"b",unit:{...first.unit,id:"b1"} };
  const payload = { id:"round-test", revision:1, players:{a:{units:[first,second]},b:{units:[enemy]}}, board:[],
    status:"playing", currentTurn:"a", currentUnitID:"a1", roundCounter:1, roundTurn:0,
    roundOrder:[{playerID:"a",unitID:"a1"},{playerID:"b",unitID:"b1"},{playerID:"a",unitID:"a2"}] };
  gameStore.setState({...payload,accountID:"a",unitID:"a2",pending:false});
  gameStore.getState().movePiece("a2",2,1);
  gameStore.getState().attackPiece("a2",1,1);
  gameStore.getState().activateAbility("a2","shield",0,1);
  assert.equal(sent.length,0);
  gameStore.getState().movePiece("a1",0,2);
  assert.equal(sent.at(-1).type,"movePiece");
  gameStore.setState({pending:false,attackMode:true,skillMode:"grenade"});
  receive({type:"gameStateUpdate",payload:{...payload,revision:2,currentTurn:"b",currentUnitID:"b1",roundTurn:1}});
  assert.equal(gameStore.getState().unitID,"b1");
  assert.equal(gameStore.getState().attackMode,false);
  assert.equal(gameStore.getState().skillMode,null);
  assert.equal(gameStore.getState().roundTurn,1);
  gameStore.getState().movePiece("a1",0,3);
  gameStore.getState().endTurn();
  assert.equal(sent.length,1);
  receive({type:"gameStateUpdate",payload:{...payload,revision:3,currentUnitID:"a2",roundTurn:2}});
  assert.equal(gameStore.getState().unitID,"a2");
  gameStore.getState().endTurn();
  assert.equal(sent.at(-1).type,"endTurn");
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
    WebSocket: Socket,
    setTimeout: (fn) => { const id = Symbol(); timers.set(id, fn); return id; },
    clearTimeout: (id) => timers.delete(id),
  });
  return { client: GameSocket, sockets, timers };
}

test("HTTP requests skip ngrok browser warning while preserving JSON headers",()=>{
  const {backendHeaders}=compile("api/backendURL.ts",{Headers});
  const remote=backendHeaders(new URL("https://example.ngrok-free.app/api/team-options"),{"Content-Type":"application/json"});
  assert.equal(remote.get("ngrok-skip-browser-warning"),"1");
  assert.equal(remote.get("Accept"),"application/json");assert.equal(remote.get("Content-Type"),"application/json");
  assert.equal(backendHeaders(new URL("http://localhost:8080/api/team-options")).has("ngrok-skip-browser-warning"),false);
  assert.equal(backendHeaders(new URL("https://example.ngrok-free.app.attacker.test/api")).has("ngrok-skip-browser-warning"),false);
});

test("team drafts validate slot penalties and remain isolated by account and session", () => {
  const {defaultTeam,validateTeam,saveMatchTeam,readMatchTeam,loadTeamDraft}=compile("internal/game/team.ts");
  const options={unitCount:3,abilitySlots:2,weapons:[{id:"pistol",toolPenalty:0},{id:"sniper",toolPenalty:1}],abilities:[{id:"shield"},{id:"trampoline"}]};
  const team=defaultTeam();team[0].abilityIDs=["shield","trampoline"];
  assert.equal(validateTeam(team,options),null);
  const records=new Map();const storage={getItem:(key)=>records.get(key)??null,setItem:(key,value)=>records.set(key,value)};
  saveMatchTeam("room","account-a",team,storage);
  assert.equal(readMatchTeam("room","account-a",storage)[0].abilityIDs[1],"trampoline");
  assert.equal(readMatchTeam("room","account-b",storage),undefined);
  assert.equal(readMatchTeam("other-room","account-a",storage),undefined);
  team[0].weaponID="sniper";assert.ok(validateTeam(team,options));
  team[0].weaponID="pistol";team[0].abilityIDs=["shield","shield"];assert.ok(validateTeam(team,options));
  assert.equal(readMatchTeam("room","account-a",storage)[0].abilityIDs[1],"trampoline");
  records.set("grid-tactics.team.draft","invalid JSON");assert.equal(loadTeamDraft(storage).length,3);
});

test("connection snapshots the chosen team and resends it unchanged when reconnecting",()=>{
  const {client,sockets,timers}=transport();
  const team=Array.from({length:3},()=>({weaponID:"pistol",abilityIDs:["shield"]}));
  client.connect("room","account",team);team[0].weaponID="sniper";team[0].abilityIDs[0]="grenade";
  sockets[0].open();assert.equal(sockets[0].sent[0].payload.team[0].weaponID,"pistol");
  sockets[0].drop();const retry=[...timers.values()][0];timers.clear();retry();sockets[1].open();
  assert.equal(sockets[1].sent[0].payload.team[0].abilityIDs[0],"shield");client.disconnect();
});

test("single lazy WebSocket joins on open, reconnects, and cancels retries on exit", () => {
  const { client, sockets, timers } = transport();
  assert.equal(sockets.length, 0);
  assert.equal(client.sendAction("movePiece", {}), false);
  client.connect("room", "account");
  client.connect("room", "account");
  assert.equal(sockets.length, 1);
  assert.equal(sockets[0].url, "ws://localhost:8080/api/game-connect");
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
