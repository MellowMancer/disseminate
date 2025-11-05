const DOT_ROWS = 8;            // Number of rows
const DOT_COLS = 60;            // Number of columns
const DOT_SIZE = 4;             // px, diameter of each dot
const DOT_GAP = 32;             // px, space between dots
const ANIMATION_DURATION = 5;   // seconds

const DotMatrixBackground = () => (
  <div
    aria-hidden="true"
    className="absolute bottom-0 inset-0 w-screen h-screen pointer-events-none -z-10 bg-background"
    style={{ overflow: "hidden" }}
  >
    {Array.from({ length: DOT_ROWS }).map((_, row) =>
      Array.from({ length: DOT_COLS }).map((_, col) => {
        // Randomize each dot's animation delay for organic look
        const delay = (Math.random() * ANIMATION_DURATION).toFixed(2);
        return (
          <span
            key={`${row}-${col}`}
            className="dot-matrix-dot"
            style={{
              left: col * DOT_GAP,
              bottom: row * DOT_GAP,
              animationDelay: `${delay}s`,
              width: DOT_SIZE,
              height: DOT_SIZE,
            }}
          />
        );
      })
    )}
  </div>
);

export default DotMatrixBackground;
