import { useState } from "react";
import { joinGame, GameSocket, MessageType } from "@/api/webSocket/GameSocket";

const pieces = [
  {
    icon: "🦸‍♂️",
    id: "11",
    lp: 3,
    shield: 3,
    stamina: 100,
    weapons: ["pistol"],
    tools: [],
  },
  {
    icon: "🦸‍♀️",
    id: "12",
    lp: 3,
    shield: 3,
    stamina: 4,
    weapons: ["pistol"],
    tools: [],
  },
  {
    icon: "🧙‍♂️",
    id: "13",
    lp: 3,
    shield: 3,
    stamina: 4,
    weapons: ["pistol"],
    tools: [],
  },
];

export const Home = () => {
  const [formData, setFormData] = useState({
    conta: "",
    secao: "",
  });

  const handleChange = (e: any) => {
    const { name, value } = e.target;
    setFormData((prevState) => ({
      ...prevState,
      [name]: value,
    }));
  };

  const handleSubmit = (e: any) => {
    e.preventDefault();
    console.log("Dados enviados:", formData);
    joinGame(formData.secao, formData.conta, pieces);
  };

  const sandEcho = () => {
    GameSocket.send(
      JSON.stringify({
        type: MessageType.echo,
        payload: "Testando 123 ...",
      })
    );
  };

  return (
    <div>
      <button
        style={{
          backgroundColor: "grey",
          color: "white",
          padding: "10px 15px",
          border: "none",
          borderRadius: "4px",
          cursor: "pointer",
        }}
        onClick={sandEcho}
      >
        Send Echo
      </button>
      <div style={{ maxWidth: "400px", margin: "0 auto", padding: "20px" }}>
        <h2>Formulário de Acesso</h2>
        <form onSubmit={handleSubmit}>
          <div style={{ marginBottom: "15px" }}>
            <label
              htmlFor="conta"
              style={{ display: "block", marginBottom: "5px" }}
            >
              Conta:
            </label>
            <input
              type="text"
              id="conta"
              name="conta"
              value={formData.conta}
              onChange={handleChange}
              style={{ width: "100%", padding: "8px", boxSizing: "border-box" }}
              placeholder="Digite sua conta"
              required
            />
          </div>

          <div style={{ marginBottom: "15px" }}>
            <label
              htmlFor="secao"
              style={{ display: "block", marginBottom: "5px" }}
            >
              Seção:
            </label>
            <input
              type="text"
              id="secao"
              name="secao"
              value={formData.secao}
              onChange={handleChange}
              style={{ width: "100%", padding: "8px", boxSizing: "border-box" }}
              placeholder="Digite sua seção"
              required
            />
          </div>

          <button
            type="submit"
            style={{
              backgroundColor: "#4CAF50",
              color: "white",
              padding: "10px 15px",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
            }}
          >
            Entrar na Partida
          </button>
        </form>
      </div>
    </div>
  );
};
