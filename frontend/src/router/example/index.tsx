import { useState } from "react";
import ProceduralMap from "@/components/ProceduralMap";

function createRandomSeed() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID().slice(0, 8);
  }

  return Math.random().toString(36).slice(2, 10);
}

export default function MapPage() {
  const [seedInput, setSeedInput] = useState("grid-tactics");
  const [activeSeed, setActiveSeed] = useState("grid-tactics");

  const applySeed = () => setActiveSeed(seedInput.trim() || "grid-tactics");
  const useNewSeed = () => {
    const nextSeed = createRandomSeed();
    setSeedInput(nextSeed);
    setActiveSeed(nextSeed);
  };

  return (
    <main style={styles.page}>
      <ProceduralMap width={15} height={15} seed={activeSeed} />
      <section style={styles.controls} aria-label="Gerador de mapa">
        <div>
          <span style={styles.eyebrow}>GRID TACTICS</span>
          <h1 style={styles.title}>Gerador de mapa</h1>
        </div>

        <label style={styles.label} htmlFor="map-seed">
          Seed
        </label>
        <div style={styles.inputRow}>
          <input
            id="map-seed"
            value={seedInput}
            onChange={(event) => setSeedInput(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") applySeed();
            }}
            style={styles.input}
            spellCheck={false}
          />
          <button
            type="button"
            onClick={applySeed}
            style={styles.primaryButton}
          >
            Gerar
          </button>
          <button
            type="button"
            onClick={useNewSeed}
            style={styles.secondaryButton}
          >
            Nova seed
          </button>
        </div>

        <p style={styles.hint}>
          A mesma seed sempre recria exatamente o mesmo mapa.
        </p>
      </section>
    </main>
  );
}

const styles: Record<string, React.CSSProperties> = {
  page: {
    position: "relative",
    width: "100vw",
    height: "100vh",
    overflow: "hidden",
    background: "#17231f",
    color: "#f4f0df",
  },
  controls: {
    position: "absolute",
    zIndex: 2,
    top: 20,
    left: 20,
    width: "min(420px, calc(100vw - 40px))",
    boxSizing: "border-box",
    padding: "18px 20px",
    border: "1px solid rgba(224, 237, 198, 0.18)",
    borderRadius: 16,
    background: "rgba(19, 31, 27, 0.9)",
    boxShadow: "0 18px 50px rgba(0, 0, 0, 0.34)",
    backdropFilter: "blur(12px)",
  },
  eyebrow: {
    color: "#a7c875",
    fontSize: 11,
    fontWeight: 800,
    letterSpacing: "0.16em",
  },
  title: {
    margin: "2px 0 14px",
    fontSize: 26,
    lineHeight: 1.1,
    fontWeight: 750,
  },
  label: {
    display: "block",
    marginBottom: 6,
    color: "#cbd4c8",
    fontSize: 12,
    fontWeight: 700,
  },
  inputRow: {
    display: "flex",
    flexWrap: "wrap",
    gap: 8,
  },
  input: {
    flex: "1 1 150px",
    minWidth: 0,
    padding: "10px 12px",
    border: "1px solid rgba(224, 237, 198, 0.24)",
    borderRadius: 9,
    outline: "none",
    background: "rgba(255, 255, 255, 0.07)",
    color: "#fffbed",
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
    fontSize: 14,
  },
  primaryButton: {
    padding: "10px 14px",
    border: 0,
    borderRadius: 9,
    background: "#a7c875",
    color: "#17231f",
    fontSize: 13,
    fontWeight: 800,
  },
  secondaryButton: {
    padding: "10px 14px",
    border: "1px solid rgba(224, 237, 198, 0.24)",
    borderRadius: 9,
    background: "transparent",
    color: "#e9eddf",
    fontSize: 13,
    fontWeight: 700,
  },
  hint: {
    margin: "10px 0 12px",
    color: "#9ea9a1",
    fontSize: 12,
  },
  legend: {
    display: "flex",
    flexWrap: "wrap",
    gap: "7px 12px",
  },
  legendItem: {
    display: "inline-flex",
    alignItems: "center",
    gap: 5,
    color: "#cbd4c8",
    fontSize: 11,
  },
  swatch: {
    width: 8,
    height: 8,
    borderRadius: 2,
  },
  loading: {
    position: "absolute",
    top: "50%",
    left: "50%",
    transform: "translate(-50%, -50%)",
    color: "#cbd4c8",
    fontSize: 14,
  },
};
