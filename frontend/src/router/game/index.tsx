import { Game } from "@/containers/game";
import { useParams } from "react-router-dom";

export const GridTactics = () => {
  const { sessionID, accoID } = useParams<{
    sessionID: string;
    accoID: string;
  }>();

  return <Game accoID={accoID ?? ""} sessionID={sessionID ?? ""} />;
};
