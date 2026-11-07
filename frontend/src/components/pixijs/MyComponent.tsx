import { Application, extend } from "@pixi/react";
import { Container, Sprite, Text, BlurFilter, Texture } from "pixi.js";
import { useMemo } from "react";

extend({ Container, Sprite, Text });

export const MyComponent = () => {
  const blurFilter = useMemo(() => new BlurFilter(4), []);

  return (
    <Application width={800} height={600} background={0x1099bb}>
      <pixiContainer x={0} y={0}>
        <pixiSprite
          texture={Texture.from("https://pixijs.io/pixi-react/img/bunny.png")}
          x={400}
          y={270}
          anchor={{ x: 0.5, y: 0.5 }}
        />
        <pixiContainer x={400} y={330}>
          <pixiText
            text="Hello World"
            anchor={{ x: 0.5, y: 0.5 }}
            filters={[blurFilter]}
            style={{
              fill: "white",
              fontSize: 32,
              fontWeight: "bold",
              fontFamily: "Arial",
            }}
          />
        </pixiContainer>
      </pixiContainer>
    </Application>
  );
};
