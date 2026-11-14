import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { GridTactics } from "@/router/game";
import { Home } from "@/containers/home";
import SkillsLab from "@/containers/skills-lab";

export const AppRouter = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/home" replace />} />
        <Route path="/home" element={<Home />} />
        <Route path="/skills-lab" element={<SkillsLab />} />
        <Route
          path="/grid-tactics/:sessionID/:accoID"
          element={<GridTactics />}
        />
        <Route path="*" element={<Navigate to="/home" replace />} />
      </Routes>
    </BrowserRouter>
  );
};
