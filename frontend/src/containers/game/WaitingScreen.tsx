import "@/styles/game/waitingScreen.css";

const WaitingScreen = ({ count }: { count: number }) => {
  return (
    <div className="waiting-screen">
      <div className="waiting-container">
        <h1>Aguardando Jogadores...</h1>
        <div className="players-list">
          <h2>Jogadores Conectados ({count}/2)</h2>
        </div>
        <div className="loading-animation">
          <div className="dot-flashing"></div>
        </div>
      </div>
    </div>
  );
};

export default WaitingScreen;
